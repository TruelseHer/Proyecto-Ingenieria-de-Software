package service

import (
	"fmt"
	"strings"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
)

type CatalogService struct {
	store *repository.Store
	trips *TripService
}

func NewCatalogService(store *repository.Store, trips *TripService) *CatalogService {
	return &CatalogService{
		store: store,
		trips: trips,
	}
}

func (s *CatalogService) ListPackages() []domain.CatalogPackage {
	return s.store.ListCatalogPackages()
}

func (s *CatalogService) GetPackage(id string) (domain.CatalogPackage, error) {
	return s.store.GetCatalogPackage(id)
}

func (s *CatalogService) ListOffers(filterType, query string) []domain.CatalogOffer {
	return s.store.ListCatalogOffers(filterType, query)
}

func (s *CatalogService) GetOffer(id string) (domain.CatalogOffer, error) {
	return s.store.GetCatalogOffer(id)
}

// BookPackage books a full multi-modal travel package on the chosen departure date,
// creates the Trip with all its flight, hotel, and ground segments, and delivers the confirmation vouchers to the user's inbox.
func (s *CatalogService) BookPackage(userID, packageID, dateID string) (*domain.Trip, int, error) {
	if _, err := s.store.GetUser(userID); err != nil {
		return nil, 0, err
	}
	pkg, err := s.store.GetCatalogPackage(packageID)
	if err != nil {
		return nil, 0, err
	}

	var chosenDate domain.CatalogPackageDate
	if len(pkg.AvailableDates) > 0 {
		chosenDate = pkg.AvailableDates[0]
		for _, d := range pkg.AvailableDates {
			if d.ID == dateID {
				chosenDate = d
				break
			}
		}
	} else {
		now := time.Now().UTC().Add(14 * 24 * time.Hour)
		chosenDate = domain.CatalogPackageDate{
			ID:             "default",
			Label:          "Próxima salida",
			StartDate:      now,
			EndDate:        now.Add(time.Duration(pkg.DurationDays) * 24 * time.Hour),
			SpotsAvailable: 8,
			PriceModifier:  1.0,
		}
	}

	now := time.Now().UTC()
	trip := &domain.Trip{
		ID:               platform.NewID("trip"),
		OwnerID:          userID,
		SharedWithEmails: []string{},
		Title:            pkg.Title,
		DestinationCity:  pkg.DestinationCity,
		CountryCode:      pkg.CountryCode,
		CoverImageURL:    pkg.ImageURL,
		Status:           domain.TripStatusActive,
		StartDate:        chosenDate.StartDate,
		EndDate:          chosenDate.EndDate,
		Reservations:     []*domain.Reservation{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	totalPrice := pkg.PricePerPerson
	if chosenDate.PriceModifier > 0 {
		totalPrice = totalPrice * chosenDate.PriceModifier
	}

	for idx, seg := range pkg.IncludedSegments {
		segStart := chosenDate.StartDate.Add(time.Duration(seg.DayOffsetStart)*24*time.Hour + time.Duration(seg.HourStart)*time.Hour)
		durHours := seg.DurationHours
		if durHours <= 0 {
			durHours = 4
		}
		segEnd := segStart.Add(time.Duration(durHours) * time.Hour)
		segPrice := totalPrice * seg.PriceShare
		if segPrice <= 0 {
			segPrice = totalPrice / float64(len(pkg.IncludedSegments))
		}

		prefix := seg.ConfirmationPrefx
		if prefix == "" {
			prefix = "TRM"
		}
		pnr := fmt.Sprintf("%s-%d%02d", strings.ToUpper(prefix), (now.Unix()%900)+100, idx+1)

		res := &domain.Reservation{
			ID:               platform.NewID("res"),
			TripID:           trip.ID,
			Type:             seg.Type,
			Status:           domain.StatusConfirmed,
			ProviderName:     seg.ProviderName,
			ConfirmationCode: pnr,
			Origin:           seg.Origin,
			Destination:      seg.Destination,
			StartDate:        segStart,
			EndDate:          segEnd,
			Price:            segPrice,
			Currency:         pkg.Currency,
			Notes:            seg.Notes,
			Source:           domain.SourceCatalog,
			Details:          seg.Details,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		trip.Reservations = append(trip.Reservations, res)

		// Deliver confirmation voucher to user's inbox
		_ = s.store.SaveEmailConfirmation(domain.EmailConfirmation{
			ID:              platform.NewID("eml"),
			UserID:          userID,
			FromEmail:       providerSenderEmail(seg.ProviderName),
			ProviderName:    seg.ProviderName,
			Subject:         fmt.Sprintf("Confirmación de reserva %s — %s (%s)", res.TypeLabel(), seg.ProviderName, pnr),
			Body:            fmt.Sprintf("Tu reserva del paquete \"%s\" está confirmada. Código PNR: %s. Trayecto/Servicio: %s → %s.", pkg.Title, pnr, seg.Origin, seg.Destination),
			PNR:             pnr,
			Type:            seg.Type,
			Origin:          seg.Origin,
			Destination:     seg.Destination,
			DestinationCity: pkg.DestinationCity,
			CountryCode:     pkg.CountryCode,
			StartDate:       segStart,
			EndDate:         segEnd,
			Price:           segPrice,
			Currency:        pkg.Currency,
			Details:         seg.Details,
			Status:          "IMPORTED",
			StatusReason:    fmt.Sprintf("Incorporada al reservar el paquete \"%s\"", pkg.Title),
			ImportedTripID:  trip.ID,
			ReceivedAt:      now,
		})
	}

	trip.RecalculateBounds()
	if err := s.store.SaveTrip(trip); err != nil {
		return nil, 0, err
	}
	if s.trips != nil {
		s.trips.invalidateTripMembers(trip)
		s.trips.emitLifecycle(trip, "TRIP_CREATED", false, pkg.DestinationCity)
	}
	return trip, len(trip.Reservations), nil
}

// SendPackageToInbox sends the confirmation emails of a package to the user's inbox as PENDING
// so the user can synchronize them via the inbox email scanner.
func (s *CatalogService) SendPackageToInbox(userID, packageID, dateID string) ([]domain.EmailConfirmation, error) {
	if _, err := s.store.GetUser(userID); err != nil {
		return nil, err
	}
	pkg, err := s.store.GetCatalogPackage(packageID)
	if err != nil {
		return nil, err
	}
	var chosenDate domain.CatalogPackageDate
	if len(pkg.AvailableDates) > 0 {
		chosenDate = pkg.AvailableDates[0]
		for _, d := range pkg.AvailableDates {
			if d.ID == dateID {
				chosenDate = d
				break
			}
		}
	} else {
		now := time.Now().UTC().Add(14 * 24 * time.Hour)
		chosenDate = domain.CatalogPackageDate{
			StartDate:     now,
			EndDate:       now.Add(7 * 24 * time.Hour),
			PriceModifier: 1.0,
		}
	}

	now := time.Now().UTC()
	var created []domain.EmailConfirmation
	for idx, seg := range pkg.IncludedSegments {
		segStart := chosenDate.StartDate.Add(time.Duration(seg.DayOffsetStart)*24*time.Hour + time.Duration(seg.HourStart)*time.Hour)
		durHours := seg.DurationHours
		if durHours <= 0 {
			durHours = 4
		}
		segEnd := segStart.Add(time.Duration(durHours) * time.Hour)
		segPrice := pkg.PricePerPerson * seg.PriceShare
		prefix := seg.ConfirmationPrefx
		if prefix == "" {
			prefix = "TRM"
		}
		pnr := fmt.Sprintf("%s-%d%02d", strings.ToUpper(prefix), (now.Unix()%900)+100, idx+1)

		tmpRes := domain.Reservation{Type: seg.Type}
		conf := domain.EmailConfirmation{
			ID:              platform.NewID("eml"),
			UserID:          userID,
			FromEmail:       providerSenderEmail(seg.ProviderName),
			ProviderName:    seg.ProviderName,
			Subject:         fmt.Sprintf("Comprobante de %s — %s (PNR: %s)", tmpRes.TypeLabel(), seg.ProviderName, pnr),
			Body:            fmt.Sprintf("Confirmación emitida para %s. Código PNR: %s. Origen: %s — Destino: %s.", pkg.Title, pnr, seg.Origin, seg.Destination),
			PNR:             pnr,
			Type:            seg.Type,
			Origin:          seg.Origin,
			Destination:     seg.Destination,
			DestinationCity: pkg.DestinationCity,
			CountryCode:     pkg.CountryCode,
			StartDate:       segStart,
			EndDate:         segEnd,
			Price:           segPrice,
			Currency:        pkg.Currency,
			Details:         seg.Details,
			Status:          "PENDING",
			StatusReason:    "Pendiente de sincronizar con tu itinerario",
			ReceivedAt:      now,
		}
		if err := s.store.SaveEmailConfirmation(conf); err == nil {
			created = append(created, conf)
		}
	}
	return created, nil
}

// BookOffer books an individual catalog segment (flight, hotel, train, bus, or car rental),
// attaches it to the selected or auto-matched trip, and records the confirmation in the user's inbox.
func (s *CatalogService) BookOffer(userID, offerID, targetTripID string) (*domain.Trip, *domain.Reservation, error) {
	return s.BookOfferWithOptions(userID, offerID, targetTripID, "")
}

// BookOfferWithOptions books an individual catalog segment into an existing trip,
// creates a brand-new trip when targetTripID is "NEW", or auto-matches by destination when "AUTO".
func (s *CatalogService) BookOfferWithOptions(userID, offerID, targetTripID, newTripTitle string) (*domain.Trip, *domain.Reservation, error) {
	if _, err := s.store.GetUser(userID); err != nil {
		return nil, nil, err
	}
	offer, err := s.store.GetCatalogOffer(offerID)
	if err != nil {
		return nil, nil, err
	}

	pnr := offer.SamplePNR
	if pnr == "" {
		prefix := offer.ConfirmationPrefix
		if prefix == "" {
			prefix = "RES"
		}
		pnr = fmt.Sprintf("%s-%d", prefix, (time.Now().Unix()%9000)+1000)
	}

	var trip *domain.Trip
	var res *domain.Reservation
	targetTripID = strings.TrimSpace(targetTripID)
	newTripTitle = strings.TrimSpace(newTripTitle)

	buildRes := func() *domain.Reservation {
		return &domain.Reservation{
			Type:             offer.Type,
			Status:           domain.StatusConfirmed,
			ProviderName:     offer.ProviderName,
			ConfirmationCode: pnr,
			Origin:           offer.Origin,
			Destination:      offer.Destination,
			StartDate:        offer.StartDate,
			EndDate:          offer.EndDate,
			Price:            offer.Price,
			Currency:         offer.Currency,
			Notes:            offer.Title,
			Source:           domain.SourceCatalog,
			Details:          offer.Details,
		}
	}

	switch {
	case strings.EqualFold(targetTripID, "NEW"):
		destCity := strings.TrimSpace(offer.DestinationCity)
		if destCity == "" {
			destCity = offer.Destination
		}
		title := newTripTitle
		if title == "" {
			title = fmt.Sprintf("Viaje a %s", destCity)
		}
		createdTrip, cErr := s.trips.CreateTrip(userID, title, destCity, offer.CountryCode)
		if cErr != nil {
			return nil, nil, cErr
		}
		trip, res, err = s.trips.AddReservation(userID, createdTrip.ID, buildRes())
		if err != nil {
			return nil, nil, err
		}
	case targetTripID != "" && !strings.EqualFold(targetTripID, "AUTO"):
		trip, res, err = s.trips.AddReservation(userID, targetTripID, buildRes())
		if err != nil {
			return nil, nil, err
		}
	default:
		msg := platform.ReservationDetectedMessage{
			EventID:          platform.NewID("evt"),
			UserID:           userID,
			Source:           domain.SourceCatalog,
			ProviderName:     offer.ProviderName,
			Type:             offer.Type,
			Status:           domain.StatusConfirmed,
			ConfirmationCode: pnr,
			Origin:           offer.Origin,
			Destination:      offer.Destination,
			StartDate:        offer.StartDate,
			EndDate:          offer.EndDate,
			Price:            offer.Price,
			Currency:         offer.Currency,
			Notes:            offer.Title,
			Details:          offer.Details,
			Timestamp:        time.Now().UTC(),
		}
		trip, res, err = s.trips.IngestDetectedReservation(msg)
		if err != nil {
			return nil, nil, err
		}
	}

	fromEmail := offer.ProviderEmail
	if fromEmail == "" {
		fromEmail = providerSenderEmail(offer.ProviderName)
	}
	_ = s.store.SaveEmailConfirmation(domain.EmailConfirmation{
		ID:              platform.NewID("eml"),
		UserID:          userID,
		FromEmail:       fromEmail,
		ProviderName:    offer.ProviderName,
		Subject:         fmt.Sprintf("Confirmación %s — %s (%s)", offer.TypeLabel(), offer.Title, pnr),
		Body:            fmt.Sprintf("Reserva confirmada en el catálogo Trotamundos. Código PNR: %s. %s → %s.", pnr, offer.Origin, offer.Destination),
		PNR:             pnr,
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
		StatusReason:    fmt.Sprintf("Agregada al viaje \"%s\"", trip.Title),
		ImportedTripID:  trip.ID,
		ReceivedAt:      time.Now().UTC(),
	})

	return trip, res, nil
}

// SendOfferToInbox delivers a single catalog offer confirmation email into the user's inbox as PENDING.
func (s *CatalogService) SendOfferToInbox(userID, offerID string) (domain.EmailConfirmation, error) {
	if _, err := s.store.GetUser(userID); err != nil {
		return domain.EmailConfirmation{}, err
	}
	offer, err := s.store.GetCatalogOffer(offerID)
	if err != nil {
		return domain.EmailConfirmation{}, err
	}
	pnr := offer.SamplePNR
	if pnr == "" {
		pnr = fmt.Sprintf("%s-%d", offer.ConfirmationPrefix, (time.Now().Unix()%9000)+1000)
	}
	fromEmail := offer.ProviderEmail
	if fromEmail == "" {
		fromEmail = providerSenderEmail(offer.ProviderName)
	}
	conf := domain.EmailConfirmation{
		ID:              platform.NewID("eml"),
		UserID:          userID,
		FromEmail:       fromEmail,
		ProviderName:    offer.ProviderName,
		Subject:         fmt.Sprintf("Confirmación de %s: %s (PNR: %s)", offer.TypeLabel(), offer.Title, pnr),
		Body:            fmt.Sprintf("Hemos emitido tu comprobante con %s. Localizador PNR: %s. Servicio: %s → %s.", offer.ProviderName, pnr, offer.Origin, offer.Destination),
		PNR:             pnr,
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
		Status:          "PENDING",
		StatusReason:    "Pendiente de sincronizar en tu bandeja",
		ReceivedAt:      time.Now().UTC(),
	}
	if err := s.store.SaveEmailConfirmation(conf); err != nil {
		return domain.EmailConfirmation{}, err
	}
	return conf, nil
}

func providerSenderEmail(providerName string) string {
	lower := strings.ToLower(providerName)
	switch {
	case strings.Contains(lower, "ana") || strings.Contains(lower, "nippon"):
		return "confirmaciones@ana.co.jp"
	case strings.Contains(lower, "jr") || strings.Contains(lower, "shinkansen"):
		return "reservas@jrpass.com"
	case strings.Contains(lower, "iberia"):
		return "billetes@iberia.com"
	case strings.Contains(lower, "renfe") || strings.Contains(lower, "ave"):
		return "venta-billetes@renfe.com"
	case strings.Contains(lower, "aerolíneas") || strings.Contains(lower, "aerolineas"):
		return "confirmaciones@aerolineas.com.ar"
	case strings.Contains(lower, "latam"):
		return "no-reply@latam.com"
	case strings.Contains(lower, "hertz") || strings.Contains(lower, "avis"):
		return "reservas@hertz.com"
	case strings.Contains(lower, "alsa") || strings.Contains(lower, "bus") || strings.Contains(lower, "bariloche"):
		return "boletos@alsa.es"
	case strings.Contains(lower, "eurostar") || strings.Contains(lower, "frecciarossa") || strings.Contains(lower, "perurail"):
		return "tickets@jrpass.com"
	default:
		return "noreply@booking.com"
	}
}
