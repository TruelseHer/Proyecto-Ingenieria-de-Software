package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
)

type EmailScanResult struct {
	Accepted         bool
	Message          string
	MatchedRule      string
	ConfirmationCode string
	ReservationType  domain.ReservationType
	ProviderName     string
	TripID           string
}

type InboxSyncBatchResult struct {
	ImportedCount int
	BlockedCount  int
	Message       string
}

type SyncService struct {
	store *repository.Store
	bus   *platform.EventBroker
	guard *platform.ProviderCircuitBreaker
	trips *TripService
}

func NewSyncService(store *repository.Store, bus *platform.EventBroker, guard *platform.ProviderCircuitBreaker, trips *TripService) *SyncService {
	return &SyncService{
		store: store,
		bus:   bus,
		guard: guard,
		trips: trips,
	}
}

func (s *SyncService) ToggleMailboxAuth(userID string, authorized bool) (domain.User, error) {
	u, err := s.store.GetUser(userID)
	if err != nil {
		return u, err
	}
	u.MailboxAuthorized = authorized
	if err := s.store.SaveUser(u); err != nil {
		return u, err
	}
	return u, nil
}

func (s *SyncService) ListSenderRules(userID string) []domain.SenderRule {
	s.store.EnsureDefaultSenderRules(userID)
	return s.store.ListSenderRules(userID)
}

func (s *SyncService) AddSenderRule(userID, pattern, providerLabel, category string) (domain.SenderRule, error) {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	providerLabel = strings.TrimSpace(providerLabel)
	category = strings.TrimSpace(category)
	if pattern == "" || (!strings.Contains(pattern, "@") && !strings.Contains(pattern, ".")) {
		return domain.SenderRule{}, fmt.Errorf("%w: ingresa un dominio o dirección válida (ej. @aerolinea.com)", domain.ErrValidation)
	}
	if providerLabel == "" {
		providerLabel = strings.TrimPrefix(pattern, "@")
	}
	if category == "" {
		category = "Proveedor autorizado"
	}
	rule := domain.SenderRule{
		ID:            platform.NewID("sr"),
		UserID:        userID,
		Pattern:       pattern,
		ProviderLabel: providerLabel,
		Category:      category,
		Enabled:       true,
		IsDefault:     false,
		CreatedAt:     time.Now().UTC(),
	}
	if err := s.store.SaveSenderRule(rule); err != nil {
		return domain.SenderRule{}, err
	}
	return rule, nil
}

func (s *SyncService) ToggleSenderRule(userID, ruleID string) (domain.SenderRule, error) {
	rules := s.ListSenderRules(userID)
	for _, r := range rules {
		if r.ID == ruleID {
			r.Enabled = !r.Enabled
			if err := s.store.SaveSenderRule(r); err != nil {
				return r, err
			}
			return r, nil
		}
	}
	return domain.SenderRule{}, domain.ErrNotFound
}

func (s *SyncService) DeleteSenderRule(userID, ruleID string) error {
	return s.store.DeleteSenderRule(ruleID, userID)
}

func (s *SyncService) ListConfirmations(userID string) []domain.EmailConfirmation {
	return s.store.ListEmailConfirmations(userID)
}

func (s *SyncService) matchSenderWhitelist(userID, fromEmail string) string {
	fromEmail = strings.ToLower(strings.TrimSpace(fromEmail))
	rules := s.ListSenderRules(userID)
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		pat := strings.ToLower(strings.TrimSpace(r.Pattern))
		if strings.HasSuffix(fromEmail, pat) || strings.Contains(fromEmail, pat) {
			return r.Pattern
		}
	}
	return ""
}

// ImportConfirmation validates mailbox consent and sender whitelist rules, then imports a single confirmation from the user's inbox.
func (s *SyncService) ImportConfirmation(userID, confirmationID string) (*domain.Trip, *domain.Reservation, error) {
	user, err := s.store.GetUser(userID)
	if err != nil {
		return nil, nil, err
	}
	if !user.MailboxAuthorized {
		return nil, nil, domain.ErrMailboxUnauthorized
	}

	conf, err := s.store.GetEmailConfirmation(confirmationID, userID)
	if err != nil {
		return nil, nil, err
	}

	matchedRule := s.matchSenderWhitelist(userID, conf.FromEmail)
	if matchedRule == "" {
		conf.Status = "BLOCKED"
		conf.StatusReason = fmt.Sprintf("Remitente %s fuera de tu lista de dominios confiables", conf.FromEmail)
		_ = s.store.SaveEmailConfirmation(conf)
		return nil, nil, domain.ErrSenderNotAllowed
	}

	msg := platform.ReservationDetectedMessage{
		EventID:          platform.NewID("evt"),
		UserID:           userID,
		Source:           domain.SourceEmailScan,
		ProviderName:     conf.ProviderName,
		Type:             conf.Type,
		Status:           domain.StatusConfirmed,
		ConfirmationCode: conf.PNR,
		Origin:           conf.Origin,
		Destination:      conf.Destination,
		StartDate:        conf.StartDate,
		EndDate:          conf.EndDate,
		Price:            conf.Price,
		Currency:         conf.Currency,
		Notes:            fmt.Sprintf("Sincronizada desde tu bandeja (%s)", matchedRule),
		Details:          conf.Details,
		Timestamp:        time.Now().UTC(),
	}

	trip, res, err := s.trips.IngestDetectedReservation(msg)
	if err != nil {
		return nil, nil, err
	}

	conf.Status = "IMPORTED"
	conf.StatusReason = fmt.Sprintf("Sincronizada en \"%s\"", trip.Title)
	conf.ImportedTripID = trip.ID
	_ = s.store.SaveEmailConfirmation(conf)
	return trip, res, nil
}

// SyncAllInbox scans all pending confirmations in the user's inbox against their sender whitelist.
func (s *SyncService) SyncAllInbox(userID string) (InboxSyncBatchResult, error) {
	user, err := s.store.GetUser(userID)
	if err != nil {
		return InboxSyncBatchResult{}, err
	}
	if !user.MailboxAuthorized {
		return InboxSyncBatchResult{}, domain.ErrMailboxUnauthorized
	}

	list := s.store.ListEmailConfirmations(userID)
	var imported, blocked int
	for _, c := range list {
		if c.Status == "IMPORTED" {
			continue
		}
		_, _, err := s.ImportConfirmation(userID, c.ID)
		if err == nil {
			imported++
		} else {
			blocked++
		}
	}

	if imported == 0 && blocked == 0 {
		return InboxSyncBatchResult{
			Message: "No hay comprobantes pendientes por sincronizar en tu bandeja.",
		}, nil
	}
	return InboxSyncBatchResult{
		ImportedCount: imported,
		BlockedCount:  blocked,
		Message:       fmt.Sprintf("Sincronización completada: %d reservas incorporadas a tus viajes (%d omitidas por filtro de remitente).", imported, blocked),
	}, nil
}

// ImportByPNR queries the external provider catalog by confirmation code (PNR) through the resilience circuit breaker,
// attaches the reservation to the user's trip, and records the receipt in the user's inbox.
func (s *SyncService) ImportByPNR(userID, pnrCode string) (*domain.Trip, *domain.Reservation, error) {
	user, err := s.store.GetUser(userID)
	if err != nil {
		return nil, nil, err
	}
	pnrCode = strings.ToUpper(strings.TrimSpace(pnrCode))
	if pnrCode == "" {
		return nil, nil, fmt.Errorf("%w: ingresa un código de reserva o localizador PNR válido", domain.ErrValidation)
	}

	var offer domain.CatalogOffer
	var findErr error
	execErr := s.guard.Execute(func() error {
		offer, findErr = s.store.FindCatalogOfferByPNR(pnrCode)
		return nil
	})
	if execErr != nil {
		return nil, nil, execErr
	}
	if findErr != nil {
		return nil, nil, fmt.Errorf("%w: no encontramos ninguna reserva emitida con el código %s", domain.ErrNotFound, pnrCode)
	}

	msg := platform.ReservationDetectedMessage{
		EventID:          platform.NewID("evt"),
		UserID:           user.ID,
		Source:           domain.SourceProviderAPI,
		ProviderName:     offer.ProviderName,
		Type:             offer.Type,
		Status:           domain.StatusConfirmed,
		ConfirmationCode: offer.SamplePNR,
		Origin:           offer.Origin,
		Destination:      offer.Destination,
		StartDate:        offer.StartDate,
		EndDate:          offer.EndDate,
		Price:            offer.Price,
		Currency:         offer.Currency,
		Notes:            fmt.Sprintf("Importada directamente por código PNR (%s)", offer.ProviderName),
		Details:          offer.Details,
		Timestamp:        time.Now().UTC(),
	}

	trip, res, err := s.trips.IngestDetectedReservation(msg)
	if err != nil {
		return nil, nil, err
	}

	fromEmail := offer.ProviderEmail
	if fromEmail == "" {
		fromEmail = providerSenderEmail(offer.ProviderName)
	}
	_ = s.store.SaveEmailConfirmation(domain.EmailConfirmation{
		ID:              platform.NewID("eml"),
		UserID:          user.ID,
		FromEmail:       fromEmail,
		ProviderName:    offer.ProviderName,
		Subject:         fmt.Sprintf("Comprobante PNR %s — %s (%s)", offer.SamplePNR, offer.ProviderName, offer.Title),
		Body:            fmt.Sprintf("Reserva recuperada por código PNR %s. Servicio: %s → %s.", offer.SamplePNR, offer.Origin, offer.Destination),
		PNR:             offer.SamplePNR,
		Type:            offer.Type,
		Origin:          offer.Origin,
		Destination:     offer.Destination,
		DestinationCity: offer.DestinationCity,
		CountryCode:     offer.CountryCode,
		StartDate:       offer.StartDate,
		EndDate:         offer.EndDate,
		Price:           offer.Price,
		Currency:        offer.Currency,
		Details:         offer.Details,
		Status:          "IMPORTED",
		StatusReason:    fmt.Sprintf("Importada por código PNR en \"%s\"", trip.Title),
		ImportedTripID:  trip.ID,
		ReceivedAt:      time.Now().UTC(),
	})

	return trip, res, nil
}

var pnrRegex = regexp.MustCompile(`(?i)(?:PNR|C[ÓO]DIGO|CONFIRMACI[ÓO]N|BOOKING|LOCALIZADOR)[:\s#]*([A-Z0-9-]{5,12})`)

// ScanEmail parses a raw confirmation email when mailbox consent and sender whitelist rules pass.
func (s *SyncService) ScanEmail(userID, fromEmail, subject, body string) (EmailScanResult, error) {
	user, err := s.store.GetUser(userID)
	if err != nil {
		return EmailScanResult{}, err
	}
	if !user.MailboxAuthorized {
		return EmailScanResult{}, domain.ErrMailboxUnauthorized
	}

	fromEmail = strings.ToLower(strings.TrimSpace(fromEmail))
	if fromEmail == "" {
		return EmailScanResult{}, fmt.Errorf("%w: indica la dirección del remitente", domain.ErrValidation)
	}

	matchedRule := s.matchSenderWhitelist(userID, fromEmail)
	if matchedRule == "" {
		return EmailScanResult{}, domain.ErrSenderNotAllowed
	}

	combined := subject + "\n" + body
	matches := pnrRegex.FindStringSubmatch(combined)
	if len(matches) < 2 {
		return EmailScanResult{}, domain.ErrNoValidReservation
	}
	code := strings.ToUpper(matches[1])

	// Check if code matches a known catalog offer first
	if offer, err := s.store.FindCatalogOfferByPNR(code); err == nil {
		trip, _, err := s.ImportByPNR(userID, offer.SamplePNR)
		if err != nil {
			return EmailScanResult{}, err
		}
		return EmailScanResult{
			Accepted:         true,
			Message:          fmt.Sprintf("Reserva %s (%s) incorporada al viaje \"%s\"", code, offer.ProviderName, trip.Title),
			MatchedRule:      matchedRule,
			ConfirmationCode: code,
			ReservationType:  offer.Type,
			ProviderName:     offer.ProviderName,
			TripID:           trip.ID,
		}, nil
	}

	resType, provider, origin, dest, price, details := classifyEmailContent(fromEmail, combined)
	start := time.Now().UTC().Add(72 * time.Hour)
	end := start.Add(4 * time.Hour)
	if resType == domain.ReservationHotel {
		end = start.Add(48 * time.Hour)
	}

	msg := platform.ReservationDetectedMessage{
		EventID:          platform.NewID("evt"),
		UserID:           userID,
		Source:           domain.SourceEmailScan,
		ProviderName:     provider,
		Type:             resType,
		Status:           domain.StatusConfirmed,
		ConfirmationCode: code,
		Origin:           origin,
		Destination:      dest,
		StartDate:        start,
		EndDate:          end,
		Price:            price,
		Currency:         "USD",
		Notes:            fmt.Sprintf("Importada desde correo verificado (%s)", matchedRule),
		Details:          details,
		Timestamp:        time.Now().UTC(),
	}

	trip, _, err := s.trips.IngestDetectedReservation(msg)
	if err != nil {
		return EmailScanResult{}, err
	}

	_ = s.store.SaveEmailConfirmation(domain.EmailConfirmation{
		ID:              platform.NewID("eml"),
		UserID:          userID,
		FromEmail:       fromEmail,
		ProviderName:    provider,
		Subject:         strings.TrimSpace(subject),
		Body:            strings.TrimSpace(body),
		PNR:             code,
		Type:            resType,
		Origin:          origin,
		Destination:     dest,
		DestinationCity: trip.DestinationCity,
		CountryCode:     trip.CountryCode,
		StartDate:       start,
		EndDate:         end,
		Price:           price,
		Currency:        "USD",
		Details:         details,
		Status:          "IMPORTED",
		StatusReason:    fmt.Sprintf("Importada desde correo (%s) en \"%s\"", matchedRule, trip.Title),
		ImportedTripID:  trip.ID,
		ReceivedAt:      time.Now().UTC(),
	})

	return EmailScanResult{
		Accepted:         true,
		Message:          fmt.Sprintf("Reserva %s (%s) incorporada al viaje \"%s\"", code, provider, trip.Title),
		MatchedRule:      matchedRule,
		ConfirmationCode: code,
		ReservationType:  resType,
		ProviderName:     provider,
		TripID:           trip.ID,
	}, nil
}

func classifyEmailContent(fromEmail, text string) (domain.ReservationType, string, string, string, float64, domain.ReservationDetails) {
	lower := strings.ToLower(fromEmail + " " + text)
	switch {
	case strings.Contains(lower, "shinkansen") || strings.Contains(lower, "jrpass") || strings.Contains(lower, "tren") || strings.Contains(lower, "renfe"):
		return domain.ReservationTrain, "JR Central Shinkansen", "Tokio Station", "Kioto Station", 138.00, domain.ReservationDetails{
			TrainLine:        "Tokaido Shinkansen",
			TrainNumber:      "Nozomi 215",
			DepartureStation: "Tokio Central",
			ArrivalStation:   "Kioto Central",
			Wagon:            "Coche 5",
			Seat:             "8A",
		}
	case strings.Contains(lower, "alsa") || strings.Contains(lower, "bus") || strings.Contains(lower, "autobús"):
		return domain.ReservationBus, "ALSA Autobuses", "Madrid Estación Sur", "Granada", 39.50, domain.ReservationDetails{
			BusCompany:        "ALSA",
			LineNumber:        "SUP-112",
			DepartureTerminal: "Estación Sur",
			ArrivalTerminal:   "Estación de Granada",
			Platform:          "9",
			Seat:              "4B",
		}
	case strings.Contains(lower, "booking") || strings.Contains(lower, "hotel") || strings.Contains(lower, "check-in"):
		return domain.ReservationHotel, "Booking.com", "Tokio", "Hotel Gracery Shinjuku", 420.00, domain.ReservationDetails{
			HotelName:    "Hotel Gracery Shinjuku",
			Address:      "1-19-1 Kabukicho, Shinjuku-ku, Tokio",
			Latitude:     35.6954,
			Longitude:    139.7019,
			RoomType:     "Double Superior",
			CheckInTime:  "14:00",
			CheckOutTime: "11:00",
		}
	case strings.Contains(lower, "hertz") || strings.Contains(lower, "auto") || strings.Contains(lower, "vehículo") || strings.Contains(lower, "rental"):
		return domain.ReservationRental, "Hertz Rent a Car", "Aeropuerto Bariloche (BRC)", "Circuito Chico", 185.00, domain.ReservationDetails{
			VehicleCategory:   "SUV 4x4 Automática",
			PickupLocation:    "Aeropuerto Bariloche",
			DropoffLocation:   "Aeropuerto Bariloche",
			InsuranceIncluded: true,
		}
	default:
		return domain.ReservationFlight, "All Nippon Airways", "Tokio Haneda (HND)", "Sapporo (CTS)", 210.00, domain.ReservationDetails{
			FlightNumber:     "NH 065",
			DepartureAirport: "HND",
			ArrivalAirport:   "CTS",
			Terminal:         "Terminal 2",
			Gate:             "61",
			Seat:             "12C",
			CabinClass:       "Economy",
		}
	}
}
