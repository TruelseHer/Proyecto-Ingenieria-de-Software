package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"trotamundos/internal/config"
	"trotamundos/internal/handler"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
	"trotamundos/internal/service"
)

func main() {
	cfg := config.Load()

	var (
		store *repository.Store
		err   error
	)
	if cfg.DatabaseURL != "" {
		store, err = repository.NewPostgresStore(cfg.DatabaseURL, cfg.ReadDatabaseURL)
		if err != nil {
			log.Printf("PostgreSQL no disponible (%v), iniciando con almacenamiento en memoria", err)
			store = repository.NewMemoryStore()
		} else {
			log.Printf("Conectado a PostgreSQL (lectura/escritura activas)")
		}
	} else {
		store = repository.NewMemoryStore()
	}
	defer store.Close()

	cache := platform.NewCacheClient(cfg.RedisAddr)
	tokens := platform.NewTokenService(cfg.TokenSecret, cache)
	limiter := platform.NewRateLimiter(cfg.RateLimitRequestsPerMin)
	bus := platform.NewEventBroker()
	defer bus.Close()
	guard := platform.NewProviderCircuitBreaker()
	extAPI := platform.NewExternalAPIClient(cache, cfg.ProviderTimeout)

	authSvc := service.NewAuthService(store, tokens)
	tripSvc := service.NewTripService(store, cache, bus)
	catalogSvc := service.NewCatalogService(store, tripSvc)
	syncSvc := service.NewSyncService(store, bus, guard, tripSvc)
	insightsSvc := service.NewInsightsService(store, bus, cfg.AnonymizationSalt)
	recSvc := service.NewRecommendationService(store, cache, insightsSvc)
	compSvc := service.NewCompanionService(store, cache, tokens, extAPI)

	renderer, err := handler.NewRenderer(cfg.WebTemplatesDir)
	if err != nil {
		log.Fatalf("Error al cargar plantillas HTML desde %s: %v", cfg.WebTemplatesDir, err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.WebStaticDir))))

	webHandler := handler.NewWebHandler(store, tokens, renderer, authSvc, catalogSvc, tripSvc, syncSvc, recSvc, compSvc, insightsSvc, extAPI)
	webHandler.Register(mux)

	apiHandler := handler.NewAPIHandler(store, cache, authSvc, catalogSvc, tripSvc, syncSvc, recSvc, compSvc, insightsSvc)
	apiHandler.Register(mux)

	var rootHandler http.Handler = mux
	rootHandler = handler.SessionMiddleware(tokens, store, rootHandler)
	rootHandler = handler.RateLimitMiddleware(limiter, rootHandler)
	rootHandler = handler.SecurityHeadersMiddleware(rootHandler)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           rootHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("Trotamundos V2 escuchando en http://localhost:%s (almacenamiento=%s, caché=%s)", cfg.Port, store.Mode(), cache.Mode())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error fatal del servidor HTTP: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
