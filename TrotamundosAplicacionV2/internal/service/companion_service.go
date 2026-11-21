package service

import (
	"fmt"
	"math"
	"strings"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
)

type CompanionService struct {
	store  *repository.Store
	cache  *platform.CacheClient
	tokens *platform.TokenService
	extAPI *platform.ExternalAPIClient
}

func NewCompanionService(store *repository.Store, cache *platform.CacheClient, tokens *platform.TokenService, extAPI *platform.ExternalAPIClient) *CompanionService {
	if extAPI == nil {
		extAPI = platform.NewExternalAPIClient(cache, 1800*time.Millisecond)
	}
	return &CompanionService{
		store:  store,
		cache:  cache,
		tokens: tokens,
		extAPI: extAPI,
	}
}

type geoSeed struct {
	tz            string
	localCurrency string
	airportLabel  string
	airportLat    float64
	airportLon    float64
	stationLabel  string
	stationLat    float64
	stationLon    float64
	hotelFallback string
	hotelAddr     string
	hotelLat      float64
	hotelLon      float64
	sampleFare    float64
}

var cityMetadata = map[string]geoSeed{
	"JP": {
		tz:            "Asia/Tokyo",
		localCurrency: "JPY",
		airportLabel:  "Aeropuerto Tokio Haneda (HND)",
		airportLat:    35.5494,
		airportLon:    139.7798,
		stationLabel:  "Estación Central de Tokio",
		stationLat:    35.6812,
		stationLon:    139.7671,
		hotelFallback: "Hotel Gracery Shinjuku",
		hotelAddr:     "1-19-1 Kabukicho, Shinjuku-ku, Tokio",
		hotelLat:      35.6954,
		hotelLon:      139.7019,
		sampleFare:    2400,
	},
	"ES": {
		tz:            "Europe/Madrid",
		localCurrency: "EUR",
		airportLabel:  "Aeropuerto Madrid Barajas T4 (MAD)",
		airportLat:    40.4983,
		airportLon:    -3.5676,
		stationLabel:  "Estación Madrid Puerta de Atocha",
		stationLat:    40.4065,
		stationLon:    -3.6897,
		hotelFallback: "Hotel Riu Plaza España",
		hotelAddr:     "Gran Vía 84, Centro, Madrid",
		hotelLat:      40.4241,
		hotelLon:      -3.7102,
		sampleFare:    33.00,
	},
	"AR": {
		tz:            "America/Argentina/Buenos_Aires",
		localCurrency: "ARS",
		airportLabel:  "Aeropuerto San Carlos de Bariloche (BRC)",
		airportLat:    -41.1512,
		airportLon:    -71.1575,
		stationLabel:  "Centro Cívico de Bariloche",
		stationLat:    -41.1335,
		stationLon:    -71.3103,
		hotelFallback: "Llao Llao Resort, Golf & Spa",
		hotelAddr:     "Av. Ezequiel Bustillo Km 25, Bariloche",
		hotelLat:      -41.0583,
		hotelLon:      -71.5332,
		sampleFare:    22500,
	},
	"IT": {
		tz:            "Europe/Rome",
		localCurrency: "EUR",
		airportLabel:  "Aeropuerto Roma Fiumicino (FCO)",
		airportLat:    41.8003,
		airportLon:    12.2389,
		stationLabel:  "Estación Roma Termini",
		stationLat:    41.9010,
		stationLon:    12.5018,
		hotelFallback: "Hotel Artemide Roma",
		hotelAddr:     "Via Nazionale 22, Roma",
		hotelLat:      41.9008,
		hotelLon:      12.4937,
		sampleFare:    50.00,
	},
	"PE": {
		tz:            "America/Lima",
		localCurrency: "PEN",
		airportLabel:  "Aeropuerto Alejandro Velasco Astete (CUZ)",
		airportLat:    -13.5357,
		airportLon:    -71.9388,
		stationLabel:  "Estación Wanchaq Cusco",
		stationLat:    -13.5244,
		stationLon:    -71.9675,
		hotelFallback: "Palacio del Inka Luxury Collection",
		hotelAddr:     "Plazoleta Santo Domingo 259, Cusco",
		hotelLat:      -13.5186,
		hotelLon:      -71.9752,
		sampleFare:    45.00,
	},
	"FR": {
		tz:            "Europe/Paris",
		localCurrency: "EUR",
		airportLabel:  "Aeropuerto París Charles de Gaulle (CDG)",
		airportLat:    49.0097,
		airportLon:    2.5479,
		stationLabel:  "Gare du Nord, París",
		stationLat:    48.8809,
		stationLon:    2.3553,
		hotelFallback: "Pullman Paris Tour Eiffel",
		hotelAddr:     "18 Avenue de Suffren, París",
		hotelLat:      48.8556,
		hotelLon:      2.2933,
		sampleFare:    56.00,
	},
}

func (s *CompanionService) BuildNavigationRoute(userID, tripID, targetCurrency, originMode string) (*domain.NavigationRoute, error) {
	start := time.Now()
	trip, err := s.store.GetTrip(tripID)
	if err != nil {
		return nil, err
	}
	user, uErr := s.store.GetUser(userID)
	canAccess := trip.CanAccess(userID)
	if !canAccess && uErr == nil {
		canAccess = trip.CanAccessByUser(user)
	}
	if !canAccess {
		_ = s.store.AppendAuditLog(domain.AuditLogEntry{
			ID:           platform.NewID("aud"),
			ActorUserID:  userID,
			Action:       "MOBILE_NAVIGATION",
			ResourceType: "TRIP",
			ResourceID:   tripID,
			Outcome:      "DENIED",
			Reason:       "Acceso denegado al contexto de navegación del viaje",
			IPAddress:    "127.0.0.1",
			CreatedAt:    time.Now().UTC(),
		})
		return nil, domain.ErrForbidden
	}

	if targetCurrency == "" {
		targetCurrency = user.HomeCurrency
	}
	if targetCurrency == "" {
		targetCurrency = "USD"
	}
	targetCurrency = strings.ToUpper(targetCurrency)

	meta, ok := cityMetadata[strings.ToUpper(trip.CountryCode)]
	if !ok {
		meta = cityMetadata["JP"]
	}

	hotelName := meta.hotelFallback
	hotelAddr := meta.hotelAddr
	hotelLat := meta.hotelLat
	hotelLon := meta.hotelLon
	if hotel := trip.PrimaryHotel(); hotel != nil {
		if hotel.Details.HotelName != "" {
			hotelName = hotel.Details.HotelName
		}
		if hotel.Details.Address != "" {
			hotelAddr = hotel.Details.Address
		}
		if hotel.Details.Latitude != 0 && hotel.Details.Longitude != 0 {
			hotelLat = hotel.Details.Latitude
			hotelLon = hotel.Details.Longitude
		}
	}

	originLabel := meta.airportLabel
	originLat := meta.airportLat
	originLon := meta.airportLon
	if originMode == "station" {
		originLabel = meta.stationLabel
		originLat = meta.stationLat
		originLon = meta.stationLon
	}

	polyCoords, distKm, taxiMins, _ := s.extAPI.GetRouteGeometry(originLat, originLon, hotelLat, hotelLon)
	transitMins := int(math.Round(float64(taxiMins)*1.25)) + 4

	rate, rateSource := s.extAPI.GetExchangeRate(meta.localCurrency, targetCurrency)
	sampleTarget := math.Round((meta.sampleFare*rate)*100) / 100

	weather := s.extAPI.GetWeather(hotelLat, hotelLon, trip.DestinationCity)

	loc, tzErr := time.LoadLocation(meta.tz)
	localNow := time.Now().UTC()
	if tzErr == nil {
		localNow = localNow.In(loc)
	}

	routingToken := ""
	if s.tokens != nil {
		routingToken, _ = s.tokens.Issue(userID, "route", 15*time.Minute)
	}

	steps := buildDirectionsSteps(originLabel, hotelName, hotelAddr, meta.localCurrency, distKm, transitMins)
	elapsed := time.Since(start).Milliseconds()
	if elapsed == 0 {
		elapsed = 4
	}

	return &domain.NavigationRoute{
		TripID:              trip.ID,
		TripTitle:           trip.Title,
		DestinationCity:     trip.DestinationCity,
		CountryCode:         trip.CountryCode,
		HotelName:           hotelName,
		HotelAddress:        hotelAddr,
		HotelLat:            hotelLat,
		HotelLon:            hotelLon,
		OriginLabel:         originLabel,
		OriginLat:           originLat,
		OriginLon:           originLon,
		DistanceKm:          distKm,
		TransitMinutes:      transitMins,
		TaxiMinutes:         taxiMins,
		DestinationTimezone: meta.tz,
		LocalTimeFormatted:  localNow.Format("15:04 (MST)"),
		LocalCurrency:       meta.localCurrency,
		TargetCurrency:      targetCurrency,
		ExchangeRate:        rate,
		RateSource:          rateSource,
		SampleFareLocal:     meta.sampleFare,
		SampleFareTarget:    sampleTarget,
		RouteSVGPoints:      "42,154 115,128 190,94 252,72 314,44",
		PolylineCoords:      polyCoords,
		Steps:               steps,
		Weather:             weather,
		RoutingToken:        routingToken,
		GeneratedInMs:       elapsed,
	}, nil
}

func buildDirectionsSteps(originLabel, hotelName, hotelAddr, localCurrency string, distKm float64, mins int) []string {
	return []string{
		fmt.Sprintf("Sal desde %s siguiendo los carteles de conexión rápida al centro urbano.", originLabel),
		fmt.Sprintf("Avanza por el corredor principal hacia la zona del alojamiento (%.1f km de recorrido).", distKm),
		fmt.Sprintf("Llega a %s (%s) en aproximadamente %d minutos.", hotelName, hotelAddr, mins),
		fmt.Sprintf("Ten lista tu reserva en Trotamundos y medio de pago en %s para el check-in.", localCurrency),
	}
}
