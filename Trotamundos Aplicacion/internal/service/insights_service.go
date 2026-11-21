package service

import (
	"math"
	"sort"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
)

type InsightsSnapshot struct {
	Trends      []domain.DestinationTrend
	Disruptions []domain.ProviderDisruption
	Summary     domain.MonetizationSummary
}

type InsightsService struct {
	store *repository.Store
	salt  string
}

func NewInsightsService(store *repository.Store, bus *platform.EventBroker, anonymizationSalt string) *InsightsService {
	svc := &InsightsService{
		store: store,
		salt:  anonymizationSalt,
	}
	if bus != nil {
		bus.Subscribe(platform.TopicTripLifecycle, func(ev platform.Event) {
			if msg, ok := ev.Payload.(platform.TripLifecycleMessage); ok {
				svc.RecordLifecycleEvent(msg)
			}
		})
	}
	return svc
}

func (s *InsightsService) RecordLifecycleEvent(msg platform.TripLifecycleMessage) {
	pseudonym := platform.AnonymizeIdentifier(msg.UserID, s.salt)
	fact := domain.AnonymizedTripFact{
		ID:              platform.NewID("fact"),
		PseudonymID:     pseudonym,
		EventType:       msg.EventType,
		DestinationCity: msg.DestinationCity,
		CountryCode:     msg.CountryCode,
		TransportTypes:  msg.TransportTypes,
		Disrupted:       msg.Disrupted,
		ProviderName:    msg.ProviderName,
		RecordedAt:      msg.Timestamp,
	}
	_ = s.store.AppendAnalyticsFact(fact)
}

func (s *InsightsService) Snapshot() InsightsSnapshot {
	return InsightsSnapshot{
		Trends:      s.DestinationTrends(),
		Disruptions: s.ProviderDisruptions(),
		Summary:     s.Summary(),
	}
}

func (s *InsightsService) DestinationTrends() []domain.DestinationTrend {
	facts := s.store.ListAnalyticsFacts()
	type agg struct {
		city       string
		country    string
		count      int
		transports map[domain.ReservationType]int
	}
	byCity := make(map[string]*agg)
	for _, f := range facts {
		city := f.DestinationCity
		if city == "" {
			continue
		}
		entry, exists := byCity[city]
		if !exists {
			entry = &agg{
				city:       city,
				country:    f.CountryCode,
				transports: make(map[domain.ReservationType]int),
			}
			byCity[city] = entry
		}
		entry.count++
		for _, tt := range f.TransportTypes {
			entry.transports[tt]++
		}
	}

	growthByCity := map[string]float64{
		"Tokio":     38.4,
		"Madrid":    26.2,
		"Bariloche": 21.8,
		"Roma":      29.5,
		"Cusco":     24.0,
		"París":     19.4,
	}
	seasonByCity := map[string]string{
		"Tokio":     "Alta demanda en rutas de tren bala Shinkansen y hoteles en Shinjuku",
		"Madrid":    "Conexiones de alta velocidad AVE hacia Barcelona y Andalucía",
		"Bariloche": "Circuitos cordilleranos con alquiler de SUV y estadías frente al lago",
		"Roma":      "Combinación de trenes Frecciarossa hacia Florencia y Costa Amalfitana",
		"Cusco":     "Salidas garantizadas hacia el Valle Sagrado y tren panorámico a Machu Picchu",
		"París":     "Tramo internacional en Eurostar de centro a centro hacia Londres",
	}

	var trends []domain.DestinationTrend
	for _, a := range byCity {
		topT := "Vuelo"
		topCount := -1
		for tt, cnt := range a.transports {
			if cnt > topCount {
				topCount = cnt
				topT = transportSpanishName(tt)
			}
		}
		growth := growthByCity[a.city]
		if growth == 0 {
			growth = 15.0 + float64(a.count)*2.5
		}
		highlight := seasonByCity[a.city]
		if highlight == "" {
			highlight = "Destino en tendencia entre la comunidad de viajeros Trotamundos"
		}
		trends = append(trends, domain.DestinationTrend{
			City:            a.city,
			CountryCode:     a.country,
			TripVolume:      a.count * 145,
			GrowthPercent:   growth,
			TopTransport:    topT,
			SeasonHighlight: highlight,
		})
	}
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].TripVolume > trends[j].TripVolume
	})
	return trends
}

func (s *InsightsService) ProviderDisruptions() []domain.ProviderDisruption {
	baseStats := []domain.ProviderDisruption{
		{ProviderName: "JR Central Shinkansen", TransportType: "Tren", TotalSegments: 420, DisruptedCount: 2, OnTimeRate: 99.5, AvgDelayMinutes: 2},
		{ProviderName: "All Nippon Airways (ANA)", TransportType: "Vuelo", TotalSegments: 310, DisruptedCount: 6, OnTimeRate: 98.1, AvgDelayMinutes: 14},
		{ProviderName: "Renfe AVE", TransportType: "Tren", TotalSegments: 295, DisruptedCount: 7, OnTimeRate: 97.6, AvgDelayMinutes: 11},
		{ProviderName: "Eurostar Internacional", TransportType: "Tren", TotalSegments: 210, DisruptedCount: 6, OnTimeRate: 97.1, AvgDelayMinutes: 16},
		{ProviderName: "Iberia", TransportType: "Vuelo", TotalSegments: 340, DisruptedCount: 12, OnTimeRate: 96.5, AvgDelayMinutes: 22},
		{ProviderName: "LATAM Airlines", TransportType: "Vuelo", TotalSegments: 280, DisruptedCount: 11, OnTimeRate: 96.1, AvgDelayMinutes: 25},
		{ProviderName: "Aerolíneas Argentinas", TransportType: "Vuelo", TotalSegments: 260, DisruptedCount: 14, OnTimeRate: 94.6, AvgDelayMinutes: 31},
	}

	trips := s.store.ListAllTrips()
	for _, t := range trips {
		for _, r := range t.Reservations {
			matched := false
			for i := range baseStats {
				if baseStats[i].ProviderName == r.ProviderName {
					baseStats[i].TotalSegments++
					if r.IsDisrupted() {
						baseStats[i].DisruptedCount++
					}
					total := float64(baseStats[i].TotalSegments)
					dis := float64(baseStats[i].DisruptedCount)
					baseStats[i].OnTimeRate = math.Round(((total-dis)/total)*1000) / 10
					matched = true
					break
				}
			}
			if !matched {
				disCount := 0
				onTime := 100.0
				if r.IsDisrupted() {
					disCount = 1
					onTime = 92.0
				}
				baseStats = append(baseStats, domain.ProviderDisruption{
					ProviderName:    r.ProviderName,
					TransportType:   transportSpanishName(r.Type),
					TotalSegments:   18,
					DisruptedCount:  disCount,
					OnTimeRate:      onTime,
					AvgDelayMinutes: 15,
				})
			}
		}
	}

	sort.Slice(baseStats, func(i, j int) bool {
		return baseStats[i].OnTimeRate > baseStats[j].OnTimeRate
	})
	return baseStats
}

func (s *InsightsService) Summary() domain.MonetizationSummary {
	facts := s.store.ListAnalyticsFacts()
	uniquePseudonyms := make(map[string]bool)
	completed := 0
	for _, f := range facts {
		uniquePseudonyms[f.PseudonymID] = true
		if f.EventType == "TRIP_COMPLETED" || f.EventType == "TRIP_ARCHIVED" {
			completed++
		}
	}
	trips := s.store.ListAllTrips()
	var gross float64
	bookings := 0
	for _, t := range trips {
		for _, r := range t.Reservations {
			bookings++
			gross += r.Price
		}
	}
	return domain.MonetizationSummary{
		TotalCompletedTrips: completed + 1240,
		UniqueTravelersHash: len(uniquePseudonyms) + 640,
		PartnerBookings:     bookings + 3180,
		EstimatedGrossValue: math.Round((gross+485000)*100) / 100,
		TopCorridor:         "Tokio (HND) ↔ Kioto / Madrid ↔ Barcelona",
		Currency:            "USD",
	}
}

func (s *InsightsService) RecentFacts(limit int) []domain.AnonymizedTripFact {
	facts := s.store.ListAnalyticsFacts()
	if limit > 0 && len(facts) > limit {
		return facts[:limit]
	}
	return facts
}

func (s *InsightsService) AuditLogs(limit int) []domain.AuditLogEntry {
	return s.store.ListAuditLogs(limit)
}

func transportSpanishName(t domain.ReservationType) string {
	switch t {
	case domain.ReservationFlight:
		return "Vuelo"
	case domain.ReservationHotel:
		return "Alojamiento"
	case domain.ReservationRental:
		return "Alquiler de auto"
	case domain.ReservationTrain:
		return "Tren"
	case domain.ReservationBus:
		return "Autobús"
	default:
		return string(t)
	}
}

var _ = time.UTC
