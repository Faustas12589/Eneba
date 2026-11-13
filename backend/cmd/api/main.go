package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"eneba-marketplace/backend/internal/config"
	"eneba-marketplace/backend/internal/handler"
	"eneba-marketplace/backend/internal/middleware"
	"eneba-marketplace/backend/internal/repository"
	"eneba-marketplace/backend/internal/router"
	"eneba-marketplace/backend/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	var repo repository.Catalog
	var verifier middleware.TokenVerifier

	switch cfg.Store {
	case "firestore":
		app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: cfg.FirebaseProj}, credOpts(cfg)...)
		if err != nil {
			log.Fatalf("firebase app: %v", err)
		}
		fs, err := firestore.NewClient(ctx, cfg.FirebaseProj, credOpts(cfg)...)
		if err != nil {
			log.Fatalf("firestore: %v", err)
		}
		defer fs.Close()
		authClient, err := app.Auth(ctx)
		if err != nil {
			log.Fatalf("firebase auth: %v", err)
		}
		repo = repository.NewFirestore(fs)
		verifier = middleware.FirebaseVerifier{Client: authClient}
	default:
		m, err := repository.NewMemory()
		if err != nil {
			log.Fatalf("memory repo: %v", err)
		}
		repo = m
		verifier = middleware.DevVerifier{}
		log.Println("STORE=memory: serving embedded seed data; auth accepts the dev token only")
	}

	h := &handler.Handler{Catalog: service.NewCatalog(repo)}
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router.New(h, verifier, cfg.CORSOrigins),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s (store=%s)", cfg.Port, cfg.Store)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

// credOpts: JSON in env (Render) wins; otherwise fall back to Application
// Default Credentials (GOOGLE_APPLICATION_CREDENTIALS file path locally).
func credOpts(cfg config.Config) []option.ClientOption {
	if cfg.FirebaseCredsJSON != "" {
		return []option.ClientOption{option.WithCredentialsJSON([]byte(cfg.FirebaseCredsJSON))}
	}
	return nil
}
