package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/vpramatarov/pdf-tools/internal/api/handlers"
	"github.com/vpramatarov/pdf-tools/internal/api/router"
	"github.com/vpramatarov/pdf-tools/internal/config"
)

func main() {
	// cancel resouces
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()

	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		panic("Failed to create upload directory: " + err.Error())
	}

	h := handlers.New(cfg)

	h.StartCleanupCron()

	r := router.New(h)

	host := "http://localhost"
	apiServer := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf(":%d", cfg.Port),
	}

	log.Printf("Server starting on %v:%d ...", host, cfg.Port)
	log.Printf("📂 Upload Limit: %d MB | Cleanup: Every %d min\n", cfg.MaxUploadSizeMB, cfg.CleanupIntervalMinutes)

	go func() {
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start on host %v: %v", host, err)
		}
	}()

	var wg sync.WaitGroup
	wg.Go(func() {
		<-ctx.Done()
		// allow server to complete any incomming requests and shut down in 10 seconds
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := apiServer.Shutdown(shutdownContext); err != nil {
			log.Fatalf("Server failed to shutdown %v", err)
		}
	})

	wg.Wait()
}
