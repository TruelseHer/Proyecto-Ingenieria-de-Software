package service

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
)

// RecommendationStrategy formalizes the plugin interface defined in Spec-Driven Development
// (Iteration 4 - Section C: RecommendationStrategy) and ADR-15.
type RecommendationStrategy interface {
	ID() string
	GetStrategyIdentifier() string
	Label() string
	Description() string
	GenerateRecommendations(ctx domain.UserTripContext) []domain.Recommendation
}

// CollaborativeStrategy ("Afinidad de viajeros") analyzes the user's trip history,
// booked destinations, transport habits, and human-in-the-loop preference weights
// to suggest complementary lodging, transport, and route alternatives.
type CollaborativeStrategy struct{}

func (c CollaborativeStrategy) ID() string                    { return "collaborative" }
func (c CollaborativeStrategy) GetStrategyIdentifier() string { return "collaborative" }
func (c CollaborativeStrategy) Label() string                 { return "Afinidad de viajeros" }
func (c CollaborativeStrategy) Description() string {
	return "Analiza tu historial de viajes, medios de transporte frecuentes y afinidad con otros viajeros para sugerir tramos, hoteles y rutas complementarias."
}

func (c CollaborativeStrategy) GenerateRecommendations(ctx domain.UserTripContext) []domain.Recommendation {
	candidates := buildCandidateCatalog(ctx, c.ID())
	var results []domain.Recommendation

	for _, item := range candidates {
		score := item.Score
		reasons := []string{}

		// 1. History & Route Continuity analysis
		if ctx.BookedCities[item.DestinationCity] || ctx.BookedCountries[item.CountryCode] {
			score += 14
			reasons = append(reasons, fmt.Sprintf("Complementa tu itinerario en %s", item.DestinationCity))
		}
		if ctx.TopUserTransport != "" && item.TransportType == ctx.TopUserTransport {
			score += 6
			reasons = append(reasons, fmt.Sprintf("Afinidad con tu medio frecuente (%s)", transportLabelES(item.TransportType)))
		}
		if ctx.HistoricalAvgCost > 0 {
			if item.EstimatedPrice <= ctx.HistoricalAvgCost*1.15 {
				score += 4
			}
		}

		// 2. Human-in-the-loop learned weights & parameter tuning
		score, reasons = applyHumanInTheLoopRefinement(item, ctx, score, reasons)
		if score < 0 {
			continue
		}

		item.Score = clampScore(score)
		if len(reasons) > 0 {
			item.MatchReason = strings.Join(reasons, " · ")
		} else if item.MatchReason == "" {
			item.MatchReason = "Sugerido por afinidad de rutas entre viajeros de la comunidad"
		}
		results = append(results, item)
	}

	sortRecommendations(results)
	return results
}

// TrendingStrategy ("Tendencias y puntualidad") combines seasonal demand growth
// and verified operator punctuality from InsightsService with the user's parameter adjustments.
type TrendingStrategy struct{}

func (t TrendingStrategy) ID() string                    { return "trending" }
func (t TrendingStrategy) GetStrategyIdentifier() string { return "trending" }
func (t TrendingStrategy) Label() string                 { return "Tendencias y puntualidad" }
func (t TrendingStrategy) Description() string {
	return "Prioriza alternativas de alojamiento, trenes, vuelos y rutas con mayor crecimiento de demanda y mejor índice de puntualidad verificado."
}

func (t TrendingStrategy) GenerateRecommendations(ctx domain.UserTripContext) []domain.Recommendation {
	candidates := buildCandidateCatalog(ctx, t.ID())

	trendGrowth := make(map[string]float64)
	for _, tr := range ctx.Trends {
		trendGrowth[tr.City] = tr.GrowthPercent
	}

	var results []domain.Recommendation
	for _, item := range candidates {
		score := item.Score
		reasons := []string{}

		if g, ok := trendGrowth[item.DestinationCity]; ok && g > 0 {
			boost := int(math.Round(g / 3.0))
			score += boost
			reasons = append(reasons, fmt.Sprintf("Demanda en alza (+%.0f%% esta temporada)", g))
		}
		if item.TransportType == domain.ReservationTrain {
			score += 7
			reasons = append(reasons, "Corredor ferroviario con 98%+ de puntualidad")
		}
		if ctx.BookedCities[item.DestinationCity] {
			score += 6
			reasons = append(reasons, fmt.Sprintf("Conexión activa para tu viaje a %s", item.DestinationCity))
		}

		score, reasons = applyHumanInTheLoopRefinement(item, ctx, score, reasons)
		if score < 0 {
			continue
		}

		item.Score = clampScore(score)
		if len(reasons) > 0 {
			item.MatchReason = strings.Join(reasons, " · ")
		} else if item.MatchReason == "" {
			item.MatchReason = "Tendencia destacada con alta puntualidad operativa"
		}
		results = append(results, item)
	}

	sortRecommendations(results)
	return results
}

// SmartRouteOptimizerStrategy ("Optimización de ruta y presupuesto") focuses on multimodal
// alternatives (high-speed rail vs flight, boutique lodging, ground connections) tailored to user parameters.
type SmartRouteOptimizerStrategy struct{}

func (s SmartRouteOptimizerStrategy) ID() string                    { return "route_optimizer" }
func (s SmartRouteOptimizerStrategy) GetStrategyIdentifier() string { return "route_optimizer" }
func (s SmartRouteOptimizerStrategy) Label() string                 { return "Optimización multimodal" }
func (s SmartRouteOptimizerStrategy) Description() string {
	return "Sugiere alternativas proactivas de transporte, alojamiento y conexiones terrestres optimizando relación costo-tiempo y continuidad de ruta."
}

func (s SmartRouteOptimizerStrategy) GenerateRecommendations(ctx domain.UserTripContext) []domain.Recommendation {
	candidates := buildCandidateCatalog(ctx, s.ID())
	var results []domain.Recommendation

	for _, item := range candidates {
		score := item.Score
		reasons := []string{}

		// Favor multimodal connections (train, rental, bus, hotel) that complete a trip
		switch item.TransportType {
		case domain.ReservationTrain, domain.ReservationRental, domain.ReservationBus:
			score += 10
			reasons = append(reasons, "Optimiza traslados en destino sin esperas de aeropuerto")
		case domain.ReservationHotel:
			score += 8
			reasons = append(reasons, "Ubicación estratégica cerca de nodos de transporte")
		}

		if item.EstimatedPrice <= 350 {
			score += 7
			reasons = append(reasons, "Excelente relación costo-beneficio para tu ruta")
		}
		if ctx.BookedCities[item.DestinationCity] {
			score += 11
			reasons = append(reasons, fmt.Sprintf("Conecta directamente con tu estadía en %s", item.DestinationCity))
		}

		score, reasons = applyHumanInTheLoopRefinement(item, ctx, score, reasons)
		if score < 0 {
			continue
		}

		item.Score = clampScore(score)
		if len(reasons) > 0 {
			item.MatchReason = strings.Join(reasons, " · ")
		} else if item.MatchReason == "" {
			item.MatchReason = "Alternativa multimodal optimizada en tiempo y presupuesto"
		}
		results = append(results, item)
	}

	sortRecommendations(results)
	return results
}

func applyHumanInTheLoopRefinement(item domain.Recommendation, ctx domain.UserTripContext, baseScore int, reasons []string) (int, []string) {
	prefs := ctx.Preferences
	score := baseScore

	// Skip dismissed recommendations unless user filters specifically
	for _, disID := range prefs.DismissedIDs {
		if disID == item.ID {
			return -1, nil
		}
	}

	// Destination focus parameter
	if prefs.DestinationFocus != "" && prefs.DestinationFocus != "ALL" {
		if !strings.EqualFold(item.DestinationCity, prefs.DestinationFocus) {
			return -1, nil
		}
		score += 12
		reasons = append(reasons, fmt.Sprintf("Filtro activo: %s", item.DestinationCity))
	}

	// Preferred transport / category parameter
	if prefs.PreferredTransport != "" && prefs.PreferredTransport != "ALL" {
		if !strings.EqualFold(string(item.TransportType), prefs.PreferredTransport) {
			return -1, nil
		}
		score += 10
		reasons = append(reasons, fmt.Sprintf("Preferencia: %s", transportLabelES(item.TransportType)))
	}

	// Budget level parameter (evaluated in USD base price)
	switch strings.ToUpper(prefs.BudgetLevel) {
	case "ECONOMY":
		if item.EstimatedPrice > 320 {
			return -1, nil
		}
		score += 8
		reasons = append(reasons, "Ajustado a presupuesto económico")
	case "COMFORT":
		if item.EstimatedPrice > 1150 {
			return -1, nil
		}
		if item.EstimatedPrice >= 120 && item.EstimatedPrice <= 1150 {
			score += 6
			reasons = append(reasons, "Rango confort equilibrado")
		}
	case "PREMIUM":
		if item.EstimatedPrice >= 700 {
			score += 10
			reasons = append(reasons, "Experiencia gama alta / confort superior")
		}
	}

	// Focus style parameter
	switch strings.ToUpper(prefs.FocusStyle) {
	case "ROUTE_CONTINUITY":
		if ctx.BookedCities[item.DestinationCity] || ctx.BookedCountries[item.CountryCode] {
			score += 12
		}
	case "PUNCTUALITY":
		if item.TransportType == domain.ReservationTrain || item.TransportType == domain.ReservationFlight {
			score += 9
		}
	case "VALUE":
		if item.EstimatedPrice <= 300 {
			score += 10
		}
	}

	// Continuous learning weights from Human-in-the-Loop interactions (Accept / Dismiss history)
	if w, ok := prefs.DestinationWeights[item.DestinationCity]; ok && w != 0 {
		score += w * 7
		if w > 0 {
			reasons = append(reasons, fmt.Sprintf("Afinidad aprendida por tus elecciones en %s (+%d%%)", item.DestinationCity, w*7))
		} else {
			reasons = append(reasons, fmt.Sprintf("Ajustado por descartes previos en %s", item.DestinationCity))
		}
	}
	if w, ok := prefs.CategoryWeights[string(item.TransportType)]; ok && w != 0 {
		score += w * 6
		if w > 0 {
			reasons = append(reasons, fmt.Sprintf("Priorizado por tu preferencia en %s", strings.ToLower(transportLabelES(item.TransportType))))
		}
	}

	return score, reasons
}

func buildCandidateCatalog(ctx domain.UserTripContext, strategyID string) []domain.Recommendation {
	var list []domain.Recommendation

	// Build recommendations from all 18 catalog offers (flights, hotels, high-speed trains, buses, car rentals)
	for _, off := range ctx.CatalogOffers {
		baseScore := int(math.Round(off.Rating * 17.5))
		if baseScore < 78 {
			baseScore = 82
		}
		priceUSD := off.Price
		if strings.EqualFold(off.Currency, "EUR") {
			priceUSD = math.Round(off.Price*1.08*100) / 100
		}

		categoryLabel := transportCategoryLabel(off.Type)
		rec := domain.Recommendation{
			ID:              "rec-off-" + off.ID,
			Strategy:        strategyID,
			Title:           off.Title,
			Destination:     fmt.Sprintf("%s (%s)", off.DestinationCity, off.CountryCode),
			DestinationCity: off.DestinationCity,
			CountryCode:     off.CountryCode,
			Category:        categoryLabel,
			TransportType:   off.Type,
			Description:     fmt.Sprintf("%s · Ruta: %s → %s (%s).", off.Subtitle, off.Origin, off.Destination, off.ProviderName),
			EstimatedPrice:  priceUSD,
			Currency:        "USD",
			Score:           baseScore,
			HighlightTag:    off.HighlightTag,
			ImageURL:        off.ImageURL,
			LinkedOfferID:   off.ID,
		}
		list = append(list, rec)
	}

	// Build curated route & lodging alternatives across all 6 destinations
	curated := []domain.Recommendation{
		{
			ID:              "rec-cur-hakone",
			Strategy:        strategyID,
			Title:           "Pase Regional Hakone + Tren Romancecar al Monte Fuji",
			Destination:     "Tokio (JP)",
			DestinationCity: "Tokio",
			CountryCode:     "JP",
			Category:        "Ruta ferroviaria alternativa",
			TransportType:   domain.ReservationTrain,
			Description:     "El 84% de los viajeros con ruta Tokio–Kioto suman esta conexión panorámica en tren hacia Hakone y el Lago Ashi.",
			EstimatedPrice:  88.0,
			Currency:        "USD",
			Score:           89,
			HighlightTag:    "Ruta complementaria Japón",
			ImageURL:        "https://images.unsplash.com/photo-1493976040374-85c8e12f0c0e?auto=format&fit=crop&w=900&q=80",
			LinkedOfferID:   "off-train-jr-9914",
		},
		{
			ID:              "rec-cur-toledo",
			Strategy:        strategyID,
			Title:           "Parador de Toledo — Alojamiento Histórico con Vista al Tajo",
			Destination:     "Madrid (ES)",
			DestinationCity: "Madrid",
			CountryCode:     "ES",
			Category:        "Alternativa de alojamiento",
			TransportType:   domain.ReservationHotel,
			Description:     "Alternativa boutique a 33 minutos en tren Avant desde Madrid Puerta de Atocha, elegida por viajeros culturales.",
			EstimatedPrice:  195.0,
			Currency:        "USD",
			Score:           88,
			HighlightTag:    "Escapada desde Madrid",
			ImageURL:        "https://images.unsplash.com/photo-1558642452-9d2a7deb7f62?auto=format&fit=crop&w=900&q=80",
			LinkedOfferID:   "off-hotel-bkg-9012",
		},
		{
			ID:              "rec-cur-florencia",
			Strategy:        strategyID,
			Title:           "Conexión Alta Velocidad Frecciarossa Roma → Florencia",
			Destination:     "Roma (IT)",
			DestinationCity: "Roma",
			CountryCode:     "IT",
			Category:        "Ruta ferroviaria rápida",
			TransportType:   domain.ReservationTrain,
			Description:     "Conecta Roma Termini con Florencia Santa Maria Novella en 1h 32m con máxima puntualidad y sin esperas de embarque.",
			EstimatedPrice:  96.0,
			Currency:        "USD",
			Score:           89,
			HighlightTag:    "Corredor Italia 300 km/h",
			ImageURL:        "https://images.unsplash.com/photo-1534445867742-43195f401b6c?auto=format&fit=crop&w=900&q=80",
			LinkedOfferID:   "off-rental-avs-4190",
		},
		{
			ID:              "rec-cur-vallesagrado",
			Strategy:        strategyID,
			Title:           "Lodge Boutique en Valle Sagrado + Conexión Ollantaytambo",
			Destination:     "Cusco (PE)",
			DestinationCity: "Cusco",
			CountryCode:     "PE",
			Category:        "Alojamiento de aclimatación",
			TransportType:   domain.ReservationHotel,
			Description:     "Recomendado para aclimatarse a menor altitud antes de abordar el tren panorámico hacia Machu Picchu.",
			EstimatedPrice:  240.0,
			Currency:        "USD",
			Score:           90,
			HighlightTag:    "Ideal aclimatación andina",
			ImageURL:        "https://images.unsplash.com/photo-1526392060635-9d6019884377?auto=format&fit=crop&w=900&q=80",
			LinkedOfferID:   "off-train-pru-8045",
		},
	}
	list = append(list, curated...)

	// Mark feedback status on items
	acceptedSet := make(map[string]bool, len(ctx.Preferences.AcceptedIDs))
	for _, id := range ctx.Preferences.AcceptedIDs {
		acceptedSet[id] = true
	}
	for i := range list {
		if acceptedSet[list[i].ID] {
			list[i].FeedbackStatus = "ACCEPTED"
		}
	}

	return list
}

func clampScore(score int) int {
	if score > 99 {
		return 99
	}
	if score < 55 {
		return 55
	}
	return score
}

func sortRecommendations(items []domain.Recommendation) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score != items[j].Score {
			return items[i].Score > items[j].Score
		}
		return items[i].EstimatedPrice < items[j].EstimatedPrice
	})
}

func transportCategoryLabel(t domain.ReservationType) string {
	switch t {
	case domain.ReservationFlight:
		return "Alternativa de vuelo"
	case domain.ReservationHotel:
		return "Alternativa de alojamiento"
	case domain.ReservationTrain:
		return "Tren de alta velocidad / Ruta"
	case domain.ReservationBus:
		return "Conexión terrestre en autobús"
	case domain.ReservationRental:
		return "Movilidad en auto / Ruta escénica"
	default:
		return "Experiencia sugerida"
	}
}

func transportLabelES(t domain.ReservationType) string {
	switch t {
	case domain.ReservationFlight:
		return "Vuelos"
	case domain.ReservationHotel:
		return "Alojamientos"
	case domain.ReservationTrain:
		return "Trenes"
	case domain.ReservationBus:
		return "Autobuses"
	case domain.ReservationRental:
		return "Autos de alquiler"
	default:
		return "Multimodal"
	}
}

type RecommendationResult struct {
	ActiveStrategy      string
	ActiveStrategyLabel string
	ActiveStrategyDesc  string
	Strategies          []RecommendationStrategy
	Preferences         domain.RecommendationPreference
	RecentFeedback      []domain.RecommendationFeedbackEntry
	ProfileSummary      UserHistoryProfileSummary
	Items               []domain.Recommendation
	AvailableCities     []string
	CacheHit            bool
	LatencyMs           int64
}

type UserHistoryProfileSummary struct {
	TotalTrips          int
	TotalSegments       int
	BookedCities        []string
	TopTransportLabel   string
	HistoricalAvgPrice  float64
	LearnedAffinities   []string
	FeedbackInteractions int
}

type RecommendationService struct {
	mu            sync.RWMutex
	store         *repository.Store
	cache         *platform.CacheClient
	insights      *InsightsService
	strategies    map[string]RecommendationStrategy
	strategyOrder []string
}

func NewRecommendationService(store *repository.Store, cache *platform.CacheClient, insights *InsightsService) *RecommendationService {
	svc := &RecommendationService{
		store:      store,
		cache:      cache,
		insights:   insights,
		strategies: make(map[string]RecommendationStrategy),
	}
	svc.RegisterStrategy(CollaborativeStrategy{})
	svc.RegisterStrategy(TrendingStrategy{})
	svc.RegisterStrategy(SmartRouteOptimizerStrategy{})
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

func (s *RecommendationService) invalidateCache(userID string) {
	if s.cache == nil {
		return
	}
	s.cache.DeletePrefix(fmt.Sprintf("user:%s:recs:", userID))
}

func (s *RecommendationService) SetUserStrategy(userID, strategyID string) error {
	s.mu.RLock()
	_, ok := s.strategies[strategyID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%w: estrategia de recomendación no registrada", domain.ErrValidation)
	}
	pref := s.store.GetRecommendationPreference(userID)
	pref.StrategyID = strategyID
	pref.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveRecommendationPreference(pref); err != nil {
		return err
	}
	s.invalidateCache(userID)
	return nil
}

// UpdateParameters implements the Human-in-the-Loop parameter tuning mechanism,
// allowing the traveler to adjust strategy, budget level, transport type, destination, and optimization focus.
func (s *RecommendationService) UpdateParameters(userID, strategyID, budgetLevel, preferredTransport, focusStyle, destinationFocus string) (domain.RecommendationPreference, error) {
	pref := s.store.GetRecommendationPreference(userID)

	if strategyID != "" {
		s.mu.RLock()
		_, ok := s.strategies[strategyID]
		s.mu.RUnlock()
		if !ok {
			return pref, fmt.Errorf("%w: estrategia no registrada", domain.ErrValidation)
		}
		pref.StrategyID = strategyID
	}

	budgetLevel = strings.ToUpper(strings.TrimSpace(budgetLevel))
	switch budgetLevel {
	case "ALL", "ECONOMY", "COMFORT", "PREMIUM":
		pref.BudgetLevel = budgetLevel
	case "":
		// keep current
	default:
		pref.BudgetLevel = "ALL"
	}

	preferredTransport = strings.ToUpper(strings.TrimSpace(preferredTransport))
	switch preferredTransport {
	case "ALL", "FLIGHT", "HOTEL", "TRAIN", "BUS", "RENTAL":
		pref.PreferredTransport = preferredTransport
	case "":
		// keep current
	default:
		pref.PreferredTransport = "ALL"
	}

	focusStyle = strings.ToUpper(strings.TrimSpace(focusStyle))
	switch focusStyle {
	case "BALANCED", "ROUTE_CONTINUITY", "PUNCTUALITY", "VALUE":
		pref.FocusStyle = focusStyle
	case "":
		// keep current
	default:
		pref.FocusStyle = "BALANCED"
	}

	destinationFocus = strings.TrimSpace(destinationFocus)
	if destinationFocus == "" {
		destinationFocus = "ALL"
	}
	pref.DestinationFocus = destinationFocus
	pref.UpdatedAt = time.Now().UTC()

	if err := s.store.SaveRecommendationPreference(pref); err != nil {
		return pref, err
	}
	s.invalidateCache(userID)
	return pref, nil
}

// RecordFeedback implements the Human-in-the-Loop feedback loop where the traveler
// accepts/upvotes or dismisses a recommendation, continuously refining destination and category weights.
func (s *RecommendationService) RecordFeedback(userID, recommendationID, action string) (domain.RecommendationPreference, *domain.Recommendation, error) {
	pref := s.store.GetRecommendationPreference(userID)
	action = strings.ToUpper(strings.TrimSpace(action))

	if action == "RESET" {
		resetPref := repository.DefaultRecommendationPreference(userID)
		resetPref.StrategyID = pref.StrategyID
		if err := s.store.SaveRecommendationPreference(resetPref); err != nil {
			return resetPref, nil, err
		}
		_ = s.store.AppendRecommendationFeedback(domain.RecommendationFeedbackEntry{
			ID:               platform.NewID("rfb"),
			UserID:           userID,
			RecommendationID: "all",
			Title:            "Reinicio de parámetros y aprendizaje",
			Action:           "RESET",
			CreatedAt:        time.Now().UTC(),
		})
		s.invalidateCache(userID)
		return resetPref, nil, nil
	}

	// Locate target recommendation across candidates
	ctx, err := s.buildUserTripContext(userID)
	if err != nil {
		return pref, nil, err
	}
	candidates := buildCandidateCatalog(ctx, pref.StrategyID)
	var target *domain.Recommendation
	for i := range candidates {
		if candidates[i].ID == recommendationID {
			cp := candidates[i]
			target = &cp
			break
		}
	}
	if target == nil {
		return pref, nil, domain.ErrNotFound
	}

	switch action {
	case "ACCEPT", "LIKE":
		if !containsStr(pref.AcceptedIDs, recommendationID) {
			pref.AcceptedIDs = append(pref.AcceptedIDs, recommendationID)
		}
		pref.DismissedIDs = removeStr(pref.DismissedIDs, recommendationID)
		if target.DestinationCity != "" {
			pref.DestinationWeights[target.DestinationCity] += 2
		}
		if target.TransportType != "" {
			pref.CategoryWeights[string(target.TransportType)] += 2
		}
		pref.FeedbackCount++
	case "DISMISS", "DISLIKE":
		if !containsStr(pref.DismissedIDs, recommendationID) {
			pref.DismissedIDs = append(pref.DismissedIDs, recommendationID)
		}
		pref.AcceptedIDs = removeStr(pref.AcceptedIDs, recommendationID)
		if target.DestinationCity != "" {
			pref.DestinationWeights[target.DestinationCity] -= 1
		}
		if target.TransportType != "" {
			pref.CategoryWeights[string(target.TransportType)] -= 1
		}
		pref.FeedbackCount++
	default:
		return pref, nil, fmt.Errorf("%w: acción de feedback inválida", domain.ErrValidation)
	}

	pref.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveRecommendationPreference(pref); err != nil {
		return pref, nil, err
	}

	_ = s.store.AppendRecommendationFeedback(domain.RecommendationFeedbackEntry{
		ID:               platform.NewID("rfb"),
		UserID:           userID,
		RecommendationID: target.ID,
		Title:            target.Title,
		Action:           action,
		DestinationCity:  target.DestinationCity,
		TransportType:    string(target.TransportType),
		CreatedAt:        time.Now().UTC(),
	})

	s.invalidateCache(userID)
	return pref, target, nil
}

func (s *RecommendationService) buildUserTripContext(userID string) (domain.UserTripContext, error) {
	user, err := s.store.GetUser(userID)
	if err != nil {
		return domain.UserTripContext{}, err
	}
	pref := s.store.GetRecommendationPreference(userID)
	if pref.StrategyID == "" {
		pref.StrategyID = "collaborative"
	}

	trips := s.store.ListTripsByUser(userID)
	var trends []domain.DestinationTrend
	var disruptions []domain.ProviderDisruption
	if s.insights != nil {
		trends = s.insights.DestinationTrends()
		disruptions = s.insights.ProviderDisruptions()
	}

	bookedCities := make(map[string]bool)
	bookedCountries := make(map[string]bool)
	bookedPNRs := make(map[string]bool)
	transportCounts := make(map[domain.ReservationType]int)
	var totalCost float64
	var totalSegments int

	for _, tr := range trips {
		if tr.DestinationCity != "" {
			bookedCities[tr.DestinationCity] = true
		}
		if tr.CountryCode != "" {
			bookedCountries[tr.CountryCode] = true
		}
		for _, r := range tr.Reservations {
			totalSegments++
			totalCost += r.Price
			transportCounts[r.Type]++
			if r.ConfirmationCode != "" {
				bookedPNRs[r.ConfirmationCode] = true
			}
		}
	}

	var topTransport domain.ReservationType
	topCount := 0
	for tt, cnt := range transportCounts {
		if cnt > topCount {
			topCount = cnt
			topTransport = tt
		}
	}

	avgCost := 0.0
	if totalSegments > 0 {
		avgCost = math.Round((totalCost/float64(totalSegments))*100) / 100
	}

	return domain.UserTripContext{
		User:              user,
		Trips:             trips,
		Trends:            trends,
		Disruptions:       disruptions,
		CatalogOffers:     s.store.ListCatalogOffers("ALL", ""),
		CatalogPackages:   s.store.ListCatalogPackages(),
		Preferences:       pref,
		BookedCities:      bookedCities,
		BookedCountries:   bookedCountries,
		BookedOfferPNRs:   bookedPNRs,
		HistoricalAvgCost: avgCost,
		TopUserTransport:  topTransport,
	}, nil
}

func (s *RecommendationService) GetRecommendations(userID string) (RecommendationResult, error) {
	start := time.Now()
	ctx, err := s.buildUserTripContext(userID)
	if err != nil {
		return RecommendationResult{}, err
	}

	s.mu.RLock()
	activeID := ctx.Preferences.StrategyID
	if _, ok := s.strategies[activeID]; !ok {
		activeID = "collaborative"
	}
	st := s.strategies[activeID]
	var all []RecommendationStrategy
	for _, id := range s.strategyOrder {
		all = append(all, s.strategies[id])
	}
	s.mu.RUnlock()

	cacheKey := fmt.Sprintf(
		"user:%s:recs:%s:%s:%s:%s:%s:%d:%d",
		userID,
		activeID,
		ctx.Preferences.BudgetLevel,
		ctx.Preferences.PreferredTransport,
		ctx.Preferences.FocusStyle,
		ctx.Preferences.DestinationFocus,
		ctx.Preferences.FeedbackCount,
		len(ctx.Trips),
	)

	var items []domain.Recommendation
	cacheHit := false
	if s.cache != nil && s.cache.GetJSON(cacheKey, &items) {
		cacheHit = true
	} else {
		items = st.GenerateRecommendations(ctx)
		if len(items) > 9 {
			items = items[:9]
		}
		if s.cache != nil {
			s.cache.SetJSON(cacheKey, items, 30*time.Second)
		}
	}

	var bookedCityList []string
	for c := range ctx.BookedCities {
		bookedCityList = append(bookedCityList, c)
	}
	sort.Strings(bookedCityList)

	var affinities []string
	for city, w := range ctx.Preferences.DestinationWeights {
		if w > 0 {
			affinities = append(affinities, fmt.Sprintf("%s (+%d)", city, w))
		}
	}
	for cat, w := range ctx.Preferences.CategoryWeights {
		if w > 0 {
			affinities = append(affinities, fmt.Sprintf("%s (+%d)", transportLabelES(domain.ReservationType(cat)), w))
		}
	}
	sort.Strings(affinities)

	topTransLabel := "Sin historial previo"
	if ctx.TopUserTransport != "" {
		topTransLabel = transportLabelES(ctx.TopUserTransport)
	}

	totalSegs := 0
	for _, tr := range ctx.Trips {
		totalSegs += len(tr.Reservations)
	}

	availableCities := []string{"Tokio", "Madrid", "Bariloche", "Roma", "Cusco", "París"}

	return RecommendationResult{
		ActiveStrategy:      activeID,
		ActiveStrategyLabel: st.Label(),
		ActiveStrategyDesc:  st.Description(),
		Strategies:          all,
		Preferences:         ctx.Preferences,
		RecentFeedback:      s.store.ListRecommendationFeedback(userID, 6),
		ProfileSummary: UserHistoryProfileSummary{
			TotalTrips:           len(ctx.Trips),
			TotalSegments:        totalSegs,
			BookedCities:         bookedCityList,
			TopTransportLabel:    topTransLabel,
			HistoricalAvgPrice:   ctx.HistoricalAvgCost,
			LearnedAffinities:    affinities,
			FeedbackInteractions: ctx.Preferences.FeedbackCount,
		},
		Items:           items,
		AvailableCities: availableCities,
		CacheHit:        cacheHit,
		LatencyMs:       time.Since(start).Milliseconds(),
	}, nil
}

func containsStr(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

func removeStr(list []string, target string) []string {
	var out []string
	for _, s := range list {
		if s != target {
			out = append(out, s)
		}
	}
	return out
}
