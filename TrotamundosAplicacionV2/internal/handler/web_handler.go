package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
	"trotamundos/internal/service"
)

type WebHandler struct {
	store       *repository.Store
	tokens      *platform.TokenService
	renderer    *Renderer
	authSvc     *service.AuthService
	catalogSvc  *service.CatalogService
	tripSvc     *service.TripService
	syncSvc     *service.SyncService
	recSvc      *service.RecommendationService
	compSvc     *service.CompanionService
	insightsSvc *service.InsightsService
	extAPI      *platform.ExternalAPIClient
}

func NewWebHandler(
	store *repository.Store,
	tokens *platform.TokenService,
	renderer *Renderer,
	authSvc *service.AuthService,
	catalogSvc *service.CatalogService,
	tripSvc *service.TripService,
	syncSvc *service.SyncService,
	recSvc *service.RecommendationService,
	compSvc *service.CompanionService,
	insightsSvc *service.InsightsService,
	extAPI *platform.ExternalAPIClient,
) *WebHandler {
	if extAPI == nil {
		extAPI = platform.NewExternalAPIClient(nil, 1800*time.Millisecond)
	}
	return &WebHandler{
		store:       store,
		tokens:      tokens,
		renderer:    renderer,
		authSvc:     authSvc,
		catalogSvc:  catalogSvc,
		tripSvc:     tripSvc,
		syncSvc:     syncSvc,
		recSvc:      recSvc,
		compSvc:     compSvc,
		insightsSvc: insightsSvc,
		extAPI:      extAPI,
	}
}

func (h *WebHandler) Register(mux *http.ServeMux) {
	// Authentication & Profile
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.LoginSubmit)
	mux.HandleFunc("GET /register", h.RegisterPage)
	mux.HandleFunc("POST /register", h.RegisterSubmit)
	mux.HandleFunc("POST /logout", h.Logout)
	mux.HandleFunc("POST /profile/preferences", h.UpdatePreferences)

	// Main Pages
	mux.HandleFunc("GET /", h.DashboardPage)
	mux.HandleFunc("GET /welcome", h.WelcomePage)
	mux.HandleFunc("GET /trips", h.DashboardPage)
	mux.HandleFunc("GET /explore", h.ExplorePage)
	mux.HandleFunc("GET /explore/packages", h.ExplorePackagesPage)
	mux.HandleFunc("GET /inbox", h.InboxPage)
	mux.HandleFunc("GET /sync", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/inbox", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /companion", h.CompanionPage)
	mux.HandleFunc("GET /trends", h.TrendsPage)
	mux.HandleFunc("GET /insights", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/trends", http.StatusSeeOther)
	})

	// Explore & Preloaded Catalog Actions (HTMX)
	mux.HandleFunc("GET /partials/catalog/offers", h.PartialCatalogOffers)
	mux.HandleFunc("GET /partials/catalog/offers/{id}/book-modal", h.BookOfferModal)
	mux.HandleFunc("POST /catalog/packages/{id}/book", h.BookCatalogPackage)
	mux.HandleFunc("POST /catalog/packages/{id}/send-email", h.SendCatalogPackageEmail)
	mux.HandleFunc("POST /catalog/offers/{id}/book", h.BookCatalogOffer)
	mux.HandleFunc("POST /catalog/offers/{id}/send-email", h.SendCatalogOfferEmail)

	// Trips & Reservations (HTMX)
	mux.HandleFunc("GET /partials/trips", h.PartialTrips)
	mux.HandleFunc("POST /trips", h.CreateTrip)
	mux.HandleFunc("POST /trips/{id}/archive", h.ArchiveTrip)
	mux.HandleFunc("POST /trips/{id}/restore", h.RestoreTrip)
	mux.HandleFunc("POST /trips/{id}/share", h.ShareTrip)

	mux.HandleFunc("GET /partials/reservations/new", h.NewReservationModal)
	mux.HandleFunc("GET /partials/reservations/{id}/edit", h.EditReservationModal)
	mux.HandleFunc("POST /trips/{id}/reservations", h.CreateReservation)
	mux.HandleFunc("PUT /reservations/{id}", h.UpdateReservation)
	mux.HandleFunc("DELETE /reservations/{id}", h.DeleteReservation)
	mux.HandleFunc("POST /reservations/{id}/support", h.RequestSupport)

	// Recommendations
	mux.HandleFunc("GET /partials/recommendations", h.PartialRecommendations)
	mux.HandleFunc("POST /recommendations/strategy", h.SwitchRecommendationStrategy)

	// Inbox, PNR Import & Trusted Senders
	mux.HandleFunc("POST /inbox/import-pnr", h.ImportByPNR)
	mux.HandleFunc("POST /inbox/scan-email", h.ScanExternalEmail)
	mux.HandleFunc("POST /inbox/mailbox-auth", h.ToggleMailboxAuth)
	mux.HandleFunc("POST /inbox/sync-all", h.SyncAllInbox)
	mux.HandleFunc("POST /inbox/confirmations/{id}/import", h.ImportConfirmation)
	mux.HandleFunc("POST /inbox/senders", h.AddSenderRule)
	mux.HandleFunc("POST /inbox/senders/{id}/toggle", h.ToggleSenderRule)
	mux.HandleFunc("DELETE /inbox/senders/{id}", h.DeleteSenderRule)

	// Destination Companion
	mux.HandleFunc("GET /companion/route", h.CompanionRoutePartial)

	// Trends
	mux.HandleFunc("GET /partials/trends", h.TrendsPartial)
}

type baseViewData struct {
	IsAuthPage      bool
	ActiveNav       string
	PageTitle       string
	CurrentUser     domain.User
	Agencies        []domain.TravelAgency
	PreferredAgency domain.TravelAgency
}

func (h *WebHandler) buildBaseData(r *http.Request, activeNav, title string) baseViewData {
	userID := UserIDFromContext(r.Context())
	currentUser, _ := h.store.GetUser(userID)
	agencies := h.store.ListAgencies()
	preferred, err := h.store.GetAgency(currentUser.PreferredAgencyID)
	if err != nil && len(agencies) > 0 {
		preferred = agencies[0]
	}
	return baseViewData{
		IsAuthPage:      false,
		ActiveNav:       activeNav,
		PageTitle:       title,
		CurrentUser:     currentUser,
		Agencies:        agencies,
		PreferredAgency: preferred,
	}
}

// ---- Authentication ----

type authPageViewData struct {
	baseViewData
	Name     string
	Email    string
	Error    string
	Agencies []domain.TravelAgency
}

func (h *WebHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if UserIDFromContext(r.Context()) != "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	h.renderer.RenderPage(w, "login.html", authPageViewData{
		baseViewData: baseViewData{
			IsAuthPage: true,
			PageTitle:  "Iniciar sesión",
		},
	})
}

func (h *WebHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	_, tok, err := h.authSvc.Login(email, password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		h.renderer.RenderPage(w, "login.html", authPageViewData{
			baseViewData: baseViewData{
				IsAuthPage: true,
				PageTitle:  "Iniciar sesión",
			},
			Email: email,
			Error: err.Error(),
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
	http.Redirect(w, r, "/welcome", http.StatusSeeOther)
}

func (h *WebHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if UserIDFromContext(r.Context()) != "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	h.renderer.RenderPage(w, "register.html", authPageViewData{
		baseViewData: baseViewData{
			IsAuthPage: true,
			PageTitle:  "Crear cuenta",
		},
		Agencies: h.store.ListAgencies(),
	})
}

func (h *WebHandler) RegisterSubmit(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	currency := r.FormValue("homeCurrency")
	agencyID := r.FormValue("preferredAgencyId")

	_, tok, err := h.authSvc.Register(name, email, password, currency, agencyID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.renderer.RenderPage(w, "register.html", authPageViewData{
			baseViewData: baseViewData{
				IsAuthPage: true,
				PageTitle:  "Crear cuenta",
			},
			Name:     name,
			Email:    email,
			Error:    err.Error(),
			Agencies: h.store.ListAgencies(),
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
	http.Redirect(w, r, "/welcome", http.StatusSeeOther)
}

func (h *WebHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(SessionCookieName); err == nil && c.Value != "" {
		h.authSvc.Logout(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *WebHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	userID := UserIDFromContext(r.Context())
	agencyID := r.FormValue("agencyId")
	currency := r.FormValue("homeCurrency")
	if _, err := h.tripSvc.UpdateUserPreferences(userID, agencyID, currency); err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "Error al guardar preferencias",
			Message: err.Error(),
		})
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

// ---- Welcome & Usage Guide (`/welcome`) ----

func (h *WebHandler) WelcomePage(w http.ResponseWriter, r *http.Request) {
	base := h.buildBaseData(r, "welcome", "Cómo usar Trotamundos")
	h.renderer.RenderPage(w, "welcome.html", base)
}

// ---- Explore Catalog (`/explore` and `/explore/packages`) ----

type exploreViewData struct {
	baseViewData
	Packages   []domain.CatalogPackage
	Offers     []domain.CatalogOffer
	UserTrips  []*domain.Trip
	Query      string
	TypeFilter string
}

func (h *WebHandler) ExplorePage(w http.ResponseWriter, r *http.Request) {
	base := h.buildBaseData(r, "explore", "Buscar viajes disponibles")
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	typeFilter := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("type")))
	if typeFilter == "" {
		typeFilter = "ALL"
	}
	targetCurrency := base.CurrentUser.HomeCurrency
	offers := h.convertOffersForCurrency(h.catalogSvc.ListOffers(typeFilter, q), targetCurrency)
	userTrips, _ := h.tripSvc.ListTrips(base.CurrentUser.ID, "ACTIVE", "ALL")

	h.renderer.RenderPage(w, "explore.html", exploreViewData{
		baseViewData: base,
		Offers:       offers,
		UserTrips:    userTrips,
		Query:        q,
		TypeFilter:   typeFilter,
	})
}

func (h *WebHandler) ExplorePackagesPage(w http.ResponseWriter, r *http.Request) {
	base := h.buildBaseData(r, "packages", "Paquetes completos")
	targetCurrency := base.CurrentUser.HomeCurrency
	packages := h.convertPackagesForCurrency(h.catalogSvc.ListPackages(), targetCurrency)

	h.renderer.RenderPage(w, "explore_packages.html", exploreViewData{
		baseViewData: base,
		Packages:     packages,
	})
}

func (h *WebHandler) PartialCatalogOffers(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	currentUser, _ := h.store.GetUser(userID)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	typeFilter := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("type")))
	if typeFilter == "" {
		typeFilter = "ALL"
	}
	offers := h.convertOffersForCurrency(h.catalogSvc.ListOffers(typeFilter, q), currentUser.HomeCurrency)
	userTrips, _ := h.tripSvc.ListTrips(userID, "ACTIVE", "ALL")
	h.renderer.RenderPartial(w, "catalog_offers.html", map[string]any{
		"Offers":     offers,
		"UserTrips":  userTrips,
		"Query":      q,
		"TypeFilter": typeFilter,
	}, nil)
}

func (h *WebHandler) BookCatalogPackage(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	packageID := r.PathValue("id")
	dateID := r.FormValue("dateId")

	trip, count, err := h.catalogSvc.BookPackage(userID, packageID, dateID)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "No se pudo reservar el paquete",
			Message: err.Error(),
		})
		return
	}
	h.renderer.RenderToastOnly(w, http.StatusOK, ToastMessage{
		Type:    "success",
		Title:   "¡Paquete reservado con éxito!",
		Message: fmt.Sprintf("Se creó tu viaje \"%s\" con %d tramos confirmados y enviamos los comprobantes a tu bandeja.", trip.Title, count),
	})
}

func (h *WebHandler) SendCatalogPackageEmail(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	packageID := r.PathValue("id")
	dateID := r.FormValue("dateId")

	confs, err := h.catalogSvc.SendPackageToInbox(userID, packageID, dateID)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "No se pudo enviar al correo",
			Message: err.Error(),
		})
		return
	}
	h.renderer.RenderToastOnly(w, http.StatusOK, ToastMessage{
		Type:    "info",
		Title:   "Comprobantes enviados a tu correo",
		Message: fmt.Sprintf("Recibiste %d correos de confirmación en tu bandeja (/inbox) listos para sincronizar.", len(confs)),
	})
}

func (h *WebHandler) BookOfferModal(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	offerID := r.PathValue("id")
	offer, err := h.catalogSvc.GetOffer(offerID)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusNotFound, ToastMessage{
			Type:    "error",
			Title:   "Servicio no encontrado",
			Message: err.Error(),
		})
		return
	}
	currentUser, _ := h.store.GetUser(userID)
	converted := h.convertOffersForCurrency([]domain.CatalogOffer{offer}, currentUser.HomeCurrency)[0]
	userTrips, _ := h.tripSvc.ListTrips(userID, "ACTIVE", "ALL")

	defaultTripID := "NEW"
	if len(userTrips) > 0 {
		defaultTripID = userTrips[0].ID
		destLower := strings.ToLower(strings.TrimSpace(offer.DestinationCity))
		for _, t := range userTrips {
			if destLower != "" && (strings.Contains(strings.ToLower(t.DestinationCity), destLower) || strings.Contains(strings.ToLower(t.Title), destLower)) {
				defaultTripID = t.ID
				break
			}
		}
	}

	destCity := strings.TrimSpace(offer.DestinationCity)
	if destCity == "" {
		destCity = offer.Destination
	}

	h.renderer.RenderPartial(w, "book_offer_modal.html", map[string]any{
		"Offer":          &converted,
		"UserTrips":      userTrips,
		"DefaultTripID":  defaultTripID,
		"SuggestedTitle": fmt.Sprintf("Viaje a %s", destCity),
	}, nil)
}

func (h *WebHandler) BookCatalogOffer(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	offerID := r.PathValue("id")
	tripID := r.FormValue("tripId")
	newTripTitle := r.FormValue("newTripTitle")

	trip, res, err := h.catalogSvc.BookOfferWithOptions(userID, offerID, tripID, newTripTitle)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "No se pudo reservar el servicio",
			Message: err.Error(),
		})
		return
	}
	w.Header().Set("HX-Trigger", "closeModal")
	h.renderer.RenderToastOnly(w, http.StatusOK, ToastMessage{
		Type:    "success",
		Title:   "Servicio incorporado a tu viaje",
		Message: fmt.Sprintf("%s %s (PNR %s) agregado a \"%s\".", res.TypeLabel(), res.ProviderName, res.ConfirmationCode, trip.Title),
	})
}

func (h *WebHandler) SendCatalogOfferEmail(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	offerID := r.PathValue("id")

	conf, err := h.catalogSvc.SendOfferToInbox(userID, offerID)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "Error al enviar comprobante",
			Message: err.Error(),
		})
		return
	}
	h.renderer.RenderToastOnly(w, http.StatusOK, ToastMessage{
		Type:    "info",
		Title:   "Confirmación recibida en tu bandeja",
		Message: fmt.Sprintf("Comprobante PNR %s (%s) disponible en Confirmaciones y correo.", conf.PNR, conf.ProviderName),
	})
}

// ---- Dashboard / My Trips (`/`) ----

type dashboardViewData struct {
	baseViewData
	Trips           []*domain.Trip
	StatusFilter    string
	TypeFilter      string
	Recommendations service.RecommendationResult
}

func (h *WebHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/trips" {
		http.NotFound(w, r)
		return
	}
	base := h.buildBaseData(r, "trips", "Mis Viajes")
	statusFilter := r.URL.Query().Get("status")
	if statusFilter == "" {
		statusFilter = "ACTIVE"
	}
	typeFilter := r.URL.Query().Get("type")
	if typeFilter == "" {
		typeFilter = "ALL"
	}

	trips, _ := h.tripSvc.ListTrips(base.CurrentUser.ID, statusFilter, typeFilter)
	trips = h.convertTripsForCurrency(trips, base.CurrentUser.HomeCurrency)
	recs, _ := h.recSvc.GetRecommendations(base.CurrentUser.ID)
	recs = h.convertRecommendationsForCurrency(recs, base.CurrentUser.HomeCurrency)

	h.renderer.RenderPage(w, "dashboard.html", dashboardViewData{
		baseViewData:    base,
		Trips:           trips,
		StatusFilter:    statusFilter,
		TypeFilter:      typeFilter,
		Recommendations: recs,
	})
}

func (h *WebHandler) PartialTrips(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	statusFilter := r.URL.Query().Get("status")
	if statusFilter == "" {
		statusFilter = "ACTIVE"
	}
	typeFilter := r.URL.Query().Get("type")
	if typeFilter == "" {
		typeFilter = "ALL"
	}
	currentUser, _ := h.store.GetUser(userID)
	trips, _ := h.tripSvc.ListTrips(userID, statusFilter, typeFilter)
	trips = h.convertTripsForCurrency(trips, currentUser.HomeCurrency)
	h.renderer.RenderPartial(w, "trip_list.html", map[string]any{
		"Trips":        trips,
		"StatusFilter": statusFilter,
		"TypeFilter":   typeFilter,
		"CurrentUser":  currentUser,
	}, nil)
}

func (h *WebHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	userID := UserIDFromContext(r.Context())
	title := r.FormValue("title")
	city := r.FormValue("destinationCity")
	country := r.FormValue("countryCode")

	trip, err := h.tripSvc.CreateTrip(userID, title, city, country)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "No se pudo crear el viaje",
			Message: err.Error(),
		})
		return
	}

	currentUser, _ := h.store.GetUser(userID)
	trips, _ := h.tripSvc.ListTrips(userID, "ACTIVE", "ALL")
	trips = h.convertTripsForCurrency(trips, currentUser.HomeCurrency)
	h.renderer.RenderPartial(w, "trip_list.html", map[string]any{
		"Trips":        trips,
		"StatusFilter": "ACTIVE",
		"TypeFilter":   "ALL",
		"CurrentUser":  currentUser,
	}, &ToastMessage{
		Type:    "success",
		Title:   "Viaje creado",
		Message: fmt.Sprintf("Se creó \"%s\" con destino a %s.", trip.Title, trip.DestinationCity),
	})
}

func (h *WebHandler) ArchiveTrip(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	tripID := r.PathValue("id")
	trip, err := h.tripSvc.ArchiveTrip(userID, tripID)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusForbidden, ToastMessage{
			Type:    "error",
			Title:   "Acción denegada",
			Message: err.Error(),
		})
		return
	}
	statusFilter := r.FormValue("status")
	if statusFilter == "" {
		statusFilter = "ACTIVE"
	}
	typeFilter := r.FormValue("type")
	if typeFilter == "" {
		typeFilter = "ALL"
	}
	currentUser, _ := h.store.GetUser(userID)
	trips, _ := h.tripSvc.ListTrips(userID, statusFilter, typeFilter)
	trips = h.convertTripsForCurrency(trips, currentUser.HomeCurrency)
	h.renderer.RenderPartial(w, "trip_list.html", map[string]any{
		"Trips":        trips,
		"StatusFilter": statusFilter,
		"TypeFilter":   typeFilter,
		"CurrentUser":  currentUser,
	}, &ToastMessage{
		Type:    "info",
		Title:   "Viaje archivado",
		Message: fmt.Sprintf("\"%s\" pasó a tu historial de viajes archivados.", trip.Title),
	})
}

func (h *WebHandler) RestoreTrip(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	tripID := r.PathValue("id")
	trip, err := h.tripSvc.RestoreTrip(userID, tripID)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusForbidden, ToastMessage{
			Type:    "error",
			Title:   "No se pudo restaurar",
			Message: err.Error(),
		})
		return
	}
	statusFilter := r.FormValue("status")
	if statusFilter == "" {
		statusFilter = "ARCHIVED"
	}
	typeFilter := r.FormValue("type")
	if typeFilter == "" {
		typeFilter = "ALL"
	}
	currentUser, _ := h.store.GetUser(userID)
	trips, _ := h.tripSvc.ListTrips(userID, statusFilter, typeFilter)
	trips = h.convertTripsForCurrency(trips, currentUser.HomeCurrency)
	h.renderer.RenderPartial(w, "trip_list.html", map[string]any{
		"Trips":        trips,
		"StatusFilter": statusFilter,
		"TypeFilter":   typeFilter,
		"CurrentUser":  currentUser,
	}, &ToastMessage{
		Type:    "success",
		Title:   "Viaje restaurado",
		Message: fmt.Sprintf("\"%s\" volvió a tus viajes activos.", trip.Title),
	})
}

func (h *WebHandler) ShareTrip(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	userID := UserIDFromContext(r.Context())
	tripID := r.PathValue("id")
	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		email = strings.TrimSpace(r.FormValue("targetUserId"))
	}

	trip, err := h.tripSvc.ShareTrip(userID, tripID, email)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusForbidden, ToastMessage{
			Type:    "error",
			Title:   "No se pudo compartir el viaje",
			Message: err.Error(),
		})
		return
	}
	currentUser, _ := h.store.GetUser(userID)
	trip = h.convertTripForCurrency(trip, currentUser.HomeCurrency)
	h.renderer.RenderPartial(w, "trip_card.html", map[string]any{
		"Trip":        trip,
		"CurrentUser": currentUser,
	}, &ToastMessage{
		Type:    "success",
		Title:   "Itinerario compartido",
		Message: fmt.Sprintf("Invitación enviada a %s para consultar \"%s\".", email, trip.Title),
	})
}

func (h *WebHandler) NewReservationModal(w http.ResponseWriter, r *http.Request) {
	tripID := r.URL.Query().Get("tripId")
	resType := domain.ReservationType(strings.ToUpper(r.URL.Query().Get("type")))
	if resType == "" {
		resType = domain.ReservationFlight
	}
	userID := UserIDFromContext(r.Context())
	user, _ := h.store.GetUser(userID)
	currency := user.HomeCurrency
	if currency == "" {
		currency = "USD"
	}
	start := time.Now().Add(72 * time.Hour)
	res := &domain.Reservation{
		TripID:    tripID,
		Type:      resType,
		Status:    domain.StatusConfirmed,
		StartDate: start,
		EndDate:   start.Add(4 * time.Hour),
		Currency:  currency,
	}
	h.renderer.RenderPartial(w, "reservation_modal.html", map[string]any{
		"IsEdit":      false,
		"TripID":      tripID,
		"Reservation": res,
	}, nil)
}

func (h *WebHandler) EditReservationModal(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	resID := r.PathValue("id")
	trip, res, err := h.store.FindReservation(resID)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusNotFound, ToastMessage{
			Type:    "error",
			Title:   "Reserva no encontrada",
			Message: err.Error(),
		})
		return
	}
	user, _ := h.store.GetUser(userID)
	if !trip.CanAccessByUser(user) {
		h.renderer.RenderToastOnly(w, http.StatusForbidden, ToastMessage{
			Type:    "error",
			Title:   "Acceso denegado",
			Message: "No tienes permisos para editar esta reserva.",
		})
		return
	}
	h.renderer.RenderPartial(w, "reservation_modal.html", map[string]any{
		"IsEdit":      true,
		"TripID":      trip.ID,
		"Reservation": res,
	}, nil)
}

func (h *WebHandler) CreateReservation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	userID := UserIDFromContext(r.Context())
	tripID := r.PathValue("id")

	res := parseReservationFromForm(r)
	res.Source = domain.SourceManual
	trip, created, err := h.tripSvc.AddReservation(userID, tripID, res)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "No se pudo agregar la reserva",
			Message: err.Error(),
		})
		return
	}

	currentUser, _ := h.store.GetUser(userID)
	trip = h.convertTripForCurrency(trip, currentUser.HomeCurrency)
	w.Header().Set("HX-Trigger", "closeModal")
	h.renderer.RenderPartial(w, "trip_card.html", map[string]any{
		"Trip":        trip,
		"CurrentUser": currentUser,
	}, &ToastMessage{
		Type:    "success",
		Title:   "Reserva agregada",
		Message: fmt.Sprintf("%s %s (%s) incorporada al itinerario.", created.TypeLabel(), created.ProviderName, created.ConfirmationCode),
	})
}

func (h *WebHandler) UpdateReservation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	userID := UserIDFromContext(r.Context())
	resID := r.PathValue("id")
	parsed := parseReservationFromForm(r)

	patch := domain.ReservationUpdateInput{
		ProviderName:     &parsed.ProviderName,
		ConfirmationCode: &parsed.ConfirmationCode,
		Origin:           &parsed.Origin,
		Destination:      &parsed.Destination,
		StartDate:        &parsed.StartDate,
		EndDate:          &parsed.EndDate,
		Status:           &parsed.Status,
		Price:            &parsed.Price,
		Currency:         &parsed.Currency,
		Notes:            &parsed.Notes,
		Details:          &parsed.Details,
	}
	trip, updated, err := h.tripSvc.UpdateReservation(userID, resID, patch)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "Error al actualizar",
			Message: err.Error(),
		})
		return
	}

	currentUser, _ := h.store.GetUser(userID)
	trip = h.convertTripForCurrency(trip, currentUser.HomeCurrency)
	w.Header().Set("HX-Trigger", "closeModal")
	h.renderer.RenderPartial(w, "trip_card.html", map[string]any{
		"Trip":        trip,
		"CurrentUser": currentUser,
	}, &ToastMessage{
		Type:    "success",
		Title:   "Reserva actualizada",
		Message: fmt.Sprintf("Los cambios en %s fueron guardados en tu itinerario.", updated.ConfirmationCode),
	})
}

func (h *WebHandler) DeleteReservation(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	resID := r.PathValue("id")
	trip, err := h.tripSvc.DeleteReservation(userID, resID)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "No se pudo eliminar",
			Message: err.Error(),
		})
		return
	}
	currentUser, _ := h.store.GetUser(userID)
	trip = h.convertTripForCurrency(trip, currentUser.HomeCurrency)
	h.renderer.RenderPartial(w, "trip_card.html", map[string]any{
		"Trip":        trip,
		"CurrentUser": currentUser,
	}, &ToastMessage{
		Type:    "info",
		Title:   "Reserva eliminada",
		Message: "El tramo fue eliminado y las fechas del viaje se recalcularon.",
	})
}

func (h *WebHandler) RequestSupport(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	resID := r.PathValue("id")
	issue := r.FormValue("issueSummary")
	ticket, trip, err := h.tripSvc.RequestAgencySupport(userID, resID, issue)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "Error al solicitar asistencia",
			Message: err.Error(),
		})
		return
	}
	currentUser, _ := h.store.GetUser(userID)
	trip = h.convertTripForCurrency(trip, currentUser.HomeCurrency)
	h.renderer.RenderPartial(w, "trip_card.html", map[string]any{
		"Trip":        trip,
		"CurrentUser": currentUser,
	}, &ToastMessage{
		Type:    "success",
		Title:   "Solicitud enviada a tu agencia",
		Message: fmt.Sprintf("Caso %s abierto con %s.", ticket.ReferenceCode, ticket.AgencyName),
	})
}

func (h *WebHandler) PartialRecommendations(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	currentUser, _ := h.store.GetUser(userID)
	recs, _ := h.recSvc.GetRecommendations(userID)
	recs = h.convertRecommendationsForCurrency(recs, currentUser.HomeCurrency)
	h.renderer.RenderPartial(w, "recommendations.html", recs, nil)
}

func (h *WebHandler) SwitchRecommendationStrategy(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	strategyID := r.FormValue("strategy")
	if err := h.recSvc.SetUserStrategy(userID, strategyID); err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{
			Type:    "error",
			Title:   "Modo no disponible",
			Message: err.Error(),
		})
		return
	}
	currentUser, _ := h.store.GetUser(userID)
	recs, _ := h.recSvc.GetRecommendations(userID)
	recs = h.convertRecommendationsForCurrency(recs, currentUser.HomeCurrency)
	h.renderer.RenderPartial(w, "recommendations.html", recs, &ToastMessage{
		Type:    "info",
		Title:   "Recomendaciones actualizadas",
		Message: "Se actualizaron las experiencias sugeridas según tu preferencia.",
	})
}

// ---- Inbox, PNR Import & Trusted Senders (`/inbox`) ----

type inboxViewData struct {
	baseViewData
	Rules         []domain.SenderRule
	Confirmations []domain.EmailConfirmation
	SampleOffers  []domain.CatalogOffer
}

func (h *WebHandler) buildInboxData(r *http.Request) inboxViewData {
	base := h.buildBaseData(r, "inbox", "Confirmaciones y Correo")
	offers := h.catalogSvc.ListOffers("ALL", "")
	if len(offers) > 8 {
		offers = offers[:8]
	}
	offers = h.convertOffersForCurrency(offers, base.CurrentUser.HomeCurrency)
	confs := h.convertConfirmationsForCurrency(h.syncSvc.ListConfirmations(base.CurrentUser.ID), base.CurrentUser.HomeCurrency)
	return inboxViewData{
		baseViewData:  base,
		Rules:         h.syncSvc.ListSenderRules(base.CurrentUser.ID),
		Confirmations: confs,
		SampleOffers:  offers,
	}
}

func (h *WebHandler) InboxPage(w http.ResponseWriter, r *http.Request) {
	h.renderer.RenderPage(w, "inbox.html", h.buildInboxData(r))
}

func (h *WebHandler) ImportByPNR(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	pnr := r.FormValue("pnr")

	trip, res, err := h.syncSvc.ImportByPNR(userID, pnr)
	data := h.buildInboxData(r)
	if err != nil {
		h.renderer.RenderPartial(w, "inbox_panel.html", data, &ToastMessage{
			Type:    "error",
			Title:   "No se pudo importar el código PNR",
			Message: err.Error(),
		})
		return
	}
	h.renderer.RenderPartial(w, "inbox_panel.html", data, &ToastMessage{
		Type:    "success",
		Title:   "Reserva importada por código PNR",
		Message: fmt.Sprintf("%s %s (%s) incorporada al viaje \"%s\".", res.TypeLabel(), res.ProviderName, res.ConfirmationCode, trip.Title),
	})
}

func (h *WebHandler) ScanExternalEmail(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	fromEmail := r.FormValue("fromEmail")
	subject := r.FormValue("subject")
	body := r.FormValue("body")

	res, err := h.syncSvc.ScanEmail(userID, fromEmail, subject, body)
	data := h.buildInboxData(r)
	if err != nil {
		h.renderer.RenderPartial(w, "inbox_panel.html", data, &ToastMessage{
			Type:    "warning",
			Title:   "Correo no importado",
			Message: err.Error(),
		})
		return
	}
	h.renderer.RenderPartial(w, "inbox_panel.html", data, &ToastMessage{
		Type:    "success",
		Title:   "Reserva detectada e importada",
		Message: res.Message,
	})
}

func (h *WebHandler) ToggleMailboxAuth(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	u, _ := h.store.GetUser(userID)
	newVal := !u.MailboxAuthorized
	if v := r.FormValue("authorized"); v != "" {
		newVal = v == "true"
	}
	updated, err := h.syncSvc.ToggleMailboxAuth(userID, newVal)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusBadRequest, ToastMessage{Type: "error", Title: "Error", Message: err.Error()})
		return
	}
	msg := "Lectura de confirmaciones en pausa."
	if updated.MailboxAuthorized {
		msg = fmt.Sprintf("Lectura autorizada para sincronizar confirmaciones en %s.", updated.ConnectedEmail)
	}
	h.renderer.RenderPartial(w, "inbox_panel.html", h.buildInboxData(r), &ToastMessage{
		Type:    "info",
		Title:   "Permisos de correo actualizados",
		Message: msg,
	})
}

func (h *WebHandler) SyncAllInbox(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	res, err := h.syncSvc.SyncAllInbox(userID)
	data := h.buildInboxData(r)
	if err != nil {
		h.renderer.RenderPartial(w, "inbox_panel.html", data, &ToastMessage{
			Type:    "error",
			Title:   "Sincronización pausada",
			Message: err.Error(),
		})
		return
	}
	toastType := "info"
	if res.ImportedCount > 0 {
		toastType = "success"
	}
	h.renderer.RenderPartial(w, "inbox_panel.html", data, &ToastMessage{
		Type:    toastType,
		Title:   "Bandeja sincronizada",
		Message: res.Message,
	})
}

func (h *WebHandler) ImportConfirmation(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	confID := r.PathValue("id")
	trip, res, err := h.syncSvc.ImportConfirmation(userID, confID)
	data := h.buildInboxData(r)
	if err != nil {
		h.renderer.RenderPartial(w, "inbox_panel.html", data, &ToastMessage{
			Type:    "warning",
			Title:   "No se pudo sincronizar el comprobante",
			Message: err.Error(),
		})
		return
	}
	h.renderer.RenderPartial(w, "inbox_panel.html", data, &ToastMessage{
		Type:    "success",
		Title:   "Comprobante sincronizado",
		Message: fmt.Sprintf("%s %s (%s) sumado a \"%s\".", res.TypeLabel(), res.ProviderName, res.ConfirmationCode, trip.Title),
	})
}

func (h *WebHandler) AddSenderRule(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	userID := UserIDFromContext(r.Context())
	rule, err := h.syncSvc.AddSenderRule(userID, r.FormValue("pattern"), r.FormValue("providerLabel"), r.FormValue("category"))
	if err != nil {
		h.renderer.RenderPartial(w, "inbox_panel.html", h.buildInboxData(r), &ToastMessage{
			Type:    "error",
			Title:   "Remitente no válido",
			Message: err.Error(),
		})
		return
	}
	h.renderer.RenderPartial(w, "inbox_panel.html", h.buildInboxData(r), &ToastMessage{
		Type:    "success",
		Title:   "Dominio autorizado",
		Message: fmt.Sprintf("%s (%s) agregado a tu lista de remitentes confiables.", rule.ProviderLabel, rule.Pattern),
	})
}

func (h *WebHandler) ToggleSenderRule(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	ruleID := r.PathValue("id")
	rule, err := h.syncSvc.ToggleSenderRule(userID, ruleID)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusNotFound, ToastMessage{Type: "error", Title: "No encontrado", Message: err.Error()})
		return
	}
	stateText := "pausado"
	if rule.Enabled {
		stateText = "activo"
	}
	h.renderer.RenderPartial(w, "inbox_panel.html", h.buildInboxData(r), &ToastMessage{
		Type:    "info",
		Title:   "Filtro actualizado",
		Message: fmt.Sprintf("Dominio %s %s.", rule.Pattern, stateText),
	})
}

func (h *WebHandler) DeleteSenderRule(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	ruleID := r.PathValue("id")
	if err := h.syncSvc.DeleteSenderRule(userID, ruleID); err != nil {
		h.renderer.RenderToastOnly(w, http.StatusNotFound, ToastMessage{Type: "error", Title: "No encontrado", Message: err.Error()})
		return
	}
	h.renderer.RenderPartial(w, "inbox_panel.html", h.buildInboxData(r), &ToastMessage{
		Type:    "info",
		Title:   "Remitente eliminado",
		Message: "Se quitó el dominio de tu lista de remitentes confiables.",
	})
}

// ---- Destination Companion (`/companion`) ----

type companionViewData struct {
	baseViewData
	Trips          []*domain.Trip
	SelectedTripID string
	OriginMode     string
	TargetCurrency string
	Route          *domain.NavigationRoute
}

func (h *WebHandler) CompanionPage(w http.ResponseWriter, r *http.Request) {
	base := h.buildBaseData(r, "companion", "Guía en Destino")
	trips, _ := h.tripSvc.ListTrips(base.CurrentUser.ID, "ACTIVE", "ALL")

	selectedTripID := r.URL.Query().Get("tripId")
	if selectedTripID == "" && len(trips) > 0 {
		selectedTripID = trips[0].ID
	}
	originMode := r.URL.Query().Get("origin")
	if originMode == "" {
		originMode = "airport"
	}
	targetCurrency := r.URL.Query().Get("currency")
	if targetCurrency == "" {
		targetCurrency = base.CurrentUser.HomeCurrency
	}

	var route *domain.NavigationRoute
	if selectedTripID != "" {
		route, _ = h.compSvc.BuildNavigationRoute(base.CurrentUser.ID, selectedTripID, targetCurrency, originMode)
	}

	h.renderer.RenderPage(w, "companion.html", companionViewData{
		baseViewData:   base,
		Trips:          trips,
		SelectedTripID: selectedTripID,
		OriginMode:     originMode,
		TargetCurrency: targetCurrency,
		Route:          route,
	})
}

func (h *WebHandler) CompanionRoutePartial(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	tripID := r.URL.Query().Get("tripId")
	originMode := r.URL.Query().Get("origin")
	if originMode == "" {
		originMode = "airport"
	}
	currency := r.URL.Query().Get("currency")
	route, err := h.compSvc.BuildNavigationRoute(userID, tripID, currency, originMode)
	if err != nil {
		h.renderer.RenderToastOnly(w, http.StatusForbidden, ToastMessage{
			Type:    "error",
			Title:   "Ruta no disponible",
			Message: err.Error(),
		})
		return
	}
	h.renderer.RenderPartial(w, "companion_route.html", map[string]any{
		"Route":          route,
		"OriginMode":     originMode,
		"TargetCurrency": route.TargetCurrency,
	}, nil)
}

// ---- Community Trends (`/trends`) ----

type trendsViewData struct {
	baseViewData
	Snapshot service.InsightsSnapshot
}

func (h *WebHandler) TrendsPage(w http.ResponseWriter, r *http.Request) {
	base := h.buildBaseData(r, "trends", "Tendencias de Destinos")
	h.renderer.RenderPage(w, "trends.html", trendsViewData{
		baseViewData: base,
		Snapshot:     h.insightsSvc.Snapshot(),
	})
}

func (h *WebHandler) TrendsPartial(w http.ResponseWriter, r *http.Request) {
	h.renderer.RenderPartial(w, "insights_panel.html", map[string]any{
		"Snapshot": h.insightsSvc.Snapshot(),
	}, &ToastMessage{
		Type:    "info",
		Title:   "Indicadores actualizados",
		Message: "Tendencias de destinos e índices de puntualidad sincronizados.",
	})
}

// ---- Form Helpers ----

func parseReservationFromForm(r *http.Request) *domain.Reservation {
	resType := domain.ReservationType(strings.ToUpper(strings.TrimSpace(r.FormValue("type"))))
	if resType == "" {
		resType = domain.ReservationFlight
	}
	status := domain.ReservationStatus(strings.ToUpper(strings.TrimSpace(r.FormValue("status"))))
	if status == "" {
		status = domain.StatusConfirmed
	}
	price, _ := strconv.ParseFloat(strings.TrimSpace(r.FormValue("price")), 64)
	start := parseDateTimeLocal(r.FormValue("startDate"), time.Now().Add(72*time.Hour))
	end := parseDateTimeLocal(r.FormValue("endDate"), start.Add(4*time.Hour))
	lat, _ := strconv.ParseFloat(strings.TrimSpace(r.FormValue("latitude")), 64)
	lon, _ := strconv.ParseFloat(strings.TrimSpace(r.FormValue("longitude")), 64)

	details := domain.ReservationDetails{
		FlightNumber:      strings.TrimSpace(r.FormValue("flightNumber")),
		DepartureAirport:  strings.ToUpper(strings.TrimSpace(r.FormValue("departureAirport"))),
		ArrivalAirport:    strings.ToUpper(strings.TrimSpace(r.FormValue("arrivalAirport"))),
		Terminal:          strings.TrimSpace(r.FormValue("terminal")),
		Gate:              strings.TrimSpace(r.FormValue("gate")),
		CabinClass:        strings.TrimSpace(r.FormValue("cabinClass")),
		HotelName:         strings.TrimSpace(r.FormValue("hotelName")),
		Address:           strings.TrimSpace(r.FormValue("address")),
		Latitude:          lat,
		Longitude:         lon,
		RoomType:          strings.TrimSpace(r.FormValue("roomType")),
		CheckInTime:       strings.TrimSpace(r.FormValue("checkInTime")),
		CheckOutTime:      strings.TrimSpace(r.FormValue("checkOutTime")),
		VehicleCategory:   strings.TrimSpace(r.FormValue("vehicleCategory")),
		PickupLocation:    strings.TrimSpace(r.FormValue("pickupLocation")),
		DropoffLocation:   strings.TrimSpace(r.FormValue("dropoffLocation")),
		InsuranceIncluded: r.FormValue("insuranceIncluded") == "on" || r.FormValue("insuranceIncluded") == "true",
		TrainLine:         strings.TrimSpace(r.FormValue("trainLine")),
		TrainNumber:       strings.TrimSpace(r.FormValue("trainNumber")),
		DepartureStation:  strings.TrimSpace(r.FormValue("departureStation")),
		ArrivalStation:    strings.TrimSpace(r.FormValue("arrivalStation")),
		Wagon:             strings.TrimSpace(r.FormValue("wagon")),
		BusCompany:        strings.TrimSpace(r.FormValue("busCompany")),
		LineNumber:        strings.TrimSpace(r.FormValue("lineNumber")),
		DepartureTerminal: strings.TrimSpace(r.FormValue("departureTerminal")),
		ArrivalTerminal:   strings.TrimSpace(r.FormValue("arrivalTerminal")),
		Platform:          strings.TrimSpace(r.FormValue("platform")),
		Seat:              strings.TrimSpace(r.FormValue("seat")),
	}

	return &domain.Reservation{
		Type:             resType,
		Status:           status,
		ProviderName:     strings.TrimSpace(r.FormValue("providerName")),
		ConfirmationCode: strings.ToUpper(strings.TrimSpace(r.FormValue("confirmationCode"))),
		Origin:           strings.TrimSpace(r.FormValue("origin")),
		Destination:      strings.TrimSpace(r.FormValue("destination")),
		StartDate:        start,
		EndDate:          end,
		Price:            price,
		Currency:         strings.ToUpper(strings.TrimSpace(r.FormValue("currency"))),
		Notes:            strings.TrimSpace(r.FormValue("notes")),
		Details:          details,
	}
}

func parseDateTimeLocal(val string, fallback time.Time) time.Time {
	val = strings.TrimSpace(val)
	if val == "" {
		return fallback.UTC()
	}
	layouts := []string{
		"2006-01-02T15:04",
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, val); err == nil {
			return t.UTC()
		}
	}
	return fallback.UTC()
}

// ---- Dynamic Currency Conversion Helpers ----

func (h *WebHandler) normalizeCurrency(currency string) string {
	c := strings.ToUpper(strings.TrimSpace(currency))
	if c == "" {
		return "USD"
	}
	return c
}

func (h *WebHandler) convertAmount(amount float64, fromCurrency, targetCurrency string) float64 {
	from := h.normalizeCurrency(fromCurrency)
	target := h.normalizeCurrency(targetCurrency)
	if from == target || amount == 0 {
		return math.Round(amount*100) / 100
	}
	rate, _ := h.extAPI.GetExchangeRate(from, target)
	if rate <= 0 {
		return math.Round(amount*100) / 100
	}
	return math.Round((amount*rate)*100) / 100
}

func (h *WebHandler) convertPackagesForCurrency(pkgs []domain.CatalogPackage, targetCurrency string) []domain.CatalogPackage {
	target := h.normalizeCurrency(targetCurrency)
	out := make([]domain.CatalogPackage, len(pkgs))
	for i, pkg := range pkgs {
		cp := pkg
		cp.PricePerPerson = h.convertAmount(pkg.PricePerPerson, pkg.Currency, target)
		cp.Currency = target
		out[i] = cp
	}
	return out
}

func (h *WebHandler) convertOffersForCurrency(offers []domain.CatalogOffer, targetCurrency string) []domain.CatalogOffer {
	target := h.normalizeCurrency(targetCurrency)
	out := make([]domain.CatalogOffer, len(offers))
	for i, off := range offers {
		cp := off
		cp.Price = h.convertAmount(off.Price, off.Currency, target)
		cp.Currency = target
		out[i] = cp
	}
	return out
}

func (h *WebHandler) convertTripForCurrency(t *domain.Trip, targetCurrency string) *domain.Trip {
	if t == nil {
		return nil
	}
	target := h.normalizeCurrency(targetCurrency)
	cp := *t
	cp.Reservations = make([]*domain.Reservation, len(t.Reservations))
	for i, r := range t.Reservations {
		rcp := *r
		rcp.Price = h.convertAmount(r.Price, r.Currency, target)
		rcp.Currency = target
		cp.Reservations[i] = &rcp
	}
	return &cp
}

func (h *WebHandler) convertTripsForCurrency(trips []*domain.Trip, targetCurrency string) []*domain.Trip {
	out := make([]*domain.Trip, len(trips))
	for i, t := range trips {
		out[i] = h.convertTripForCurrency(t, targetCurrency)
	}
	return out
}

func (h *WebHandler) convertRecommendationsForCurrency(recs service.RecommendationResult, targetCurrency string) service.RecommendationResult {
	target := h.normalizeCurrency(targetCurrency)
	items := make([]domain.Recommendation, len(recs.Items))
	for i, item := range recs.Items {
		cp := item
		// Recommendations store base estimates in USD
		cp.EstimatedPrice = h.convertAmount(item.EstimatedPrice, "USD", target)
		cp.Currency = target
		items[i] = cp
	}
	recs.Items = items
	return recs
}

func (h *WebHandler) convertConfirmationsForCurrency(confs []domain.EmailConfirmation, targetCurrency string) []domain.EmailConfirmation {
	target := h.normalizeCurrency(targetCurrency)
	out := make([]domain.EmailConfirmation, len(confs))
	for i, c := range confs {
		cp := c
		cp.Price = h.convertAmount(c.Price, c.Currency, target)
		cp.Currency = target
		out[i] = cp
	}
	return out
}
