package handler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/handler"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
	"trotamundos/internal/service"
)

type testApp struct {
	server      *httptest.Server
	store       *repository.Store
	tokens      *platform.TokenService
	authSvc     *service.AuthService
	catalogSvc  *service.CatalogService
	tripSvc     *service.TripService
	syncSvc     *service.SyncService
	compSvc     *service.CompanionService
	insightsSvc *service.InsightsService
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	store := repository.NewMemoryStore()
	cache := platform.NewCacheClient("")
	tokens := platform.NewTokenService("test-secret-2026", cache)
	limiter := platform.NewRateLimiter(500)
	bus := platform.NewEventBroker()
	t.Cleanup(func() { bus.Close() })
	guard := platform.NewProviderCircuitBreaker()
	extAPI := platform.NewExternalAPIClient(cache, 500*time.Millisecond)

	authSvc := service.NewAuthService(store, tokens)
	tripSvc := service.NewTripService(store, cache, bus)
	catalogSvc := service.NewCatalogService(store, tripSvc)
	syncSvc := service.NewSyncService(store, bus, guard, tripSvc)
	insightsSvc := service.NewInsightsService(store, bus, "salt-2026")
	recSvc := service.NewRecommendationService(store, cache, insightsSvc)
	compSvc := service.NewCompanionService(store, cache, tokens, extAPI)

	templatesDir := filepath.Join("..", "..", "web", "templates")
	renderer, err := handler.NewRenderer(templatesDir)
	if err != nil {
		t.Fatalf("failed to load templates: %v", err)
	}

	mux := http.NewServeMux()
	webH := handler.NewWebHandler(store, tokens, renderer, authSvc, catalogSvc, tripSvc, syncSvc, recSvc, compSvc, insightsSvc, extAPI)
	webH.Register(mux)
	apiH := handler.NewAPIHandler(store, cache, authSvc, catalogSvc, tripSvc, syncSvc, recSvc, compSvc, insightsSvc)
	apiH.Register(mux)

	var root http.Handler = mux
	root = handler.SessionMiddleware(tokens, store, root)
	root = handler.RateLimitMiddleware(limiter, root)
	root = handler.SecurityHeadersMiddleware(root)

	srv := httptest.NewServer(root)
	t.Cleanup(srv.Close)

	return &testApp{
		server:      srv,
		store:       store,
		tokens:      tokens,
		authSvc:     authSvc,
		catalogSvc:  catalogSvc,
		tripSvc:     tripSvc,
		syncSvc:     syncSvc,
		compSvc:     compSvc,
		insightsSvc: insightsSvc,
	}
}

func TestEndToEndRealUserExperience(t *testing.T) {
	app := newTestApp(t)
	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 1. Unauthenticated visitor to "/" is redirected to "/login"
	resp, err := noRedirectClient.Get(app.server.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/login" {
		t.Fatalf("expected unauthenticated GET / to redirect to /login, got status=%d loc=%s", resp.StatusCode, resp.Header.Get("Location"))
	}

	// 2. Register a brand-new user account via POST /register
	regForm := url.Values{
		"name":              {"Valentina Rossi"},
		"email":             {"valentina@viajera.com"},
		"password":          {"MiClaveSegura2026"},
		"homeCurrency":      {"EUR"},
		"preferredAgencyId": {"ag-meridian"},
	}
	regResp, err := noRedirectClient.PostForm(app.server.URL+"/register", regForm)
	if err != nil {
		t.Fatalf("POST /register failed: %v", err)
	}
	regResp.Body.Close()
	if regResp.StatusCode != http.StatusSeeOther || regResp.Header.Get("Location") != "/welcome" {
		t.Fatalf("expected POST /register to redirect to /welcome, got status=%d loc=%s", regResp.StatusCode, regResp.Header.Get("Location"))
	}

	var sessionCookie *http.Cookie
	for _, c := range regResp.Cookies() {
		if c.Name == handler.SessionCookieName && c.Value != "" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("expected session cookie after registration")
	}

	// Verify Idea B: newly registered user starts with 0 trips and 0 inbox confirmations
	valentina, err := app.store.GetUserByEmail("valentina@viajera.com")
	if err != nil {
		t.Fatalf("expected registered user in store: %v", err)
	}
	initialTrips, _ := app.tripSvc.ListTrips(valentina.ID, "ALL", "ALL")
	if len(initialTrips) != 0 {
		t.Fatalf("expected 0 initial trips for newly registered user, got %d", len(initialTrips))
	}
	initialInbox := app.syncSvc.ListConfirmations(valentina.ID)
	if len(initialInbox) != 0 {
		t.Fatalf("expected 0 initial inbox emails for newly registered user, got %d", len(initialInbox))
	}

	// 3. Verify Login & Password Validation
	_, _, badLoginErr := app.authSvc.Login("valentina@viajera.com", "clave-incorrecta")
	if badLoginErr == nil {
		t.Fatalf("expected login with wrong password to fail")
	}
	loggedUser, newTok, goodLoginErr := app.authSvc.Login("valentina@viajera.com", "MiClaveSegura2026")
	if goodLoginErr != nil || loggedUser.ID != valentina.ID || newTok == "" {
		t.Fatalf("expected valid login to succeed, got err=%v", goodLoginErr)
	}

	// 4. Book a Preloaded Travel Package from the Catalog
	packages := app.catalogSvc.ListPackages()
	if len(packages) < 4 {
		t.Fatalf("expected preloaded catalog packages, got %d", len(packages))
	}
	japanPkg := packages[0]
	bookedTrip, segCount, err := app.catalogSvc.BookPackage(valentina.ID, japanPkg.ID, "")
	if err != nil {
		t.Fatalf("BookPackage failed: %v", err)
	}
	if segCount < 2 || len(bookedTrip.Reservations) < 2 {
		t.Fatalf("expected package to create trip with multiple segments, got %d", segCount)
	}
	// Verify confirmation vouchers were delivered to Valentina's inbox
	inboxAfterPkg := app.syncSvc.ListConfirmations(valentina.ID)
	if len(inboxAfterPkg) < 2 {
		t.Fatalf("expected confirmation emails in inbox after booking package, got %d", len(inboxAfterPkg))
	}

	// 5. Send an individual offer to inbox and synchronize it
	offers := app.catalogSvc.ListOffers("TRAIN", "")
	if len(offers) == 0 {
		t.Fatalf("expected preloaded train offers in catalog")
	}
	pendingConf, err := app.catalogSvc.SendOfferToInbox(valentina.ID, offers[0].ID)
	if err != nil || pendingConf.Status != "PENDING" {
		t.Fatalf("SendOfferToInbox failed: %v", err)
	}
	syncBatch, err := app.syncSvc.SyncAllInbox(valentina.ID)
	if err != nil || syncBatch.ImportedCount < 1 {
		t.Fatalf("expected SyncAllInbox to import pending confirmation, got %+v err=%v", syncBatch, err)
	}

	// 6. Import an external reservation by PNR code
	_, importedRes, err := app.syncSvc.ImportByPNR(valentina.ID, "BKG-9012")
	if err != nil || importedRes == nil {
		t.Fatalf("ImportByPNR failed: %v", err)
	}

	// 7. Manual edit protection & Sharing trip by email
	customNote := "Solicité habitación en piso alto con vista"
	_, editedRes, err := app.tripSvc.UpdateReservation(valentina.ID, importedRes.ID, serviceReservationNotePatch(customNote))
	if err != nil || !editedRes.ManuallyModified {
		t.Fatalf("UpdateReservation failed: %v", err)
	}

	// Register a second user and share the trip by email
	companionUser, _, err := app.authSvc.Register("Mateo Silva", "mateo@viajero.com", "ClaveMateo123", "USD", "ag-meridian")
	if err != nil {
		t.Fatalf("register companion failed: %v", err)
	}
	// Before sharing, Mateo cannot access Valentina's trip
	if _, err := app.tripSvc.GetTripForUser(companionUser.ID, bookedTrip.ID, "127.0.0.1"); err == nil {
		t.Fatalf("expected uninvited user to be denied access to Valentina's trip")
	}
	// Share by email
	if _, err := app.tripSvc.ShareTripByEmail(valentina.ID, bookedTrip.ID, "mateo@viajero.com"); err != nil {
		t.Fatalf("ShareTripByEmail failed: %v", err)
	}
	// After sharing by email, Mateo can access the trip
	if _, err := app.tripSvc.GetTripForUser(companionUser.ID, bookedTrip.ID, "127.0.0.1"); err != nil {
		t.Fatalf("expected invited companion to access shared trip, got: %v", err)
	}

	// 8. Destination Companion Route, Weather & Currency Exchange
	route, err := app.compSvc.BuildNavigationRoute(valentina.ID, bookedTrip.ID, "EUR", "airport")
	if err != nil {
		t.Fatalf("BuildNavigationRoute failed: %v", err)
	}
	if len(route.PolylineCoords) < 2 || route.Weather.ConditionText == "" || route.ExchangeRate <= 0 {
		t.Fatalf("expected valid route polyline, weather, and exchange rate, got %+v", route)
	}

	// 9. Verify all HTML pages render cleanly without any academic ADD jargon or SwitchUser dropdown
	pagesToCheck := []string{"/", "/welcome", "/explore", "/explore/packages", "/inbox", "/companion", "/trends"}
	bannedTerms := []string{"switch-user", "QA-1", "US-01", "ADR-", "Circuit Breaker", "DLQ", "Simular caída"}
	for _, path := range pagesToCheck {
		req, _ := http.NewRequest(http.MethodGet, app.server.URL+path, nil)
		req.AddCookie(sessionCookie)
		res, err := noRedirectClient.Do(req)
		if err != nil {
			t.Fatalf("GET %s failed: %v", path, err)
		}
		raw, _ := io.ReadAll(res.Body)
		res.Body.Close()
		body := string(raw)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("GET %s returned status %d: %s", path, res.StatusCode, body)
		}
		for _, banned := range bannedTerms {
			if strings.Contains(body, banned) {
				t.Fatalf("page %s contains prohibited term %q", path, banned)
			}
		}
	}

	// 10. Verify login page wording & login redirect to /welcome
	loginPageResp, err := noRedirectClient.Get(app.server.URL + "/login")
	if err != nil {
		t.Fatalf("GET /login failed: %v", err)
	}
	loginRaw, _ := io.ReadAll(loginPageResp.Body)
	loginPageResp.Body.Close()
	loginBody := string(loginRaw)
	if !strings.Contains(loginBody, "Crear mi cuenta") || strings.Contains(loginBody, "Crear mi cuenta gratis") {
		t.Fatalf("expected login page to show 'Crear mi cuenta' without 'gratis'")
	}

	loginPostResp, err := noRedirectClient.PostForm(app.server.URL+"/login", url.Values{
		"email":    {"valentina@viajera.com"},
		"password": {"MiClaveSegura2026"},
	})
	if err != nil {
		t.Fatalf("POST /login failed: %v", err)
	}
	loginPostResp.Body.Close()
	if loginPostResp.StatusCode != http.StatusSeeOther || loginPostResp.Header.Get("Location") != "/welcome" {
		t.Fatalf("expected POST /login to redirect to /welcome, got status=%d loc=%s", loginPostResp.StatusCode, loginPostResp.Header.Get("Location"))
	}
}

func serviceReservationNotePatch(note string) domain.ReservationUpdateInput {
	return domain.ReservationUpdateInput{
		Notes: &note,
	}
}

func TestHTMXEndpointsAndPartialsRenderCleanly(t *testing.T) {
	app := newTestApp(t)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Register a user and capture session cookie
	regResp, err := client.PostForm(app.server.URL+"/register", url.Values{
		"name":              {"Martín Gómez"},
		"email":             {"martin@viajero.com"},
		"password":          {"Martin2026!"},
		"homeCurrency":      {"USD"},
		"preferredAgencyId": {"ag-meridian"},
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	regResp.Body.Close()

	var cookie *http.Cookie
	for _, c := range regResp.Cookies() {
		if c.Name == handler.SessionCookieName && c.Value != "" {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatalf("missing session cookie")
	}

	doPost := func(path string, form url.Values) (int, string) {
		req, _ := http.NewRequest(http.MethodPost, app.server.URL+path, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Request", "true")
		req.AddCookie(cookie)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("POST %s failed: %v", path, err)
		}
		buf := make([]byte, 64*1024)
		n, _ := res.Body.Read(buf)
		res.Body.Close()
		return res.StatusCode, string(buf[:n])
	}

	doGet := func(path string) (int, string) {
		req, _ := http.NewRequest(http.MethodGet, app.server.URL+path, nil)
		req.Header.Set("HX-Request", "true")
		req.AddCookie(cookie)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("GET %s failed: %v", path, err)
		}
		buf := make([]byte, 64*1024)
		n, _ := res.Body.Read(buf)
		res.Body.Close()
		return res.StatusCode, string(buf[:n])
	}

	// 1. Catalog partial filtering
	code, body := doGet("/partials/catalog/offers?type=FLIGHT&q=Tokio")
	if code != http.StatusOK || !strings.Contains(body, "All Nippon Airways") {
		t.Fatalf("expected flight offer for Tokio, status=%d body=%s", code, body)
	}

	// 2. Book package via HTMX
	pkgs := app.catalogSvc.ListPackages()
	code, body = doPost("/catalog/packages/"+pkgs[0].ID+"/book", url.Values{"dateId": {pkgs[0].AvailableDates[0].ID}})
	if code != http.StatusOK || !strings.Contains(body, "Paquete reservado") {
		t.Fatalf("expected package booking toast, status=%d body=%s", code, body)
	}

	// 3. Send offer to email and sync all inbox via HTMX
	offers := app.catalogSvc.ListOffers("HOTEL", "")
	code, _ = doPost("/catalog/offers/"+offers[0].ID+"/send-email", url.Values{})
	if code != http.StatusOK {
		t.Fatalf("send offer email failed with status=%d", code)
	}
	code, body = doPost("/inbox/sync-all", url.Values{})
	if code != http.StatusOK || !strings.Contains(body, "Bandeja sincronizada") {
		t.Fatalf("sync-all failed: status=%d body=%s", code, body)
	}

	// 4. Import by PNR via HTMX
	code, body = doPost("/inbox/import-pnr", url.Values{"pnr": {"AVE-6102"}})
	if code != http.StatusOK || !strings.Contains(body, "AVE-6102") {
		t.Fatalf("import-pnr failed: status=%d body=%s", code, body)
	}

	// 5. Companion route partial and Trends partial
	u, _ := app.store.GetUserByEmail("martin@viajero.com")
	trips, _ := app.tripSvc.ListTrips(u.ID, "ACTIVE", "ALL")
	if len(trips) == 0 {
		t.Fatalf("expected active trips for user")
	}
	code, body = doGet("/companion/route?tripId=" + trips[0].ID + "&origin=station&currency=EUR")
	if code != http.StatusOK || !strings.Contains(body, "leaflet-companion-map") {
		t.Fatalf("companion route partial failed: status=%d body=%s", code, body)
	}

	code, body = doGet("/partials/trends")
	if code != http.StatusOK || !strings.Contains(body, "Destinos más reservados") {
		t.Fatalf("trends partial failed: status=%d body=%s", code, body)
	}

	// 6. Verify UpdatePreferences returns HX-Refresh: true and converts catalog prices dynamically
	prefReq, _ := http.NewRequest(http.MethodPost, app.server.URL+"/profile/preferences", strings.NewReader(url.Values{
		"agencyId":     {"ag-iberia"},
		"homeCurrency": {"EUR"},
	}.Encode()))
	prefReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	prefReq.Header.Set("HX-Request", "true")
	prefReq.AddCookie(cookie)
	prefRes, err := client.Do(prefReq)
	if err != nil {
		t.Fatalf("POST /profile/preferences failed: %v", err)
	}
	prefRes.Body.Close()
	if prefRes.StatusCode != http.StatusOK || prefRes.Header.Get("HX-Refresh") != "true" {
		t.Fatalf("expected HX-Refresh: true on /profile/preferences, got status=%d header=%q", prefRes.StatusCode, prefRes.Header.Get("HX-Refresh"))
	}

	code, pkgBody := doGet("/explore/packages")
	if code != http.StatusOK || !strings.Contains(pkgBody, "EUR ") || strings.Contains(pkgBody, "USD 2890.00") {
		t.Fatalf("expected /explore/packages to render converted EUR prices without USD 2890.00, status=%d", code)
	}

	// Switch currency to ARS and verify offers convert cleanly
	prefARSReq, _ := http.NewRequest(http.MethodPost, app.server.URL+"/profile/preferences", strings.NewReader(url.Values{
		"agencyId":     {"ag-iberia"},
		"homeCurrency": {"ARS"},
	}.Encode()))
	prefARSReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	prefARSReq.Header.Set("HX-Request", "true")
	prefARSReq.AddCookie(cookie)
	prefARSRes, err := client.Do(prefARSReq)
	if err != nil {
		t.Fatalf("POST /profile/preferences (ARS) failed: %v", err)
	}
	prefARSRes.Body.Close()

	code, arsBody := doGet("/explore")
	if code != http.StatusOK || !strings.Contains(arsBody, "ARS ") || strings.Contains(arsBody, "USD ") {
		t.Fatalf("expected /explore to render converted ARS prices without USD prices, status=%d", code)
	}

	// 7. Logout revokes session
	logoutCode, _ := doPost("/logout", url.Values{})
	if logoutCode != http.StatusSeeOther {
		t.Fatalf("expected logout redirect, got %d", logoutCode)
	}
	postLogoutCode, _ := doGet("/")
	if postLogoutCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 on HTMX request with revoked session, got %d", postLogoutCode)
	}
}

func TestBookOfferTripSelectionModal(t *testing.T) {
	app := newTestApp(t)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	regResp, err := client.PostForm(app.server.URL+"/register", url.Values{
		"name":              {"Lucía Pereyra"},
		"email":             {"lucia.modal@viajera.com"},
		"password":          {"Lucia2026!"},
		"homeCurrency":      {"USD"},
		"preferredAgencyId": {"ag-andes"},
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	regResp.Body.Close()

	var cookie *http.Cookie
	for _, c := range regResp.Cookies() {
		if c.Name == handler.SessionCookieName && c.Value != "" {
			cookie = c
			break
		}
	}
	if cookie == nil {
		t.Fatalf("missing session cookie")
	}

	u, err := app.store.GetUserByEmail("lucia.modal@viajera.com")
	if err != nil {
		t.Fatalf("user lookup failed: %v", err)
	}

	italyTrip, err := app.tripSvc.CreateTrip(u.ID, "Viaje a Italia", "Roma", "IT")
	if err != nil {
		t.Fatalf("create italy trip failed: %v", err)
	}
	argentinaTrip, err := app.tripSvc.CreateTrip(u.ID, "Viaje a Argentina", "Bariloche", "AR")
	if err != nil {
		t.Fatalf("create argentina trip failed: %v", err)
	}

	barilocheOffers := app.catalogSvc.ListOffers("ALL", "Bariloche")
	if len(barilocheOffers) == 0 {
		t.Fatalf("expected Bariloche offers in catalog")
	}
	offer := barilocheOffers[0]

	// 1. Open trip selection modal
	modalReq, _ := http.NewRequest(http.MethodGet, app.server.URL+"/partials/catalog/offers/"+offer.ID+"/book-modal", nil)
	modalReq.Header.Set("HX-Request", "true")
	modalReq.AddCookie(cookie)
	modalRes, err := client.Do(modalReq)
	if err != nil {
		t.Fatalf("GET book-modal failed: %v", err)
	}
	modalBytes, _ := io.ReadAll(modalRes.Body)
	modalRes.Body.Close()
	modalBody := string(modalBytes)
	if modalRes.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on book-modal, got %d: %s", modalRes.StatusCode, modalBody)
	}
	if !strings.Contains(modalBody, "Viaje a Italia") || !strings.Contains(modalBody, "Viaje a Argentina") || !strings.Contains(modalBody, `value="NEW"`) {
		t.Fatalf("book-modal missing expected trip choices or NEW option: %s", modalBody)
	}

	// 2. Book explicitly into Viaje a Argentina
	bookReq, _ := http.NewRequest(http.MethodPost, app.server.URL+"/catalog/offers/"+offer.ID+"/book", strings.NewReader(url.Values{
		"tripId": {argentinaTrip.ID},
	}.Encode()))
	bookReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	bookReq.Header.Set("HX-Request", "true")
	bookReq.AddCookie(cookie)
	bookRes, err := client.Do(bookReq)
	if err != nil {
		t.Fatalf("POST book offer failed: %v", err)
	}
	bookRes.Body.Close()
	if bookRes.StatusCode != http.StatusOK || bookRes.Header.Get("HX-Trigger") != "closeModal" {
		t.Fatalf("expected 200 with HX-Trigger: closeModal, got status=%d trigger=%q", bookRes.StatusCode, bookRes.Header.Get("HX-Trigger"))
	}

	updatedItaly, _ := app.store.GetTrip(italyTrip.ID)
	updatedArg, _ := app.store.GetTrip(argentinaTrip.ID)
	if len(updatedItaly.Reservations) != 0 {
		t.Fatalf("expected 0 reservations in Viaje a Italia, got %d", len(updatedItaly.Reservations))
	}
	if len(updatedArg.Reservations) != 1 {
		t.Fatalf("expected 1 reservation in Viaje a Argentina, got %d", len(updatedArg.Reservations))
	}

	// 3. Book creating a NEW trip with custom title
	newBookReq, _ := http.NewRequest(http.MethodPost, app.server.URL+"/catalog/offers/"+offer.ID+"/book", strings.NewReader(url.Values{
		"tripId":       {"NEW"},
		"newTripTitle": {"Escapada Bariloche"},
	}.Encode()))
	newBookReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	newBookReq.Header.Set("HX-Request", "true")
	newBookReq.AddCookie(cookie)
	newBookRes, err := client.Do(newBookReq)
	if err != nil {
		t.Fatalf("POST book offer NEW failed: %v", err)
	}
	newBookRes.Body.Close()
	if newBookRes.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for tripId=NEW, got %d", newBookRes.StatusCode)
	}

	tripsAfter, _ := app.tripSvc.ListTrips(u.ID, "ACTIVE", "ALL")
	foundNew := false
	for _, tr := range tripsAfter {
		if tr.Title == "Escapada Bariloche" && len(tr.Reservations) == 1 {
			foundNew = true
			break
		}
	}
	if !foundNew {
		t.Fatalf("expected new trip 'Escapada Bariloche' with 1 reservation to be created")
	}

	// 4. Verify /explore and /explore/packages only have direct booking buttons (no send-email buttons)
	exploreReq, _ := http.NewRequest(http.MethodGet, app.server.URL+"/explore", nil)
	exploreReq.AddCookie(cookie)
	exploreRes, err := client.Do(exploreReq)
	if err != nil {
		t.Fatalf("GET /explore failed: %v", err)
	}
	exploreBytes, _ := io.ReadAll(exploreRes.Body)
	exploreRes.Body.Close()
	if strings.Contains(string(exploreBytes), "Recibir en mi correo") {
		t.Fatalf("expected /explore not to contain 'Recibir en mi correo'")
	}

	pkgsReq, _ := http.NewRequest(http.MethodGet, app.server.URL+"/explore/packages", nil)
	pkgsReq.AddCookie(cookie)
	pkgsRes, err := client.Do(pkgsReq)
	if err != nil {
		t.Fatalf("GET /explore/packages failed: %v", err)
	}
	pkgsBytes, _ := io.ReadAll(pkgsRes.Body)
	pkgsRes.Body.Close()
	if strings.Contains(string(pkgsBytes), "Enviar reserva a mi correo") {
		t.Fatalf("expected /explore/packages not to contain 'Enviar reserva a mi correo'")
	}

	// 5. Verify external email forwarding/scanning inside /inbox
	scanGoodReq, _ := http.NewRequest(http.MethodPost, app.server.URL+"/inbox/scan-email", strings.NewReader(url.Values{
		"fromEmail": {"reservations@booking.com"},
		"subject":   {"Confirmación de hotel en Roma"},
		"body":      {"Tu reserva está confirmada. Código PNR: EXT-7711."},
	}.Encode()))
	scanGoodReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	scanGoodReq.Header.Set("HX-Request", "true")
	scanGoodReq.AddCookie(cookie)
	scanGoodRes, err := client.Do(scanGoodReq)
	if err != nil {
		t.Fatalf("POST /inbox/scan-email (trusted) failed: %v", err)
	}
	scanGoodBytes, _ := io.ReadAll(scanGoodRes.Body)
	scanGoodRes.Body.Close()
	if scanGoodRes.StatusCode != http.StatusOK || !strings.Contains(string(scanGoodBytes), "EXT-7711") {
		t.Fatalf("expected trusted external email to be imported, status=%d", scanGoodRes.StatusCode)
	}

	scanBadReq, _ := http.NewRequest(http.MethodPost, app.server.URL+"/inbox/scan-email", strings.NewReader(url.Values{
		"fromEmail": {"ofertas@agencia-desconocida.xyz"},
		"subject":   {"Promo sospechosa"},
		"body":      {"Código PNR: SPAM-1111."},
	}.Encode()))
	scanBadReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	scanBadReq.Header.Set("HX-Request", "true")
	scanBadReq.AddCookie(cookie)
	scanBadRes, err := client.Do(scanBadReq)
	if err != nil {
		t.Fatalf("POST /inbox/scan-email (untrusted) failed: %v", err)
	}
	scanBadBytes, _ := io.ReadAll(scanBadRes.Body)
	scanBadRes.Body.Close()
	if !strings.Contains(string(scanBadBytes), "Correo no importado") {
		t.Fatalf("expected untrusted sender to be rejected")
	}
}
