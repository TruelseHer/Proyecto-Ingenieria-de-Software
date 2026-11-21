package service

import (
	"fmt"
	"sync"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
)

type RecommendationStrategy interface {
	ID() string
	Label() string
	Description() string
	Generate(user domain.User, trips []*domain.Trip, trends []domain.DestinationTrend) []domain.Recommendation
}

type CollaborativeStrategy struct{}

func (c CollaborativeStrategy) ID() string    { return "collaborative" }
func (c CollaborativeStrategy) Label() string { return "Afinidad de viajeros" }
func (c CollaborativeStrategy) Description() string {
	return "Sugerencias basadas en viajeros con rutas e intereses similares a los tuyos."
}

func (c CollaborativeStrategy) Generate(user domain.User, trips []*domain.Trip, _ []domain.DestinationTrend) []domain.Recommendation {
	hasJapan := false
	hasSpain := false
	for _, t := range trips {
		if t.CountryCode == "JP" || t.DestinationCity == "Tokio" {
			hasJapan = true
		}
		if t.CountryCode == "ES" || t.DestinationCity == "Madrid" {
			hasSpain = true
		}
	}

	var recs []domain.Recommendation
	if hasJapan {
		recs = append(recs,
			domain.Recommendation{
				ID:             "rec-col-1",
				Strategy:       c.ID(),
				Title:          "Pase Regional Hakone + Excursión Monte Fuji",
				Destination:    "Hakone, Japón",
				Category:       "Experiencia ferroviaria",
				Description:    "El 84% de los viajeros con ruta Tokio–Kioto combinan un traslado en Romancecar hacia Hakone.",
				EstimatedPrice: 82.0,
				Currency:       user.HomeCurrency,
				Score:          97,
				HighlightTag:   "Ideal para tu ruta a Japón",
			},
			domain.Recommendation{
				ID:             "rec-col-2",
				Strategy:       c.ID(),
				Title:          "Lounge Haneda Terminal 3 + Fast Track",
				Destination:    "Tokio Haneda (HND)",
				Category:       "Confort en aeropuerto",
				Description:    "Acceso directo tras tu vuelo internacional para ducha, desayuno y conexión rápida al Monorail.",
				EstimatedPrice: 45.0,
				Currency:       user.HomeCurrency,
				Score:          92,
				HighlightTag:   "Ahorra 40 min en arribo",
			},
		)
	}
	if hasSpain {
		recs = append(recs,
			domain.Recommendation{
				ID:             "rec-col-3",
				Strategy:       c.ID(),
				Title:          "Parador de Toledo — Escapada Histórica",
				Destination:    "Toledo, España",
				Category:       "Alojamiento boutique",
				Description:    "Recomendado por viajeros que conectan Madrid y el sur de España en tren de alta velocidad.",
				EstimatedPrice: 165.0,
				Currency:       user.HomeCurrency,
				Score:          91,
				HighlightTag:   "Ruta panorámica",
			},
		)
	}
	if len(recs) == 0 {
		recs = append(recs,
			domain.Recommendation{
				ID:             "rec-col-4",
				Strategy:       c.ID(),
				Title:          "Japón Imperial: Tokio & Kioto en Shinkansen",
				Destination:    "Tokio y Kioto, Japón",
				Category:       "Paquete destacado",
				Description:    "Circuito favorito de la comunidad con vuelo directo, hotel en Shinjuku y tren bala.",
				EstimatedPrice: 1890.0,
				Currency:       user.HomeCurrency,
				Score:          98,
				HighlightTag:   "Más elegido",
			},
			domain.Recommendation{
				ID:             "rec-col-5",
				Strategy:       c.ID(),
				Title:          "Patagonia Andina: Bariloche & 7 Lagos",
				Destination:    "Bariloche, Argentina",
				Category:       "Naturaleza y Lagos",
				Description:    "Incluye vuelo, estadía en Llao Llao y vehículo SUV 4x4 para recorrer la cordillera.",
				EstimatedPrice: 920.0,
				Currency:       user.HomeCurrency,
				Score:          95,
				HighlightTag:   "Salidas confirmadas",
			},
		)
	}
	return recs
}

type TrendingStrategy struct{}

func (t TrendingStrategy) ID() string    { return "trending" }
func (t TrendingStrategy) Label() string { return "Tendencias de temporada" }
func (t TrendingStrategy) Description() string {
	return "Destinos y conexiones con mayor crecimiento de demanda y puntualidad verificada."
}

func (t TrendingStrategy) Generate(user domain.User, _ []*domain.Trip, trends []domain.DestinationTrend) []domain.Recommendation {
	recs := []domain.Recommendation{
		{
			ID:             "rec-trd-1",
			Strategy:       t.ID(),
			Title:          "Kioto Otoño — Ryokan Tradicional en Gion",
			Destination:    "Kioto, Japón",
			Category:       "Alojamiento destacado",
			Description:    "Demanda en alza (+38% esta temporada) con conexión directa vía Tokaido Shinkansen (99.1% puntualidad).",
			EstimatedPrice: 290.0,
			Currency:       user.HomeCurrency,
			Score:          98,
			HighlightTag:   "+38% demanda",
		},
		{
			ID:             "rec-trd-2",
			Strategy:       t.ID(),
			Title:          "Corredor Ferroviario Madrid–Barcelona AVE",
			Destination:    "Barcelona, España",
			Category:       "Alta velocidad",
			Description:    "Alternativa rápida de centro a centro con 98% de puntualidad verificada.",
			EstimatedPrice: 74.0,
			Currency:       user.HomeCurrency,
			Score:          94,
			HighlightTag:   "Puntualidad 98%",
		},
	}
	if len(trends) > 0 {
		top := trends[0]
		recs = append(recs, domain.Recommendation{
			ID:             "rec-trd-3",
			Strategy:       t.ID(),
			Title:          fmt.Sprintf("Explorar salidas a %s", top.City),
			Destination:    fmt.Sprintf("%s (%s)", top.City, top.CountryCode),
			Category:       "Destino líder",
			Description:    fmt.Sprintf("%s • Medio favorito: %s.", top.SeasonHighlight, top.TopTransport),
			EstimatedPrice: 145.0,
			Currency:       user.HomeCurrency,
			Score:          96,
			HighlightTag:   fmt.Sprintf("+%.0f%% reservas", top.GrowthPercent),
		})
	}
	return recs
}

type RecommendationResult struct {
	ActiveStrategy string
	Strategies     []RecommendationStrategy
	Items          []domain.Recommendation
	CacheHit       bool
	LatencyMs      int64
}

type RecommendationService struct {
	mu             sync.RWMutex
	store          *repository.Store
	cache          *platform.CacheClient
	insights       *InsightsService
	strategies     map[string]RecommendationStrategy
	strategyOrder  []string
	userStrategies map[string]string
}

func NewRecommendationService(store *repository.Store, cache *platform.CacheClient, insights *InsightsService) *RecommendationService {
	svc := &RecommendationService{
		store:          store,
		cache:          cache,
		insights:       insights,
		strategies:     make(map[string]RecommendationStrategy),
		userStrategies: make(map[string]string),
	}
	svc.RegisterStrategy(CollaborativeStrategy{})
	svc.RegisterStrategy(TrendingStrategy{})
	return svc
}

func (s *RecommendationService) RegisterStrategy(st RecommendationStrategy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.strategies[st.ID()]; !exists {
		s.strategyOrder = append(s.strategyOrder, st.ID())
	}
	s.strategies[st.ID()] = st
}

func (s *RecommendationService) SetUserStrategy(userID, strategyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.strategies[strategyID]; !ok {
		return fmt.Errorf("%w: estrategia no registrada", domain.ErrValidation)
	}
	s.userStrategies[userID] = strategyID
	return nil
}

func (s *RecommendationService) GetRecommendations(userID string) (RecommendationResult, error) {
	start := time.Now()
	user, err := s.store.GetUser(userID)
	if err != nil {
		return RecommendationResult{}, err
	}

	s.mu.RLock()
	activeID := s.userStrategies[userID]
	if activeID == "" {
		activeID = "collaborative"
	}
	st := s.strategies[activeID]
	var all []RecommendationStrategy
	for _, id := range s.strategyOrder {
		all = append(all, s.strategies[id])
	}
	s.mu.RUnlock()

	cacheKey := fmt.Sprintf("user:%s:recs:%s", userID, activeID)
	var cachedItems []domain.Recommendation
	if s.cache != nil && s.cache.GetJSON(cacheKey, &cachedItems) {
		return RecommendationResult{
			ActiveStrategy: activeID,
			Strategies:     all,
			Items:          cachedItems,
			CacheHit:       true,
			LatencyMs:      time.Since(start).Milliseconds(),
		}, nil
	}

	trips := s.store.ListTripsByUser(userID)
	var trends []domain.DestinationTrend
	if s.insights != nil {
		trends = s.insights.DestinationTrends()
	}
	items := st.Generate(user, trips, trends)
	if s.cache != nil {
		s.cache.SetJSON(cacheKey, items, 30*time.Second)
	}

	return RecommendationResult{
		ActiveStrategy: activeID,
		Strategies:     all,
		Items:          items,
		CacheHit:       false,
		LatencyMs:      time.Since(start).Milliseconds(),
	}, nil
}
