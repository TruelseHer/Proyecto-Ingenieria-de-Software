package service

import (
	"fmt"
	"strings"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
)

type TripService struct {
	store *repository.Store
	cache *platform.CacheClient
	bus   *platform.EventBroker
}

func NewTripService(store *repository.Store, cache *platform.CacheClient, bus *platform.EventBroker) *TripService {
	svc := &TripService{
		store: store,
		cache: cache,
		bus:   bus,
	}
	if bus != nil {
		bus.Subscribe(platform.TopicReservationDetected, func(ev platform.Event) {
			if msg, ok := ev.Payload.(platform.ReservationDetectedMessage); ok {
				_, _, _ = svc.IngestDetectedReservation(msg)
			}
		})
	}
	return svc
}

func (s *TripService) ListTrips(userID, statusFilter, typeFilter string) ([]*domain.Trip, error) {
	if _, err := s.store.GetUser(userID); err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("user:%s:trips:%s:%s", userID, statusFilter, typeFilter)
	var cached []*domain.Trip
	if s.cache != nil && s.cache.GetJSON(cacheKey, &cached) {
		return cached, nil
	}

	trips := s.store.ListTripsByUser(userID)
	now := time.Now().UTC()
	var filtered []*domain.Trip
	for _, t := range trips {
		if t.EvaluateLifecycle(now) {
			_ = s.store.SaveTrip(t)
			s.emitLifecycle(t, "TRIP_ARCHIVED", false, "")
		}
		if statusFilter != "" && statusFilter != "ALL" && string(t.Status) != strings.ToUpper(statusFilter) {
			continue
		}
		if typeFilter != "" && typeFilter != "ALL" {
			matched := false
			for _, r := range t.Reservations {
				if string(r.Type) == strings.ToUpper(typeFilter) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		filtered = append(filtered, t)
	}

	if s.cache != nil {
		s.cache.SetJSON(cacheKey, filtered, 30*time.Second)
	}
	return filtered, nil
}

func (s *TripService) GetTripForUser(userID, tripID, ipAddress string) (*domain.Trip, error) {
	t, err := s.store.GetTrip(tripID)
	if err != nil {
		return nil, err
	}
	user, uErr := s.store.GetUser(userID)
	canAccess := t.CanAccess(userID)
	if !canAccess && uErr == nil {
		canAccess = t.CanAccessByUser(user)
	}
	if !canAccess {
		_ = s.store.AppendAuditLog(domain.AuditLogEntry{
			ID:           platform.NewID("aud"),
			ActorUserID:  userID,
			Action:       "ACCESS_TRIP",
			ResourceType: "TRIP",
			ResourceID:   tripID,
			Outcome:      "DENIED",
			Reason:       "El usuario no es propietario ni invitado autorizado del viaje",
			IPAddress:    ipAddress,
			CreatedAt:    time.Now().UTC(),
		})
		return nil, domain.ErrForbidden
	}
	return t, nil
}

func (s *TripService) CreateTrip(userID, title, destinationCity, countryCode string) (*domain.Trip, error) {
	if _, err := s.store.GetUser(userID); err != nil {
		return nil, err
	}
	title = strings.TrimSpace(title)
	destinationCity = strings.TrimSpace(destinationCity)
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	if title == "" || destinationCity == "" {
		return nil, fmt.Errorf("%w: el título y la ciudad de destino son obligatorios", domain.ErrValidation)
	}
	if countryCode == "" {
		countryCode = inferCountryCode(destinationCity)
	}
	now := time.Now().UTC()
	trip := &domain.Trip{
		ID:               platform.NewID("trip"),
		OwnerID:          userID,
		SharedWithEmails: []string{},
		Title:            title,
		DestinationCity:  destinationCity,
		CountryCode:      countryCode,
		CoverImageURL:    defaultCoverForCity(destinationCity),
		Status:           domain.TripStatusActive,
		StartDate:        now.Add(7 * 24 * time.Hour),
		EndDate:          now.Add(12 * 24 * time.Hour),
		Reservations:     []*domain.Reservation{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.store.SaveTrip(trip); err != nil {
		return nil, err
	}
	s.invalidateUser(userID)
	s.emitLifecycle(trip, "TRIP_CREATED", false, "")
	return trip, nil
}

func (s *TripService) ArchiveTrip(userID, tripID string) (*domain.Trip, error) {
	t, err := s.GetTripForUser(userID, tripID, "127.0.0.1")
	if err != nil {
		return nil, err
	}
	t.Archive()
	if err := s.store.SaveTrip(t); err != nil {
		return nil, err
	}
	s.invalidateTripMembers(t)
	s.emitLifecycle(t, "TRIP_ARCHIVED", false, "")
	return t, nil
}

func (s *TripService) RestoreTrip(userID, tripID string) (*domain.Trip, error) {
	t, err := s.GetTripForUser(userID, tripID, "127.0.0.1")
	if err != nil {
		return nil, err
	}
	t.Restore()
	if err := s.store.SaveTrip(t); err != nil {
		return nil, err
	}
	s.invalidateTripMembers(t)
	s.emitLifecycle(t, "TRIP_MODIFIED", false, "")
	return t, nil
}

// ShareTripByEmail invites a companion by email address (linking their UserID if already registered).
func (s *TripService) ShareTripByEmail(ownerID, tripID, companionEmail string) (*domain.Trip, error) {
	t, err := s.GetTripForUser(ownerID, tripID, "127.0.0.1")
	if err != nil {
		return nil, err
	}
	if t.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	cleanEmail := strings.ToLower(strings.TrimSpace(companionEmail))
	if cleanEmail == "" || !strings.Contains(cleanEmail, "@") {
		return nil, fmt.Errorf("%w: ingresa un correo electrónico válido para compartir el viaje", domain.ErrValidation)
	}
	owner, err := s.store.GetUser(ownerID)
	if err == nil && strings.EqualFold(owner.Email, cleanEmail) {
		return nil, fmt.Errorf("%w: ya eres el propietario de este viaje", domain.ErrValidation)
	}

	resolvedUserID := ""
	if targetUser, err := s.store.GetUserByEmail(cleanEmail); err == nil {
		resolvedUserID = targetUser.ID
	}

	t.ShareWithEmail(cleanEmail, resolvedUserID)
	if err := s.store.SaveTrip(t); err != nil {
		return nil, err
	}
	_ = s.store.AppendAuditLog(domain.AuditLogEntry{
		ID:           platform.NewID("aud"),
		ActorUserID:  ownerID,
		Action:       "SHARE_TRIP",
		ResourceType: "TRIP",
		ResourceID:   tripID,
		Outcome:      "ALLOWED",
		Reason:       fmt.Sprintf("Itinerario compartido por correo con %s", cleanEmail),
		IPAddress:    "127.0.0.1",
		CreatedAt:    time.Now().UTC(),
	})
	s.invalidateTripMembers(t)
	return t, nil
}

// ShareTrip supports sharing either by email or by user ID.
func (s *TripService) ShareTrip(ownerID, tripID, target string) (*domain.Trip, error) {
	target = strings.TrimSpace(target)
	if strings.Contains(target, "@") {
		return s.ShareTripByEmail(ownerID, tripID, target)
	}
	u, err := s.store.GetUser(target)
	if err != nil {
		return nil, fmt.Errorf("%w: ingresa un correo electrónico válido para invitar", domain.ErrValidation)
	}
	return s.ShareTripByEmail(ownerID, tripID, u.Email)
}

func (s *TripService) AddReservation(userID, tripID string, res *domain.Reservation) (*domain.Trip, *domain.Reservation, error) {
	t, err := s.GetTripForUser(userID, tripID, "127.0.0.1")
	if err != nil {
		return nil, nil, err
	}
	if res == nil || strings.TrimSpace(res.ProviderName) == "" || strings.TrimSpace(res.ConfirmationCode) == "" {
		return nil, nil, fmt.Errorf("%w: proveedor y código de confirmación son obligatorios", domain.ErrValidation)
	}
	if res.StartDate.IsZero() {
		return nil, nil, fmt.Errorf("%w: la fecha de inicio es obligatoria", domain.ErrValidation)
	}
	if res.EndDate.IsZero() || res.EndDate.Before(res.StartDate) {
		res.EndDate = res.StartDate.Add(3 * time.Hour)
	}
	now := time.Now().UTC()
	if res.ID == "" {
		res.ID = platform.NewID("res")
	}
	res.TripID = t.ID
	if res.Status == "" {
		res.Status = domain.StatusConfirmed
	}
	if res.Source == "" {
		res.Source = domain.SourceManual
	}
	if res.Currency == "" {
		res.Currency = "USD"
	}
	res.CreatedAt = now
	res.UpdatedAt = now

	t.Reservations = append(t.Reservations, res)
	t.RecalculateBounds()
	if err := s.store.SaveTrip(t); err != nil {
		return nil, nil, err
	}
	s.invalidateTripMembers(t)
	s.emitLifecycle(t, "RESERVATION_ADDED", res.IsDisrupted(), res.ProviderName)
	return t, res, nil
}

func (s *TripService) UpdateReservation(userID, reservationID string, in domain.ReservationUpdateInput) (*domain.Trip, *domain.Reservation, error) {
	t, res, err := s.store.FindReservation(reservationID)
	if err != nil {
		return nil, nil, err
	}
	user, _ := s.store.GetUser(userID)
	if !t.CanAccessByUser(user) {
		return nil, nil, domain.ErrForbidden
	}

	res.ApplyManualEdit(in)
	t.RecalculateBounds()
	if err := s.store.SaveTrip(t); err != nil {
		return nil, nil, err
	}
	s.invalidateTripMembers(t)
	s.emitLifecycle(t, "TRIP_MODIFIED", res.IsDisrupted(), res.ProviderName)
	return t, res, nil
}

func (s *TripService) DeleteReservation(userID, reservationID string) (*domain.Trip, error) {
	t, res, err := s.store.FindReservation(reservationID)
	if err != nil {
		return nil, err
	}
	user, _ := s.store.GetUser(userID)
	if !t.CanAccessByUser(user) {
		return nil, domain.ErrForbidden
	}
	updatedTrip, err := s.store.DeleteReservation(reservationID)
	if err != nil {
		return nil, err
	}
	s.invalidateTripMembers(updatedTrip)
	s.emitLifecycle(updatedTrip, "TRIP_MODIFIED", false, res.ProviderName)
	return updatedTrip, nil
}

func (s *TripService) RequestAgencySupport(userID, reservationID, issueSummary string) (*domain.SupportTicket, *domain.Trip, error) {
	user, err := s.store.GetUser(userID)
	if err != nil {
		return nil, nil, err
	}
	t, res, err := s.store.FindReservation(reservationID)
	if err != nil {
		return nil, nil, err
	}
	if !t.CanAccessByUser(user) {
		return nil, nil, domain.ErrForbidden
	}
	agency, err := s.store.GetAgency(user.PreferredAgencyID)
	if err != nil {
		agencies := s.store.ListAgencies()
		if len(agencies) == 0 {
			return nil, nil, domain.ErrNotFound
		}
		agency = agencies[0]
	}

	summary := strings.TrimSpace(issueSummary)
	if summary == "" {
		summary = fmt.Sprintf("Asistencia solicitada para %s (%s) — %s", res.ProviderName, res.ConfirmationCode, res.StatusLabel())
	}

	ticket := &domain.SupportTicket{
		ID:            platform.NewID("tkt"),
		ReferenceCode: "TRM-" + strings.ToUpper(platform.NewID("AG")[3:9]),
		ReservationID: res.ID,
		TripID:        t.ID,
		UserID:        userID,
		AgencyID:      agency.ID,
		AgencyName:    agency.Name,
		IssueSummary:  summary,
		Status:        "OPEN",
		CreatedAt:     time.Now().UTC(),
	}
	res.SupportTicket = ticket
	res.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveTrip(t); err != nil {
		return nil, nil, err
	}
	s.invalidateTripMembers(t)
	return ticket, t, nil
}

// IngestDetectedReservation groups a newly booked or imported segment into a matching active trip,
// or creates a new trip automatically when no match exists.
func (s *TripService) IngestDetectedReservation(msg platform.ReservationDetectedMessage) (*domain.Trip, *domain.Reservation, error) {
	trips := s.store.ListTripsByUser(msg.UserID)
	var target *domain.Trip
	for _, t := range trips {
		if t.Status != domain.TripStatusActive {
			continue
		}
		destMatch := strings.Contains(strings.ToLower(msg.Destination), strings.ToLower(t.DestinationCity)) ||
			strings.Contains(strings.ToLower(msg.Origin), strings.ToLower(t.DestinationCity)) ||
			strings.Contains(strings.ToLower(t.Title), strings.ToLower(msg.Destination))
		if destMatch {
			target = t
			break
		}
	}

	if target == nil {
		destCity := extractCityName(msg.Destination)
		if destCity == "" {
			destCity = "Nuevo Destino"
		}
		var err error
		target, err = s.CreateTrip(msg.UserID, fmt.Sprintf("Viaje a %s", destCity), destCity, inferCountryCode(destCity))
		if err != nil {
			return nil, nil, err
		}
	}

	for _, existing := range target.Reservations {
		if strings.EqualFold(existing.ConfirmationCode, msg.ConfirmationCode) {
			existing.ApplyExternalUpdate(msg.Status, msg.StartDate, msg.EndDate, msg.Notes, msg.Details)
			target.RecalculateBounds()
			_ = s.store.SaveTrip(target)
			s.invalidateTripMembers(target)
			return target, existing, nil
		}
	}

	res := &domain.Reservation{
		ID:               platform.NewID("res"),
		TripID:           target.ID,
		Type:             msg.Type,
		Status:           msg.Status,
		ProviderName:     msg.ProviderName,
		ConfirmationCode: msg.ConfirmationCode,
		Origin:           msg.Origin,
		Destination:      msg.Destination,
		StartDate:        msg.StartDate,
		EndDate:          msg.EndDate,
		Price:            msg.Price,
		Currency:         msg.Currency,
		Notes:            msg.Notes,
		Source:           msg.Source,
		Details:          msg.Details,
	}
	return s.AddReservation(msg.UserID, target.ID, res)
}

func (s *TripService) UpdateUserPreferences(userID, agencyID, currency string) (domain.User, error) {
	u, err := s.store.GetUser(userID)
	if err != nil {
		return u, err
	}
	if agencyID != "" {
		if _, err := s.store.GetAgency(agencyID); err != nil {
			return u, fmt.Errorf("%w: agencia seleccionada no válida", domain.ErrValidation)
		}
		u.PreferredAgencyID = agencyID
	}
	if currency != "" {
		u.HomeCurrency = strings.ToUpper(strings.TrimSpace(currency))
	}
	if err := s.store.SaveUser(u); err != nil {
		return u, err
	}
	s.invalidateUser(userID)
	return u, nil
}

func (s *TripService) invalidateUser(userID string) {
	if s.cache != nil {
		s.cache.DeletePrefix("user:" + userID + ":")
	}
}

func (s *TripService) invalidateTripMembers(t *domain.Trip) {
	s.invalidateUser(t.OwnerID)
	for _, memberID := range t.SharedWithUserIDs {
		s.invalidateUser(memberID)
	}
	for _, email := range t.SharedWithEmails {
		if u, err := s.store.GetUserByEmail(email); err == nil {
			s.invalidateUser(u.ID)
		}
	}
}

func (s *TripService) emitLifecycle(t *domain.Trip, eventType string, disrupted bool, providerName string) {
	if s.bus == nil {
		return
	}
	s.bus.Publish(platform.TopicTripLifecycle, t.ID, platform.TripLifecycleMessage{
		EventID:         platform.NewID("evt"),
		TripID:          t.ID,
		UserID:          t.OwnerID,
		EventType:       eventType,
		DestinationCity: t.DestinationCity,
		CountryCode:     t.CountryCode,
		TransportTypes:  t.TransportTypes(),
		Disrupted:       disrupted,
		ProviderName:    providerName,
		Timestamp:       time.Now().UTC(),
	})
}

func extractCityName(raw string) string {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, "("); idx > 0 {
		raw = strings.TrimSpace(raw[:idx])
	}
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "tokio") || strings.Contains(lower, "tokyo") || strings.Contains(lower, "haneda") || strings.Contains(lower, "shinjuku"):
		return "Tokio"
	case strings.Contains(lower, "kioto") || strings.Contains(lower, "kyoto"):
		return "Kioto"
	case strings.Contains(lower, "madrid") || strings.Contains(lower, "barajas") || strings.Contains(lower, "atocha"):
		return "Madrid"
	case strings.Contains(lower, "barcelona") || strings.Contains(lower, "sants"):
		return "Barcelona"
	case strings.Contains(lower, "bariloche") || strings.Contains(lower, "llao"):
		return "Bariloche"
	case strings.Contains(lower, "roma") || strings.Contains(lower, "fiumicino"):
		return "Roma"
	case strings.Contains(lower, "cusco") || strings.Contains(lower, "machu"):
		return "Cusco"
	case strings.Contains(lower, "parís") || strings.Contains(lower, "paris") || strings.Contains(lower, "londres"):
		return "París"
	default:
		return raw
	}
}

func inferCountryCode(city string) string {
	lower := strings.ToLower(strings.TrimSpace(city))
	switch {
	case strings.Contains(lower, "tokio") || strings.Contains(lower, "kioto") || strings.Contains(lower, "osaka"):
		return "JP"
	case strings.Contains(lower, "madrid") || strings.Contains(lower, "barcelona") || strings.Contains(lower, "sevilla"):
		return "ES"
	case strings.Contains(lower, "bariloche") || strings.Contains(lower, "buenos aires") || strings.Contains(lower, "mendoza"):
		return "AR"
	case strings.Contains(lower, "roma") || strings.Contains(lower, "florencia") || strings.Contains(lower, "milán"):
		return "IT"
	case strings.Contains(lower, "cusco") || strings.Contains(lower, "lima"):
		return "PE"
	case strings.Contains(lower, "parís") || strings.Contains(lower, "paris"):
		return "FR"
	case strings.Contains(lower, "londres") || strings.Contains(lower, "london"):
		return "GB"
	default:
		return "INT"
	}
}

func defaultCoverForCity(city string) string {
	lower := strings.ToLower(strings.TrimSpace(city))
	switch {
	case strings.Contains(lower, "tokio") || strings.Contains(lower, "kioto"):
		return "https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?auto=format&fit=crop&w=1000&q=80"
	case strings.Contains(lower, "madrid") || strings.Contains(lower, "barcelona"):
		return "https://images.unsplash.com/photo-1539037116277-4db20889f2d4?auto=format&fit=crop&w=1000&q=80"
	case strings.Contains(lower, "bariloche"):
		return "https://images.unsplash.com/photo-1518709268805-4e9042af9f23?auto=format&fit=crop&w=1000&q=80"
	case strings.Contains(lower, "roma"):
		return "https://images.unsplash.com/photo-1552832230-c0197dd311b5?auto=format&fit=crop&w=1000&q=80"
	case strings.Contains(lower, "cusco"):
		return "https://images.unsplash.com/photo-1526392060635-9d6019884377?auto=format&fit=crop&w=1000&q=80"
	default:
		return "https://images.unsplash.com/photo-1502602898657-3e91760cbb34?auto=format&fit=crop&w=1000&q=80"
	}
}
