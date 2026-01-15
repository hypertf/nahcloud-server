package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/hypertf/nahcloud/api"
	"github.com/hypertf/nahcloud/service"
	"github.com/hypertf/nahcloud/service/chaos"
	"github.com/hypertf/nahcloud/storage/sqlite"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	rootCmd := &cobra.Command{
		Use:   "nahcloud-server",
		Short: "NahCloud Server - A mock cloud provider for Terraform testing",
		Long: `NahCloud Server is a mock cloud provider designed for testing Terraform configurations.

It provides fake implementations of cloud resources like compute instances, projects,
and metadata services, with optional chaos engineering features for testing error
handling and resilience.` + printConfigHelp(),
		Version: Version,
		RunE:    runServer,
	}

	setupConfig(rootCmd)

	return rootCmd.Execute()
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	config, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database
	db, err := sqlite.NewDB(config.SQLiteDSN)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Initialize repositories
	projectRepo := sqlite.NewProjectRepository(db)
	instanceRepo := sqlite.NewInstanceRepository(db)
	metadataRepo := sqlite.NewMetadataRepository(db)
	bucketRepo := sqlite.NewBucketRepository(db)
	objectRepo := sqlite.NewObjectRepository(db)

	// Initialize service layer
	svc := service.NewService(projectRepo, instanceRepo, metadataRepo, bucketRepo, objectRepo)

	// Initialize chaos service with config
	chaosConfig := config.ToChaosConfig()
	chaosService := chaos.NewChaosServiceWithConfig(chaosConfig)

	// Initialize API handlers
	handler := api.NewHandler(svc, chaosService, config.Token)

	// Setup router
	router := api.SetupRouter(handler, Version)

	// Create HTTP server
	server := &http.Server{
		Addr:         config.Addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("NahCloud server starting on %s", config.Addr)
		if chaosConfig.Enabled {
			log.Printf("Chaos engineering enabled (seed: %d)", chaosConfig.Seed)
		}
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Printf("Received signal %v, starting graceful shutdown", sig)

		// Give outstanding requests 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			if err := server.Close(); err != nil {
				log.Printf("Force close failed: %v", err)
			}
		}
	}

	log.Println("Server stopped")
	return nil
}
