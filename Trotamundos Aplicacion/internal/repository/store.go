package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	mu                 sync.RWMutex
	writeDB            *sql.DB
	readDB             *sql.DB
	users              map[string]domain.User
	agencies           map[string]domain.TravelAgency
	packages           map[string]domain.CatalogPackage
	offers             map[string]domain.CatalogOffer
	trips              map[string]*domain.Trip
	emailConfirmations map[string]domain.EmailConfirmation
	senderRules        map[string]domain.SenderRule
	pendingSyncs       map[string]domain.PendingSyncItem
	facts              []domain.AnonymizedTripFact
	auditLogs          []domain.AuditLogEntry
}

func NewMemoryStore() *Store {
	s := &Store{
		users:              make(map[string]domain.User),
		agencies:           make(map[string]domain.TravelAgency),
		packages:           make(map[string]domain.CatalogPackage),
		offers:             make(map[string]domain.CatalogOffer),
		trips:              make(map[string]*domain.Trip),
		emailConfirmations: make(map[string]domain.EmailConfirmation),
		senderRules:        make(map[string]domain.SenderRule),
		pendingSyncs:       make(map[string]domain.PendingSyncItem),
	}
	SeedInitialData(s)
	return s
}

func NewPostgresStore(primaryDSN, readDSN string) (*Store, error) {
	writeDB, err := sql.Open("pgx", primaryDSN)
	if err != nil {
		return nil, fmt.Errorf("open primary db: %w", err)
	}
	writeDB.SetMaxOpenConns(20)
	writeDB.SetMaxIdleConns(5)
	writeDB.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	for i := 0; i < 15; i++ {
		if err = writeDB.PingContext(ctx); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		_ = writeDB.Close()
		return nil, fmt.Errorf("ping primary db: %w", err)
	}

	readDB := writeDB
	if readDSN != "" && readDSN != primaryDSN {
		if rdb, rerr := sql.Open("pgx", readDSN); rerr == nil {
			rdb.SetMaxOpenConns(30)
			rdb.SetMaxIdleConns(10)
			if rdb.PingContext(ctx) == nil {
				readDB = rdb
			} else {
				_ = rdb.Close()
			}
		}
	}

	s := &Store{
		writeDB:            writeDB,
		readDB:             readDB,
		users:              make(map[string]domain.User),
		agencies:           make(map[string]domain.TravelAgency),
		packages:           make(map[string]domain.CatalogPackage),
		offers:             make(map[string]domain.CatalogOffer),
		trips:              make(map[string]*domain.Trip),
		emailConfirmations: make(map[string]domain.EmailConfirmation),
		senderRules:        make(map[string]domain.SenderRule),
		pendingSyncs:       make(map[string]domain.PendingSyncItem),
	}

	if err := s.migrate(ctx); err != nil {
		return nil, err
	}

	var pkgCount int
	if err := s.readDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM catalog_packages").Scan(&pkgCount); err != nil {
		return nil, fmt.Errorf("check catalog seed status: %w", err)
	}
	if pkgCount == 0 {
		SeedInitialData(s)
	}
	return s, nil
}

func (s *Store) Close() {
	if s.readDB != nil && s.readDB != s.writeDB {
		_ = s.readDB.Close()
	}
	if s.writeDB != nil {
		_ = s.writeDB.Close()
	}
}

func (s *Store) Mode() string {
	if s.writeDB != nil {
		return "postgresql"
	}
	return "memory"
}

func (s *Store) Ping() error {
	if s.writeDB == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.writeDB.PingContext(ctx); err != nil {
		return err
	}
	if s.readDB != nil && s.readDB != s.writeDB {
		return s.readDB.PingContext(ctx)
	}
	return nil
}

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS travel_agencies (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			support_email TEXT NOT NULL,
			phone TEXT NOT NULL,
			sla_hours INTEGER NOT NULL,
			specialty TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL,
			home_currency TEXT NOT NULL,
			preferred_agency_id TEXT NOT NULL,
			mailbox_authorized BOOLEAN NOT NULL DEFAULT FALSE,
			connected_email TEXT NOT NULL,
			avatar_initials TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS trips (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			shared_with JSONB NOT NULL DEFAULT '[]'::jsonb,
			shared_with_emails JSONB NOT NULL DEFAULT '[]'::jsonb,
			title TEXT NOT NULL,
			destination_city TEXT NOT NULL,
			country_code TEXT NOT NULL,
			cover_image_url TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			start_date TIMESTAMPTZ NOT NULL,
			end_date TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`ALTER TABLE trips ADD COLUMN IF NOT EXISTS shared_with_emails JSONB NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE trips ADD COLUMN IF NOT EXISTS cover_image_url TEXT NOT NULL DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS idx_trips_owner ON trips(owner_id)`,
		`CREATE TABLE IF NOT EXISTS reservations (
			id TEXT PRIMARY KEY,
			trip_id TEXT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
			type TEXT NOT NULL,
			status TEXT NOT NULL,
			provider_name TEXT NOT NULL,
			confirmation_code TEXT NOT NULL,
			origin TEXT NOT NULL,
			destination TEXT NOT NULL,
			start_date TIMESTAMPTZ NOT NULL,
			end_date TIMESTAMPTZ NOT NULL,
			price DOUBLE PRECISION NOT NULL,
			currency TEXT NOT NULL,
			notes TEXT NOT NULL DEFAULT '',
			source TEXT NOT NULL,
			manually_modified BOOLEAN NOT NULL DEFAULT FALSE,
			modified_fields JSONB NOT NULL DEFAULT '[]'::jsonb,
			details JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_reservations_trip ON reservations(trip_id)`,
		`CREATE TABLE IF NOT EXISTS support_tickets (
			id TEXT PRIMARY KEY,
			reference_code TEXT NOT NULL,
			reservation_id TEXT NOT NULL UNIQUE REFERENCES reservations(id) ON DELETE CASCADE,
			trip_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			agency_id TEXT NOT NULL,
			agency_name TEXT NOT NULL,
			issue_summary TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS catalog_packages (
			id TEXT PRIMARY KEY,
			slug TEXT NOT NULL,
			title TEXT NOT NULL,
			subtitle TEXT NOT NULL,
			destination_city TEXT NOT NULL,
			country_code TEXT NOT NULL,
			region TEXT NOT NULL,
			summary TEXT NOT NULL,
			image_url TEXT NOT NULL,
			badge TEXT NOT NULL,
			duration_days INTEGER NOT NULL,
			price_per_person DOUBLE PRECISION NOT NULL,
			currency TEXT NOT NULL,
			rating DOUBLE PRECISION NOT NULL,
			reviews_count INTEGER NOT NULL,
			includes_summary JSONB NOT NULL DEFAULT '[]'::jsonb,
			available_dates JSONB NOT NULL DEFAULT '[]'::jsonb,
			included_segments JSONB NOT NULL DEFAULT '[]'::jsonb
		)`,
		`CREATE TABLE IF NOT EXISTS catalog_offers (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			provider_name TEXT NOT NULL,
			provider_email TEXT NOT NULL,
			title TEXT NOT NULL,
			subtitle TEXT NOT NULL,
			origin TEXT NOT NULL,
			destination TEXT NOT NULL,
			destination_city TEXT NOT NULL,
			country_code TEXT NOT NULL,
			start_date TIMESTAMPTZ NOT NULL,
			end_date TIMESTAMPTZ NOT NULL,
			price DOUBLE PRECISION NOT NULL,
			currency TEXT NOT NULL,
			seats_available INTEGER NOT NULL,
			rating DOUBLE PRECISION NOT NULL,
			image_url TEXT NOT NULL,
			highlight_tag TEXT NOT NULL,
			confirmation_prefix TEXT NOT NULL,
			sample_pnr TEXT NOT NULL,
			details JSONB NOT NULL DEFAULT '{}'::jsonb
		)`,
		`CREATE INDEX IF NOT EXISTS idx_catalog_offers_type ON catalog_offers(type)`,
		`CREATE INDEX IF NOT EXISTS idx_catalog_offers_pnr ON catalog_offers(sample_pnr)`,
		`CREATE TABLE IF NOT EXISTS email_confirmations (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			from_email TEXT NOT NULL,
			provider_name TEXT NOT NULL,
			subject TEXT NOT NULL,
			body TEXT NOT NULL,
			pnr TEXT NOT NULL,
			type TEXT NOT NULL,
			origin TEXT NOT NULL,
			destination TEXT NOT NULL,
			destination_city TEXT NOT NULL,
			country_code TEXT NOT NULL,
			start_date TIMESTAMPTZ NOT NULL,
			end_date TIMESTAMPTZ NOT NULL,
			price DOUBLE PRECISION NOT NULL,
			currency TEXT NOT NULL,
			details JSONB NOT NULL DEFAULT '{}'::jsonb,
			status TEXT NOT NULL,
			status_reason TEXT NOT NULL DEFAULT '',
			imported_trip_id TEXT NOT NULL DEFAULT '',
			received_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_email_confirmations_user ON email_confirmations(user_id)`,
		`CREATE TABLE IF NOT EXISTS sender_rules (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			pattern TEXT NOT NULL,
			provider_label TEXT NOT NULL,
			category TEXT NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			is_default BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sender_rules_user ON sender_rules(user_id)`,
		`CREATE TABLE IF NOT EXISTS pending_syncs (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			source TEXT NOT NULL,
			provider_name TEXT NOT NULL,
			summary TEXT NOT NULL,
			raw_payload TEXT NOT NULL,
			error_reason TEXT NOT NULL,
			attempts INTEGER NOT NULL,
			status TEXT NOT NULL,
			next_retry_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS analytics_facts (
			id TEXT PRIMARY KEY,
			pseudonym_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			destination_city TEXT NOT NULL,
			country_code TEXT NOT NULL,
			transport_types JSONB NOT NULL DEFAULT '[]'::jsonb,
			disrupted BOOLEAN NOT NULL DEFAULT FALSE,
			provider_name TEXT NOT NULL DEFAULT '',
			recorded_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			actor_user_id TEXT NOT NULL,
			action TEXT NOT NULL,
			resource_type TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			outcome TEXT NOT NULL,
			reason TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)`,
	}
	for _, stmt := range stmts {
		if _, err := s.writeDB.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
	}
	return nil
}

// ---- Travel Agencies ----

func (s *Store) SaveAgency(a domain.TravelAgency) error {
	if s.writeDB != nil {
		_, err := s.writeDB.Exec(`
			INSERT INTO travel_agencies (id, name, support_email, phone, sla_hours, specialty)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (id) DO UPDATE SET
				name=EXCLUDED.name, support_email=EXCLUDED.support_email,
				phone=EXCLUDED.phone, sla_hours=EXCLUDED.sla_hours, specialty=EXCLUDED.specialty
		`, a.ID, a.Name, a.SupportEmail, a.Phone, a.SLAHours, a.Specialty)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agencies[a.ID] = a
	return nil
}

func (s *Store) ListAgencies() []domain.TravelAgency {
	if s.readDB != nil {
		rows, err := s.readDB.Query(`SELECT id, name, support_email, phone, sla_hours, specialty FROM travel_agencies ORDER BY name`)
		if err == nil {
			defer rows.Close()
			var list []domain.TravelAgency
			for rows.Next() {
				var a domain.TravelAgency
				if err := rows.Scan(&a.ID, &a.Name, &a.SupportEmail, &a.Phone, &a.SLAHours, &a.Specialty); err == nil {
					list = append(list, a)
				}
			}
			return list
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]domain.TravelAgency, 0, len(s.agencies))
	for _, a := range s.agencies {
		list = append(list, a)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}

func (s *Store) GetAgency(id string) (domain.TravelAgency, error) {
	if s.readDB != nil {
		var a domain.TravelAgency
		err := s.readDB.QueryRow(`SELECT id, name, support_email, phone, sla_hours, specialty FROM travel_agencies WHERE id=$1`, id).
			Scan(&a.ID, &a.Name, &a.SupportEmail, &a.Phone, &a.SLAHours, &a.Specialty)
		if err == sql.ErrNoRows {
			return domain.TravelAgency{}, domain.ErrNotFound
		}
		return a, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.agencies[id]
	if !ok {
		return domain.TravelAgency{}, domain.ErrNotFound
	}
	return a, nil
}

// ---- Users ----

func (s *Store) SaveUser(u domain.User) error {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	if s.writeDB != nil {
		var existingID string
		err := s.readDB.QueryRow(`SELECT id FROM users WHERE LOWER(email)=LOWER($1)`, u.Email).Scan(&existingID)
		if err == nil && existingID != "" && existingID != u.ID {
			return domain.ErrEmailAlreadyUsed
		}
		_, err = s.writeDB.Exec(`
			INSERT INTO users (id, name, email, password_hash, role, home_currency, preferred_agency_id, mailbox_authorized, connected_email, avatar_initials, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			ON CONFLICT (id) DO UPDATE SET
				name=EXCLUDED.name, email=EXCLUDED.email, password_hash=EXCLUDED.password_hash,
				role=EXCLUDED.role, home_currency=EXCLUDED.home_currency,
				preferred_agency_id=EXCLUDED.preferred_agency_id,
				mailbox_authorized=EXCLUDED.mailbox_authorized,
				connected_email=EXCLUDED.connected_email,
				avatar_initials=EXCLUDED.avatar_initials
		`, u.ID, u.Name, u.Email, u.PasswordHash, u.Role, u.HomeCurrency, u.PreferredAgencyID, u.MailboxAuthorized, u.ConnectedEmail, u.AvatarInitials, u.CreatedAt)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for id, existing := range s.users {
		if id != u.ID && strings.EqualFold(existing.Email, u.Email) {
			return domain.ErrEmailAlreadyUsed
		}
	}
	s.users[u.ID] = u
	return nil
}

func (s *Store) ListUsers() []domain.User {
	if s.readDB != nil {
		rows, err := s.readDB.Query(`
			SELECT id, name, email, password_hash, role, home_currency, preferred_agency_id, mailbox_authorized, connected_email, avatar_initials, created_at
			FROM users ORDER BY created_at
		`)
		if err == nil {
			defer rows.Close()
			var list []domain.User
			for rows.Next() {
				var u domain.User
				if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.HomeCurrency, &u.PreferredAgencyID, &u.MailboxAuthorized, &u.ConnectedEmail, &u.AvatarInitials, &u.CreatedAt); err == nil {
					list = append(list, u)
				}
			}
			return list
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]domain.User, 0, len(s.users))
	for _, u := range s.users {
		list = append(list, u)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.Before(list[j].CreatedAt) })
	return list
}

func (s *Store) GetUser(id string) (domain.User, error) {
	if s.readDB != nil {
		var u domain.User
		err := s.readDB.QueryRow(`
			SELECT id, name, email, password_hash, role, home_currency, preferred_agency_id, mailbox_authorized, connected_email, avatar_initials, created_at
			FROM users WHERE id=$1
		`, id).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.HomeCurrency, &u.PreferredAgencyID, &u.MailboxAuthorized, &u.ConnectedEmail, &u.AvatarInitials, &u.CreatedAt)
		if err == sql.ErrNoRows {
			return domain.User{}, domain.ErrNotFound
		}
		return u, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (s *Store) GetUserByEmail(email string) (domain.User, error) {
	clean := strings.ToLower(strings.TrimSpace(email))
	if clean == "" {
		return domain.User{}, domain.ErrNotFound
	}
	if s.readDB != nil {
		var u domain.User
		err := s.readDB.QueryRow(`
			SELECT id, name, email, password_hash, role, home_currency, preferred_agency_id, mailbox_authorized, connected_email, avatar_initials, created_at
			FROM users WHERE LOWER(email)=LOWER($1)
		`, clean).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.HomeCurrency, &u.PreferredAgencyID, &u.MailboxAuthorized, &u.ConnectedEmail, &u.AvatarInitials, &u.CreatedAt)
		if err == sql.ErrNoRows {
			return domain.User{}, domain.ErrNotFound
		}
		return u, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if strings.EqualFold(strings.TrimSpace(u.Email), clean) {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}

// ---- Preloaded Travel Catalog ----

func (s *Store) SaveCatalogPackage(p domain.CatalogPackage) error {
	if s.writeDB != nil {
		incJSON, _ := json.Marshal(p.IncludesSummary)
		datesJSON, _ := json.Marshal(p.AvailableDates)
		segsJSON, _ := json.Marshal(p.IncludedSegments)
		_, err := s.writeDB.Exec(`
			INSERT INTO catalog_packages (
				id, slug, title, subtitle, destination_city, country_code, region, summary,
				image_url, badge, duration_days, price_per_person, currency, rating, reviews_count,
				includes_summary, available_dates, included_segments
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
			ON CONFLICT (id) DO UPDATE SET
				slug=EXCLUDED.slug, title=EXCLUDED.title, subtitle=EXCLUDED.subtitle,
				destination_city=EXCLUDED.destination_city, country_code=EXCLUDED.country_code,
				region=EXCLUDED.region, summary=EXCLUDED.summary, image_url=EXCLUDED.image_url,
				badge=EXCLUDED.badge, duration_days=EXCLUDED.duration_days,
				price_per_person=EXCLUDED.price_per_person, currency=EXCLUDED.currency,
				rating=EXCLUDED.rating, reviews_count=EXCLUDED.reviews_count,
				includes_summary=EXCLUDED.includes_summary, available_dates=EXCLUDED.available_dates,
				included_segments=EXCLUDED.included_segments
		`, p.ID, p.Slug, p.Title, p.Subtitle, p.DestinationCity, p.CountryCode, p.Region, p.Summary,
			p.ImageURL, p.Badge, p.DurationDays, p.PricePerPerson, p.Currency, p.Rating, p.ReviewsCount,
			incJSON, datesJSON, segsJSON)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.packages[p.ID] = cloneCatalogPackage(p)
	return nil
}

func (s *Store) ListCatalogPackages() []domain.CatalogPackage {
	if s.readDB != nil {
		rows, err := s.readDB.Query(`
			SELECT id, slug, title, subtitle, destination_city, country_code, region, summary,
			       image_url, badge, duration_days, price_per_person, currency, rating, reviews_count,
			       includes_summary, available_dates, included_segments
			FROM catalog_packages ORDER BY rating DESC, id ASC
		`)
		if err == nil {
			defer rows.Close()
			var list []domain.CatalogPackage
			for rows.Next() {
				var p domain.CatalogPackage
				var incRaw, datesRaw, segsRaw []byte
				if err := rows.Scan(
					&p.ID, &p.Slug, &p.Title, &p.Subtitle, &p.DestinationCity, &p.CountryCode, &p.Region, &p.Summary,
					&p.ImageURL, &p.Badge, &p.DurationDays, &p.PricePerPerson, &p.Currency, &p.Rating, &p.ReviewsCount,
					&incRaw, &datesRaw, &segsRaw,
				); err == nil {
					_ = json.Unmarshal(incRaw, &p.IncludesSummary)
					_ = json.Unmarshal(datesRaw, &p.AvailableDates)
					_ = json.Unmarshal(segsRaw, &p.IncludedSegments)
					list = append(list, p)
				}
			}
			return list
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]domain.CatalogPackage, 0, len(s.packages))
	for _, p := range s.packages {
		list = append(list, cloneCatalogPackage(p))
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Rating == list[j].Rating {
			return list[i].ID < list[j].ID
		}
		return list[i].Rating > list[j].Rating
	})
	return list
}

func (s *Store) GetCatalogPackage(id string) (domain.CatalogPackage, error) {
	clean := strings.TrimSpace(id)
	if s.readDB != nil {
		var p domain.CatalogPackage
		var incRaw, datesRaw, segsRaw []byte
		err := s.readDB.QueryRow(`
			SELECT id, slug, title, subtitle, destination_city, country_code, region, summary,
			       image_url, badge, duration_days, price_per_person, currency, rating, reviews_count,
			       includes_summary, available_dates, included_segments
			FROM catalog_packages WHERE id=$1 OR slug=$1
		`, clean).Scan(
			&p.ID, &p.Slug, &p.Title, &p.Subtitle, &p.DestinationCity, &p.CountryCode, &p.Region, &p.Summary,
			&p.ImageURL, &p.Badge, &p.DurationDays, &p.PricePerPerson, &p.Currency, &p.Rating, &p.ReviewsCount,
			&incRaw, &datesRaw, &segsRaw,
		)
		if err == sql.ErrNoRows {
			return domain.CatalogPackage{}, domain.ErrNotFound
		}
		if err != nil {
			return domain.CatalogPackage{}, err
		}
		_ = json.Unmarshal(incRaw, &p.IncludesSummary)
		_ = json.Unmarshal(datesRaw, &p.AvailableDates)
		_ = json.Unmarshal(segsRaw, &p.IncludedSegments)
		return p, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.packages[clean]; ok {
		return cloneCatalogPackage(p), nil
	}
	for _, p := range s.packages {
		if p.Slug == clean {
			return cloneCatalogPackage(p), nil
		}
	}
	return domain.CatalogPackage{}, domain.ErrNotFound
}

func (s *Store) SaveCatalogOffer(o domain.CatalogOffer) error {
	if s.writeDB != nil {
		detailsJSON, _ := json.Marshal(o.Details)
		_, err := s.writeDB.Exec(`
			INSERT INTO catalog_offers (
				id, type, provider_name, provider_email, title, subtitle, origin, destination,
				destination_city, country_code, start_date, end_date, price, currency,
				seats_available, rating, image_url, highlight_tag, confirmation_prefix, sample_pnr, details
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
			ON CONFLICT (id) DO UPDATE SET
				type=EXCLUDED.type, provider_name=EXCLUDED.provider_name, provider_email=EXCLUDED.provider_email,
				title=EXCLUDED.title, subtitle=EXCLUDED.subtitle, origin=EXCLUDED.origin,
				destination=EXCLUDED.destination, destination_city=EXCLUDED.destination_city,
				country_code=EXCLUDED.country_code, start_date=EXCLUDED.start_date, end_date=EXCLUDED.end_date,
				price=EXCLUDED.price, currency=EXCLUDED.currency, seats_available=EXCLUDED.seats_available,
				rating=EXCLUDED.rating, image_url=EXCLUDED.image_url, highlight_tag=EXCLUDED.highlight_tag,
				confirmation_prefix=EXCLUDED.confirmation_prefix, sample_pnr=EXCLUDED.sample_pnr,
				details=EXCLUDED.details
		`, o.ID, string(o.Type), o.ProviderName, o.ProviderEmail, o.Title, o.Subtitle, o.Origin, o.Destination,
			o.DestinationCity, o.CountryCode, o.StartDate, o.EndDate, o.Price, o.Currency,
			o.SeatsAvailable, o.Rating, o.ImageURL, o.HighlightTag, o.ConfirmationPrefix, o.SamplePNR, detailsJSON)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offers[o.ID] = o
	return nil
}

func (s *Store) ListCatalogOffers(filterType, query string) []domain.CatalogOffer {
	var all []domain.CatalogOffer
	if s.readDB != nil {
		rows, err := s.readDB.Query(`
			SELECT id, type, provider_name, provider_email, title, subtitle, origin, destination,
			       destination_city, country_code, start_date, end_date, price, currency,
			       seats_available, rating, image_url, highlight_tag, confirmation_prefix, sample_pnr, details
			FROM catalog_offers ORDER BY start_date ASC, id ASC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var o domain.CatalogOffer
				var typeStr string
				var detailsRaw []byte
				if err := rows.Scan(
					&o.ID, &typeStr, &o.ProviderName, &o.ProviderEmail, &o.Title, &o.Subtitle, &o.Origin, &o.Destination,
					&o.DestinationCity, &o.CountryCode, &o.StartDate, &o.EndDate, &o.Price, &o.Currency,
					&o.SeatsAvailable, &o.Rating, &o.ImageURL, &o.HighlightTag, &o.ConfirmationPrefix, &o.SamplePNR, &detailsRaw,
				); err == nil {
					o.Type = domain.ReservationType(typeStr)
					_ = json.Unmarshal(detailsRaw, &o.Details)
					all = append(all, o)
				}
			}
		}
	} else {
		s.mu.RLock()
		all = make([]domain.CatalogOffer, 0, len(s.offers))
		for _, o := range s.offers {
			all = append(all, o)
		}
		s.mu.RUnlock()
		sort.Slice(all, func(i, j int) bool {
			if all[i].StartDate.Equal(all[j].StartDate) {
				return all[i].ID < all[j].ID
			}
			return all[i].StartDate.Before(all[j].StartDate)
		})
	}

	ft := strings.ToUpper(strings.TrimSpace(filterType))
	q := strings.ToLower(strings.TrimSpace(query))

	var filtered []domain.CatalogOffer
	for _, o := range all {
		if ft != "" && ft != "ALL" && string(o.Type) != ft {
			continue
		}
		if q != "" {
			haystack := strings.ToLower(strings.Join([]string{
				o.Title,
				o.Subtitle,
				o.ProviderName,
				o.Origin,
				o.Destination,
				o.DestinationCity,
				o.CountryCode,
				o.SamplePNR,
				o.HighlightTag,
			}, " "))
			if !strings.Contains(haystack, q) {
				continue
			}
		}
		filtered = append(filtered, o)
	}
	return filtered
}

func (s *Store) GetCatalogOffer(id string) (domain.CatalogOffer, error) {
	clean := strings.TrimSpace(id)
	if s.readDB != nil {
		var o domain.CatalogOffer
		var typeStr string
		var detailsRaw []byte
		err := s.readDB.QueryRow(`
			SELECT id, type, provider_name, provider_email, title, subtitle, origin, destination,
			       destination_city, country_code, start_date, end_date, price, currency,
			       seats_available, rating, image_url, highlight_tag, confirmation_prefix, sample_pnr, details
			FROM catalog_offers WHERE id=$1
		`, clean).Scan(
			&o.ID, &typeStr, &o.ProviderName, &o.ProviderEmail, &o.Title, &o.Subtitle, &o.Origin, &o.Destination,
			&o.DestinationCity, &o.CountryCode, &o.StartDate, &o.EndDate, &o.Price, &o.Currency,
			&o.SeatsAvailable, &o.Rating, &o.ImageURL, &o.HighlightTag, &o.ConfirmationPrefix, &o.SamplePNR, &detailsRaw,
		)
		if err == sql.ErrNoRows {
			return domain.CatalogOffer{}, domain.ErrNotFound
		}
		if err != nil {
			return domain.CatalogOffer{}, err
		}
		o.Type = domain.ReservationType(typeStr)
		_ = json.Unmarshal(detailsRaw, &o.Details)
		return o, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.offers[clean]
	if !ok {
		return domain.CatalogOffer{}, domain.ErrNotFound
	}
	return o, nil
}

func (s *Store) FindCatalogOfferByPNR(pnr string) (domain.CatalogOffer, error) {
	clean := strings.ToUpper(strings.TrimSpace(pnr))
	if clean == "" {
		return domain.CatalogOffer{}, domain.ErrNotFound
	}
	offers := s.ListCatalogOffers("", "")
	for _, o := range offers {
		if strings.EqualFold(o.SamplePNR, clean) {
			return o, nil
		}
	}
	for _, o := range offers {
		prefix := strings.ToUpper(strings.TrimSpace(o.ConfirmationPrefix))
		if prefix != "" && strings.HasPrefix(clean, prefix+"-") {
			return o, nil
		}
	}
	return domain.CatalogOffer{}, domain.ErrNotFound
}

// ---- Trips & Reservations ----

func (s *Store) SaveTrip(t *domain.Trip) error {
	if s.writeDB != nil {
		tx, err := s.writeDB.Begin()
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			}
		}()

		sharedJSON, _ := json.Marshal(t.SharedWithUserIDs)
		sharedEmailsJSON, _ := json.Marshal(t.SharedWithEmails)
		_, err = tx.Exec(`
			INSERT INTO trips (
				id, owner_id, shared_with, shared_with_emails, title, destination_city,
				country_code, cover_image_url, status, start_date, end_date, created_at, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			ON CONFLICT (id) DO UPDATE SET
				shared_with=EXCLUDED.shared_with,
				shared_with_emails=EXCLUDED.shared_with_emails,
				title=EXCLUDED.title,
				destination_city=EXCLUDED.destination_city,
				country_code=EXCLUDED.country_code,
				cover_image_url=EXCLUDED.cover_image_url,
				status=EXCLUDED.status,
				start_date=EXCLUDED.start_date,
				end_date=EXCLUDED.end_date,
				updated_at=EXCLUDED.updated_at
		`, t.ID, t.OwnerID, sharedJSON, sharedEmailsJSON, t.Title, t.DestinationCity,
			t.CountryCode, t.CoverImageURL, string(t.Status), t.StartDate, t.EndDate, t.CreatedAt, t.UpdatedAt)
		if err != nil {
			return err
		}

		for _, r := range t.Reservations {
			modFieldsJSON, _ := json.Marshal(r.ModifiedFields)
			detailsJSON, _ := json.Marshal(r.Details)
			_, err = tx.Exec(`
				INSERT INTO reservations (
					id, trip_id, type, status, provider_name, confirmation_code, origin, destination,
					start_date, end_date, price, currency, notes, source, manually_modified,
					modified_fields, details, created_at, updated_at
				) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
				ON CONFLICT (id) DO UPDATE SET
					trip_id=EXCLUDED.trip_id, type=EXCLUDED.type, status=EXCLUDED.status,
					provider_name=EXCLUDED.provider_name, confirmation_code=EXCLUDED.confirmation_code,
					origin=EXCLUDED.origin, destination=EXCLUDED.destination,
					start_date=EXCLUDED.start_date, end_date=EXCLUDED.end_date,
					price=EXCLUDED.price, currency=EXCLUDED.currency, notes=EXCLUDED.notes,
					source=EXCLUDED.source, manually_modified=EXCLUDED.manually_modified,
					modified_fields=EXCLUDED.modified_fields, details=EXCLUDED.details,
					updated_at=EXCLUDED.updated_at
			`, r.ID, t.ID, string(r.Type), string(r.Status), r.ProviderName, r.ConfirmationCode,
				r.Origin, r.Destination, r.StartDate, r.EndDate, r.Price, r.Currency, r.Notes,
				string(r.Source), r.ManuallyModified, modFieldsJSON, detailsJSON, r.CreatedAt, r.UpdatedAt)
			if err != nil {
				return err
			}
			if r.SupportTicket != nil {
				tk := r.SupportTicket
				_, err = tx.Exec(`
					INSERT INTO support_tickets (
						id, reference_code, reservation_id, trip_id, user_id, agency_id, agency_name, issue_summary, status, created_at
					) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
					ON CONFLICT (reservation_id) DO UPDATE SET
						status=EXCLUDED.status, issue_summary=EXCLUDED.issue_summary
				`, tk.ID, tk.ReferenceCode, tk.ReservationID, tk.TripID, tk.UserID, tk.AgencyID, tk.AgencyName, tk.IssueSummary, tk.Status, tk.CreatedAt)
				if err != nil {
					return err
				}
			}
		}

		err = tx.Commit()
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.trips[t.ID] = cloneTrip(t)
	return nil
}

func (s *Store) GetTrip(tripID string) (*domain.Trip, error) {
	if s.readDB != nil {
		var t domain.Trip
		var sharedRaw, sharedEmailsRaw []byte
		var statusStr string
		err := s.readDB.QueryRow(`
			SELECT id, owner_id, shared_with, shared_with_emails, title, destination_city,
			       country_code, cover_image_url, status, start_date, end_date, created_at, updated_at
			FROM trips WHERE id=$1
		`, tripID).Scan(
			&t.ID, &t.OwnerID, &sharedRaw, &sharedEmailsRaw, &t.Title, &t.DestinationCity,
			&t.CountryCode, &t.CoverImageURL, &statusStr, &t.StartDate, &t.EndDate, &t.CreatedAt, &t.UpdatedAt,
		)
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		t.Status = domain.TripStatus(statusStr)
		_ = json.Unmarshal(sharedRaw, &t.SharedWithUserIDs)
		_ = json.Unmarshal(sharedEmailsRaw, &t.SharedWithEmails)
		res, err := s.loadReservationsForTrip(t.ID)
		if err != nil {
			return nil, err
		}
		t.Reservations = res
		return &t, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.trips[tripID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cloneTrip(t), nil
}

func (s *Store) ListTripsByUser(userID string) []*domain.Trip {
	user, userErr := s.GetUser(userID)
	all := s.ListAllTrips()
	var result []*domain.Trip
	for _, t := range all {
		if t.CanAccess(userID) || (userErr == nil && t.CanAccessByUser(user)) {
			result = append(result, t)
		}
	}
	return result
}

func (s *Store) ListAllTrips() []*domain.Trip {
	if s.readDB != nil {
		rows, err := s.readDB.Query(`
			SELECT id, owner_id, shared_with, shared_with_emails, title, destination_city,
			       country_code, cover_image_url, status, start_date, end_date, created_at, updated_at
			FROM trips ORDER BY start_date ASC
		`)
		if err == nil {
			defer rows.Close()
			var trips []*domain.Trip
			for rows.Next() {
				var t domain.Trip
				var sharedRaw, sharedEmailsRaw []byte
				var statusStr string
				if err := rows.Scan(
					&t.ID, &t.OwnerID, &sharedRaw, &sharedEmailsRaw, &t.Title, &t.DestinationCity,
					&t.CountryCode, &t.CoverImageURL, &statusStr, &t.StartDate, &t.EndDate, &t.CreatedAt, &t.UpdatedAt,
				); err == nil {
					t.Status = domain.TripStatus(statusStr)
					_ = json.Unmarshal(sharedRaw, &t.SharedWithUserIDs)
					_ = json.Unmarshal(sharedEmailsRaw, &t.SharedWithEmails)
					t.Reservations, _ = s.loadReservationsForTrip(t.ID)
					trips = append(trips, &t)
				}
			}
			return trips
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*domain.Trip, 0, len(s.trips))
	for _, t := range s.trips {
		list = append(list, cloneTrip(t))
	}
	sort.Slice(list, func(i, j int) bool { return list[i].StartDate.Before(list[j].StartDate) })
	return list
}

func (s *Store) loadReservationsForTrip(tripID string) ([]*domain.Reservation, error) {
	rows, err := s.readDB.Query(`
		SELECT r.id, r.trip_id, r.type, r.status, r.provider_name, r.confirmation_code, r.origin, r.destination,
		       r.start_date, r.end_date, r.price, r.currency, r.notes, r.source, r.manually_modified,
		       r.modified_fields, r.details, r.created_at, r.updated_at,
		       tk.id, tk.reference_code, tk.user_id, tk.agency_id, tk.agency_name, tk.issue_summary, tk.status, tk.created_at
		FROM reservations r
		LEFT JOIN support_tickets tk ON tk.reservation_id = r.id
		WHERE r.trip_id = $1
		ORDER BY r.start_date ASC
	`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*domain.Reservation
	for rows.Next() {
		var r domain.Reservation
		var typeStr, statusStr, sourceStr string
		var modRaw, detailsRaw []byte
		var tkID, tkRef, tkUser, tkAgencyID, tkAgencyName, tkIssue, tkStatus sql.NullString
		var tkCreated sql.NullTime
		if err := rows.Scan(
			&r.ID, &r.TripID, &typeStr, &statusStr, &r.ProviderName, &r.ConfirmationCode, &r.Origin, &r.Destination,
			&r.StartDate, &r.EndDate, &r.Price, &r.Currency, &r.Notes, &sourceStr, &r.ManuallyModified,
			&modRaw, &detailsRaw, &r.CreatedAt, &r.UpdatedAt,
			&tkID, &tkRef, &tkUser, &tkAgencyID, &tkAgencyName, &tkIssue, &tkStatus, &tkCreated,
		); err != nil {
			continue
		}
		r.Type = domain.ReservationType(typeStr)
		r.Status = domain.ReservationStatus(statusStr)
		r.Source = domain.ReservationSource(sourceStr)
		_ = json.Unmarshal(modRaw, &r.ModifiedFields)
		_ = json.Unmarshal(detailsRaw, &r.Details)
		if tkID.Valid {
			r.SupportTicket = &domain.SupportTicket{
				ID:            tkID.String,
				ReferenceCode: tkRef.String,
				ReservationID: r.ID,
				TripID:        r.TripID,
				UserID:        tkUser.String,
				AgencyID:      tkAgencyID.String,
				AgencyName:    tkAgencyName.String,
				IssueSummary:  tkIssue.String,
				Status:        tkStatus.String,
				CreatedAt:     tkCreated.Time,
			}
		}
		list = append(list, &r)
	}
	return list, nil
}

func (s *Store) FindReservation(reservationID string) (*domain.Trip, *domain.Reservation, error) {
	if s.readDB != nil {
		var tripID string
		err := s.readDB.QueryRow(`SELECT trip_id FROM reservations WHERE id=$1`, reservationID).Scan(&tripID)
		if err == sql.ErrNoRows {
			return nil, nil, domain.ErrNotFound
		}
		if err != nil {
			return nil, nil, err
		}
		t, err := s.GetTrip(tripID)
		if err != nil {
			return nil, nil, err
		}
		for _, r := range t.Reservations {
			if r.ID == reservationID {
				return t, r, nil
			}
		}
		return nil, nil, domain.ErrNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.trips {
		for _, r := range t.Reservations {
			if r.ID == reservationID {
				ct := cloneTrip(t)
				for _, cr := range ct.Reservations {
					if cr.ID == reservationID {
						return ct, cr, nil
					}
				}
			}
		}
	}
	return nil, nil, domain.ErrNotFound
}

func (s *Store) DeleteReservation(reservationID string) (*domain.Trip, error) {
	t, _, err := s.FindReservation(reservationID)
	if err != nil {
		return nil, err
	}
	filtered := make([]*domain.Reservation, 0, len(t.Reservations))
	for _, r := range t.Reservations {
		if r.ID != reservationID {
			filtered = append(filtered, r)
		}
	}
	t.Reservations = filtered
	t.RecalculateBounds()

	if s.writeDB != nil {
		if _, err := s.writeDB.Exec(`DELETE FROM reservations WHERE id=$1`, reservationID); err != nil {
			return nil, err
		}
		if err := s.SaveTrip(t); err != nil {
			return nil, err
		}
		return t, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.trips[t.ID] = cloneTrip(t)
	return t, nil
}

// ---- Email Confirmations Inbox ----

func (s *Store) SaveEmailConfirmation(c domain.EmailConfirmation) error {
	if s.writeDB != nil {
		detailsJSON, _ := json.Marshal(c.Details)
		_, err := s.writeDB.Exec(`
			INSERT INTO email_confirmations (
				id, user_id, from_email, provider_name, subject, body, pnr, type,
				origin, destination, destination_city, country_code, start_date, end_date,
				price, currency, details, status, status_reason, imported_trip_id, received_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
			ON CONFLICT (id) DO UPDATE SET
				status=EXCLUDED.status,
				status_reason=EXCLUDED.status_reason,
				imported_trip_id=EXCLUDED.imported_trip_id,
				details=EXCLUDED.details
		`, c.ID, c.UserID, c.FromEmail, c.ProviderName, c.Subject, c.Body, c.PNR, string(c.Type),
			c.Origin, c.Destination, c.DestinationCity, c.CountryCode, c.StartDate, c.EndDate,
			c.Price, c.Currency, detailsJSON, c.Status, c.StatusReason, c.ImportedTripID, c.ReceivedAt)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emailConfirmations[c.ID] = c
	return nil
}

func (s *Store) ListEmailConfirmations(userID string) []domain.EmailConfirmation {
	if s.readDB != nil {
		rows, err := s.readDB.Query(`
			SELECT id, user_id, from_email, provider_name, subject, body, pnr, type,
			       origin, destination, destination_city, country_code, start_date, end_date,
			       price, currency, details, status, status_reason, imported_trip_id, received_at
			FROM email_confirmations WHERE user_id=$1 ORDER BY received_at DESC
		`, userID)
		if err == nil {
			defer rows.Close()
			var list []domain.EmailConfirmation
			for rows.Next() {
				var c domain.EmailConfirmation
				var typeStr string
				var detailsRaw []byte
				if err := rows.Scan(
					&c.ID, &c.UserID, &c.FromEmail, &c.ProviderName, &c.Subject, &c.Body, &c.PNR, &typeStr,
					&c.Origin, &c.Destination, &c.DestinationCity, &c.CountryCode, &c.StartDate, &c.EndDate,
					&c.Price, &c.Currency, &detailsRaw, &c.Status, &c.StatusReason, &c.ImportedTripID, &c.ReceivedAt,
				); err == nil {
					c.Type = domain.ReservationType(typeStr)
					_ = json.Unmarshal(detailsRaw, &c.Details)
					list = append(list, c)
				}
			}
			return list
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []domain.EmailConfirmation
	for _, c := range s.emailConfirmations {
		if c.UserID == userID {
			list = append(list, c)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ReceivedAt.After(list[j].ReceivedAt) })
	return list
}

func (s *Store) GetEmailConfirmation(id, userID string) (domain.EmailConfirmation, error) {
	if s.readDB != nil {
		var c domain.EmailConfirmation
		var typeStr string
		var detailsRaw []byte
		err := s.readDB.QueryRow(`
			SELECT id, user_id, from_email, provider_name, subject, body, pnr, type,
			       origin, destination, destination_city, country_code, start_date, end_date,
			       price, currency, details, status, status_reason, imported_trip_id, received_at
			FROM email_confirmations WHERE id=$1 AND user_id=$2
		`, id, userID).Scan(
			&c.ID, &c.UserID, &c.FromEmail, &c.ProviderName, &c.Subject, &c.Body, &c.PNR, &typeStr,
			&c.Origin, &c.Destination, &c.DestinationCity, &c.CountryCode, &c.StartDate, &c.EndDate,
			&c.Price, &c.Currency, &detailsRaw, &c.Status, &c.StatusReason, &c.ImportedTripID, &c.ReceivedAt,
		)
		if err == sql.ErrNoRows {
			return domain.EmailConfirmation{}, domain.ErrNotFound
		}
		if err != nil {
			return domain.EmailConfirmation{}, err
		}
		c.Type = domain.ReservationType(typeStr)
		_ = json.Unmarshal(detailsRaw, &c.Details)
		return c, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.emailConfirmations[id]
	if !ok || c.UserID != userID {
		return domain.EmailConfirmation{}, domain.ErrNotFound
	}
	return c, nil
}

// ---- Sender Rules ----

func (s *Store) ListSenderRules(userID string) []domain.SenderRule {
	if s.readDB != nil {
		rows, err := s.readDB.Query(`
			SELECT id, user_id, pattern, provider_label, category, enabled, is_default, created_at
			FROM sender_rules WHERE user_id=$1 ORDER BY created_at ASC
		`, userID)
		if err == nil {
			defer rows.Close()
			var list []domain.SenderRule
			for rows.Next() {
				var r domain.SenderRule
				if err := rows.Scan(&r.ID, &r.UserID, &r.Pattern, &r.ProviderLabel, &r.Category, &r.Enabled, &r.IsDefault, &r.CreatedAt); err == nil {
					list = append(list, r)
				}
			}
			return list
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []domain.SenderRule
	for _, r := range s.senderRules {
		if r.UserID == userID {
			list = append(list, r)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.Before(list[j].CreatedAt) })
	return list
}

func (s *Store) SaveSenderRule(rule domain.SenderRule) error {
	if s.writeDB != nil {
		_, err := s.writeDB.Exec(`
			INSERT INTO sender_rules (id, user_id, pattern, provider_label, category, enabled, is_default, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (id) DO UPDATE SET
				pattern=EXCLUDED.pattern, provider_label=EXCLUDED.provider_label,
				category=EXCLUDED.category, enabled=EXCLUDED.enabled
		`, rule.ID, rule.UserID, rule.Pattern, rule.ProviderLabel, rule.Category, rule.Enabled, rule.IsDefault, rule.CreatedAt)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.senderRules[rule.ID] = rule
	return nil
}

func (s *Store) DeleteSenderRule(id, userID string) error {
	if s.writeDB != nil {
		res, err := s.writeDB.Exec(`DELETE FROM sender_rules WHERE id=$1 AND user_id=$2`, id, userID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrNotFound
		}
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rule, ok := s.senderRules[id]
	if !ok || rule.UserID != userID {
		return domain.ErrNotFound
	}
	delete(s.senderRules, id)
	return nil
}

func (s *Store) EnsureDefaultSenderRules(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return nil
	}
	existing := s.ListSenderRules(userID)
	if len(existing) > 0 {
		return nil
	}
	now := time.Now().UTC()
	defaults := []struct {
		pattern  string
		label    string
		category string
	}{
		{"@trotamundos.app", "Catálogo Oficial Trotamundos", "Plataforma"},
		{"@booking.com", "Booking.com", "Alojamientos"},
		{"@ana.co.jp", "All Nippon Airways", "Aerolíneas"},
		{"@iberia.com", "Iberia Líneas Aéreas", "Aerolíneas"},
		{"@aerolineas.com.ar", "Aerolíneas Argentinas", "Aerolíneas"},
		{"@latam.com", "LATAM Airlines", "Aerolíneas"},
		{"@ita-airways.com", "ITA Airways", "Aerolíneas"},
		{"@airfrance.fr", "Air France", "Aerolíneas"},
		{"@jrpass.com", "JR Central Shinkansen", "Trenes"},
		{"@renfe.com", "Renfe AVE", "Trenes"},
		{"@eurostar.com", "Eurostar International", "Trenes"},
		{"@perurail.com", "PeruRail Vistadome", "Trenes"},
		{"@hertz.com", "Hertz Rent a Car", "Vehículos"},
		{"@avis.com", "Avis Prestige", "Vehículos"},
		{"@alsa.es", "ALSA Autobuses", "Autobuses"},
		{"@viabariloche.com.ar", "Vía Bariloche", "Autobuses"},
	}
	for idx, ds := range defaults {
		rule := domain.SenderRule{
			ID:            platform.NewID("sr"),
			UserID:        userID,
			Pattern:       ds.pattern,
			ProviderLabel: ds.label,
			Category:      ds.category,
			Enabled:       true,
			IsDefault:     true,
			CreatedAt:     now.Add(time.Duration(idx) * time.Second),
		}
		if err := s.SaveSenderRule(rule); err != nil {
			return err
		}
	}
	return nil
}

// ---- Pending Sync Queue ----

func (s *Store) SavePendingSync(item domain.PendingSyncItem) error {
	if s.writeDB != nil {
		_, err := s.writeDB.Exec(`
			INSERT INTO pending_syncs (id, user_id, source, provider_name, summary, raw_payload, error_reason, attempts, status, next_retry_at, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			ON CONFLICT (id) DO UPDATE SET
				attempts=EXCLUDED.attempts, status=EXCLUDED.status,
				error_reason=EXCLUDED.error_reason, next_retry_at=EXCLUDED.next_retry_at
		`, item.ID, item.UserID, item.Source, item.ProviderName, item.Summary, item.RawPayload, item.ErrorReason, item.Attempts, item.Status, item.NextRetryAt, item.CreatedAt)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingSyncs[item.ID] = item
	return nil
}

func (s *Store) ListPendingSyncs(userID string) []domain.PendingSyncItem {
	if s.readDB != nil {
		rows, err := s.readDB.Query(`
			SELECT id, user_id, source, provider_name, summary, raw_payload, error_reason, attempts, status, next_retry_at, created_at
			FROM pending_syncs WHERE user_id=$1 ORDER BY created_at DESC
		`, userID)
		if err == nil {
			defer rows.Close()
			var list []domain.PendingSyncItem
			for rows.Next() {
				var item domain.PendingSyncItem
				if err := rows.Scan(&item.ID, &item.UserID, &item.Source, &item.ProviderName, &item.Summary, &item.RawPayload, &item.ErrorReason, &item.Attempts, &item.Status, &item.NextRetryAt, &item.CreatedAt); err == nil {
					list = append(list, item)
				}
			}
			return list
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []domain.PendingSyncItem
	for _, item := range s.pendingSyncs {
		if item.UserID == userID {
			list = append(list, item)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
	return list
}

// ---- Analytics Facts & Audit Logs ----

func (s *Store) AppendAnalyticsFact(fact domain.AnonymizedTripFact) error {
	if s.writeDB != nil {
		ttJSON, _ := json.Marshal(fact.TransportTypes)
		_, err := s.writeDB.Exec(`
			INSERT INTO analytics_facts (id, pseudonym_id, event_type, destination_city, country_code, transport_types, disrupted, provider_name, recorded_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT (id) DO NOTHING
		`, fact.ID, fact.PseudonymID, fact.EventType, fact.DestinationCity, fact.CountryCode, ttJSON, fact.Disrupted, fact.ProviderName, fact.RecordedAt)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.facts = append([]domain.AnonymizedTripFact{fact}, s.facts...)
	return nil
}

func (s *Store) ListAnalyticsFacts() []domain.AnonymizedTripFact {
	if s.readDB != nil {
		rows, err := s.readDB.Query(`
			SELECT id, pseudonym_id, event_type, destination_city, country_code, transport_types, disrupted, provider_name, recorded_at
			FROM analytics_facts ORDER BY recorded_at DESC LIMIT 250
		`)
		if err == nil {
			defer rows.Close()
			var list []domain.AnonymizedTripFact
			for rows.Next() {
				var f domain.AnonymizedTripFact
				var ttRaw []byte
				if err := rows.Scan(&f.ID, &f.PseudonymID, &f.EventType, &f.DestinationCity, &f.CountryCode, &ttRaw, &f.Disrupted, &f.ProviderName, &f.RecordedAt); err == nil {
					_ = json.Unmarshal(ttRaw, &f.TransportTypes)
					list = append(list, f)
				}
			}
			return list
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AnonymizedTripFact, len(s.facts))
	copy(out, s.facts)
	return out
}

func (s *Store) AppendAuditLog(entry domain.AuditLogEntry) error {
	if s.writeDB != nil {
		_, err := s.writeDB.Exec(`
			INSERT INTO audit_logs (id, actor_user_id, action, resource_type, resource_id, outcome, reason, ip_address, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT (id) DO NOTHING
		`, entry.ID, entry.ActorUserID, entry.Action, entry.ResourceType, entry.ResourceID, entry.Outcome, entry.Reason, entry.IPAddress, entry.CreatedAt)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auditLogs = append([]domain.AuditLogEntry{entry}, s.auditLogs...)
	return nil
}

func (s *Store) ListAuditLogs(limit int) []domain.AuditLogEntry {
	if limit <= 0 {
		limit = 50
	}
	if s.readDB != nil {
		rows, err := s.readDB.Query(`
			SELECT id, actor_user_id, action, resource_type, resource_id, outcome, reason, ip_address, created_at
			FROM audit_logs ORDER BY created_at DESC LIMIT $1
		`, limit)
		if err == nil {
			defer rows.Close()
			var list []domain.AuditLogEntry
			for rows.Next() {
				var e domain.AuditLogEntry
				if err := rows.Scan(&e.ID, &e.ActorUserID, &e.Action, &e.ResourceType, &e.ResourceID, &e.Outcome, &e.Reason, &e.IPAddress, &e.CreatedAt); err == nil {
					list = append(list, e)
				}
			}
			return list
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := len(s.auditLogs)
	if n > limit {
		n = limit
	}
	out := make([]domain.AuditLogEntry, n)
	copy(out, s.auditLogs[:n])
	return out
}

func cloneTrip(t *domain.Trip) *domain.Trip {
	cp := *t
	cp.SharedWithUserIDs = append([]string(nil), t.SharedWithUserIDs...)
	cp.SharedWithEmails = append([]string(nil), t.SharedWithEmails...)
	cp.Reservations = make([]*domain.Reservation, len(t.Reservations))
	for i, r := range t.Reservations {
		rcp := *r
		rcp.ModifiedFields = append([]string(nil), r.ModifiedFields...)
		if r.SupportTicket != nil {
			tk := *r.SupportTicket
			rcp.SupportTicket = &tk
		}
		cp.Reservations[i] = &rcp
	}
	return &cp
}

func cloneCatalogPackage(p domain.CatalogPackage) domain.CatalogPackage {
	cp := p
	cp.IncludesSummary = append([]string(nil), p.IncludesSummary...)
	cp.AvailableDates = append([]domain.CatalogPackageDate(nil), p.AvailableDates...)
	cp.IncludedSegments = append([]domain.CatalogSegmentTemplate(nil), p.IncludedSegments...)
	return cp
}
