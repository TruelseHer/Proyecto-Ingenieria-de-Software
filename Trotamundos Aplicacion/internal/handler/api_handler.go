package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
	"trotamundos/internal/service"
)

type APIHandler struct {
	store       *repository.Store
	cache       *platform.CacheClient
	authSvc     *service.AuthService
	catalogSvc  *service.CatalogService
	tripSvc     *service.TripService
	syncSvc     *service.SyncService
	recSvc      *service.RecommendationService
	compSvc     *service.CompanionService
	insightsSvc *service.InsightsService
}

func NewAPIHandler(
	store *repository.Store,
	cache *platform.CacheClient,
	authSvc *service.AuthService,
	catalogSvc *service.CatalogService,
	tripSvc *service.TripService,
	syncSvc *service.SyncService,
	recSvc *service.RecommendationService,
	compSvc *service.CompanionService,
	insightsSvc *service.InsightsService,
) *APIHandler {
	return &APIHandler{
		store:       store,
		cache:       cache,
		authSvc:     authSvc,
		catalogSvc:  catalogSvc,
		tripSvc:     tripSvc,
		syncSvc:     syncSvc,
		recSvc:      recSvc,
		compSvc:     compSvc,
		insightsSvc: insightsSvc,
	}
}

func (a *APIHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", a.Health)

	// Catalog API
	mux.HandleFunc("GET /api/v1/catalog/packages", a.ListCatalogPackages)
	mux.HandleFunc("GET /api/v1/catalog/offers", a.ListCatalogOffers)

	// Trips & Reservations API
	mux.HandleFunc("GET /api/v1/trips/active", a.ListActiveTrips)
	mux.HandleFunc("GET /api/v1/trips/archived", a.ListArchivedTrips)
	mux.HandleFunc("GET /api/v1/trips/{tripId}", a.GetTrip)
	mux.HandleFunc("POST /api/v1/trips/{tripId}/restore", a.RestoreTrip)
	mux.HandleFunc("POST /api/v1/trips/{tripId}/reservations", a.CreateManualReservation)
	mux.HandleFunc("PUT /api/v1/reservations/{reservationId}", a.UpdateManualReservation)
	mux.HandleFunc("DELETE /api/v1/reservations/{reservationId}", a.DeleteReservation)
	mux.HandleFunc("POST /api/v1/reservations/{reservationId}/agency-support", a.AgencySupport)

	// Mobile Navigation Context
	mux.HandleFunc("GET /mobile/v1/trips/{tripId}/navigation-context", a.MobileNavigationContext)

	// Community Trends & Analytics
	mux.HandleFunc("GET /api/v1/analytics/trends/destinations", a.AnalyticsDestinationTrends)
	mux.HandleFunc("GET /api/v1/analytics/metrics/disruptions", a.AnalyticsDisruptions)
	mux.HandleFunc("GET /api/v1/analytics/monetization/summary", a.AnalyticsMonetizationSummary)

	// Recommendations & Human-in-the-Loop Optimization API
	mux.HandleFunc("GET /api/v1/recommendations", a.GetRecommendations)
	mux.HandleFunc("POST /api/v1/recommendations/parameters", a.UpdateRecommendationParameters)
	mux.HandleFunc("POST /api/v1/recommendations/feedback", a.RecordRecommendationFeedback)
}

func (a *APIHandler) Health(w http.ResponseWriter, _ *http.Request) {
	dbStatus := "UP"
	if err := a.store.Ping(); err != nil {
		dbStatus = "DEGRADED"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "UP",
		"storageMode": a.store.Mode(),
		"cacheMode":   a.cache.Mode(),
		"database":    dbStatus,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (a *APIHandler) ListCatalogPackages(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"packages": a.catalogSvc.ListPackages(),
	})
}

func (a *APIHandler) ListCatalogOffers(w http.ResponseWriter, r *http.Request) {
	typeFilter := r.URL.Query().Get("type")
	query := r.URL.Query().Get("q")
	writeJSON(w, http.StatusOK, map[string]any{
		"offers": a.catalogSvc.ListOffers(typeFilter, query),
	})
}

func (a *APIHandler) ListActiveTrips(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	trips, err := a.tripSvc.ListTrips(userID, "ACTIVE", "ALL")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"trips": trips})
}

func (a *APIHandler) ListArchivedTrips(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	trips, err := a.tripSvc.ListTrips(userID, "ARCHIVED", "ALL")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"trips": trips})
}

func (a *APIHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	tripID := r.PathValue("tripId")
	trip, err := a.tripSvc.GetTripForUser(userID, tripID, ClientIPFromContext(r.Context()))
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, trip)
}

func (a *APIHandler) RestoreTrip(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	tripID := r.PathValue("tripId")
	trip, err := a.tripSvc.RestoreTrip(userID, tripID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, trip)
}

func (a *APIHandler) CreateManualReservation(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	tripID := r.PathValue("tripId")
	var in struct {
		Type             string                    `json:"type"`
		ProviderName     string                    `json:"providerName"`
		ConfirmationCode string                    `json:"confirmationCode"`
		Origin           string                    `json:"origin"`
		Destination      string                    `json:"destination"`
		StartDate        time.Time                 `json:"startDate"`
		EndDate          time.Time                 `json:"endDate"`
		Price            float64                   `json:"price"`
		Currency         string                    `json:"currency"`
		Notes            string                    `json:"notes"`
		Details          domain.ReservationDetails `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	res := &domain.Reservation{
		Type:             domain.ReservationType(strings.ToUpper(in.Type)),
		Status:           domain.StatusConfirmed,
		ProviderName:     in.ProviderName,
		ConfirmationCode: in.ConfirmationCode,
		Origin:           in.Origin,
		Destination:      in.Destination,
		StartDate:        in.StartDate,
		EndDate:          in.EndDate,
		Price:            in.Price,
		Currency:         in.Currency,
		Notes:            in.Notes,
		Source:           domain.SourceManual,
		Details:          in.Details,
	}
	_, created, err := a.tripSvc.AddReservation(userID, tripID, res)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *APIHandler) UpdateManualReservation(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	resID := r.PathValue("reservationId")
	var in struct {
		ProviderName     *string                    `json:"providerName"`
		ConfirmationCode *string                    `json:"confirmationCode"`
		Origin           *string                    `json:"origin"`
		Destination      *string                    `json:"destination"`
		StartDate        *time.Time                 `json:"startDate"`
		EndDate          *time.Time                 `json:"endDate"`
		Status           *domain.ReservationStatus  `json:"status"`
		Price            *float64                   `json:"price"`
		Currency         *string                    `json:"currency"`
		Notes            *string                    `json:"notes"`
		Details          *domain.ReservationDetails `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	_, updated, err := a.tripSvc.UpdateReservation(userID, resID, domain.ReservationUpdateInput{
		ProviderName:     in.ProviderName,
		ConfirmationCode: in.ConfirmationCode,
		Origin:           in.Origin,
		Destination:      in.Destination,
		StartDate:        in.StartDate,
		EndDate:          in.EndDate,
		Status:           in.Status,
		Price:            in.Price,
		Currency:         in.Currency,
		Notes:            in.Notes,
		Details:          in.Details,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *APIHandler) DeleteReservation(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	resID := r.PathValue("reservationId")
	if _, err := a.tripSvc.DeleteReservation(userID, resID); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *APIHandler) AgencySupport(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	resID := r.PathValue("reservationId")
	var in struct {
		IssueSummary string `json:"issueSummary"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	ticket, _, err := a.tripSvc.RequestAgencySupport(userID, resID, in.IssueSummary)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, ticket)
}

func (a *APIHandler) MobileNavigationContext(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	tripID := r.PathValue("tripId")
	targetCurrency := r.URL.Query().Get("targetCurrency")
	route, err := a.compSvc.BuildNavigationRoute(userID, tripID, targetCurrency, "airport")
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tripId":              route.TripID,
		"destinationTimeZone": route.DestinationTimezone,
		"localTimeFormatted":  route.LocalTimeFormatted,
		"hotelLocation": map[string]any{
			"hotelName": route.HotelName,
			"address":   route.HotelAddress,
			"latitude":  route.HotelLat,
			"longitude": route.HotelLon,
		},
		"currencyContext": map[string]any{
			"localCurrency":    route.LocalCurrency,
			"targetCurrency":   route.TargetCurrency,
			"exchangeRate":     route.ExchangeRate,
			"rateSource":       route.RateSource,
			"sampleFareLocal":  route.SampleFareLocal,
			"sampleFareTarget": route.SampleFareTarget,
		},
		"weather":         route.Weather,
		"distanceKm":      route.DistanceKm,
		"transitMinutes":  route.TransitMinutes,
		"polylineCoords":  route.PolylineCoords,
		"mapRoutingToken": route.RoutingToken,
	})
}

func (a *APIHandler) AnalyticsDestinationTrends(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"trends": a.insightsSvc.DestinationTrends(),
	})
}

func (a *APIHandler) AnalyticsDisruptions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"disruptions": a.insightsSvc.ProviderDisruptions(),
	})
}

func (a *APIHandler) AnalyticsMonetizationSummary(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.insightsSvc.Summary())
}

func (a *APIHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	res, err := a.recSvc.GetRecommendations(userID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"activeStrategy": res.ActiveStrategy,
		"preferences":    res.Preferences,
		"profileSummary": res.ProfileSummary,
		"items":          res.Items,
		"cacheHit":       res.CacheHit,
		"latencyMs":      res.LatencyMs,
	})
}

func (a *APIHandler) UpdateRecommendationParameters(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	var in struct {
		StrategyID         string `json:"strategyId"`
		BudgetLevel        string `json:"budgetLevel"`
		PreferredTransport string `json:"preferredTransport"`
		FocusStyle         string `json:"focusStyle"`
		DestinationFocus   string `json:"destinationFocus"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	pref, err := a.recSvc.UpdateParameters(userID, in.StrategyID, in.BudgetLevel, in.PreferredTransport, in.FocusStyle, in.DestinationFocus)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, _ := a.recSvc.GetRecommendations(userID)
	writeJSON(w, http.StatusOK, map[string]any{
		"preferences": pref,
		"items":       res.Items,
	})
}

func (a *APIHandler) RecordRecommendationFeedback(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	var in struct {
		RecommendationID string `json:"recommendationId"`
		Action           string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSONError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	pref, target, err := a.recSvc.RecordFeedback(userID, in.RecommendationID, in.Action)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, _ := a.recSvc.GetRecommendations(userID)
	writeJSON(w, http.StatusOK, map[string]any{
		"preferences": pref,
		"target":      target,
		"items":       res.Items,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error":  message,
		"status": status,
	})
}
