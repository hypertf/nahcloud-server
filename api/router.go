package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/hypertf/nahcloud/web"
)

var startTime = time.Now()

// BuildInfo contains server build and runtime information
type BuildInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Uptime    string `json:"uptime"`
}

// SetupRouter creates and configures the HTTP router
func SetupRouter(handler *Handler, version string) *mux.Router {
	router := mux.NewRouter()

	// Build info endpoint
	router.HandleFunc("/buildz", func(w http.ResponseWriter, r *http.Request) {
		info := BuildInfo{
			Version:   version,
			GoVersion: runtime.Version(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			Uptime:    time.Since(startTime).Round(time.Second).String(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}).Methods("GET")

	// Web console routes
	webHandler := web.NewHandler(handler.service)
	webRouter := router.PathPrefix("/web").Subrouter()
	
	// Dashboard
	webRouter.HandleFunc("", webHandler.Dashboard).Methods("GET")
	webRouter.HandleFunc("/", webHandler.Dashboard).Methods("GET")
	
	// Project routes
	webRouter.HandleFunc("/projects", webHandler.ListProjects).Methods("GET")
	webRouter.HandleFunc("/projects", webHandler.CreateProject).Methods("POST")
	webRouter.HandleFunc("/projects/new", webHandler.NewProjectForm).Methods("GET")
	webRouter.HandleFunc("/projects/{id}/edit", webHandler.EditProjectForm).Methods("GET")
	webRouter.HandleFunc("/projects/{id}", webHandler.UpdateProject).Methods("PUT")
	webRouter.HandleFunc("/projects/{id}", webHandler.DeleteProject).Methods("DELETE")
	
	// Instance routes
	webRouter.HandleFunc("/instances", webHandler.ListInstances).Methods("GET")
	webRouter.HandleFunc("/instances", webHandler.CreateInstance).Methods("POST")
	webRouter.HandleFunc("/instances/new", webHandler.NewInstanceForm).Methods("GET")
	webRouter.HandleFunc("/instances/{id}/edit", webHandler.EditInstanceForm).Methods("GET")
	webRouter.HandleFunc("/instances/{id}", webHandler.UpdateInstance).Methods("PUT")
	webRouter.HandleFunc("/instances/{id}", webHandler.DeleteInstance).Methods("DELETE")
	
	// Metadata routes
	webRouter.HandleFunc("/metadata", webHandler.ListMetadata).Methods("GET")
	webRouter.HandleFunc("/metadata", webHandler.CreateMetadata).Methods("POST")
	webRouter.HandleFunc("/metadata/new", webHandler.NewMetadataForm).Methods("GET")
	webRouter.HandleFunc("/metadata/edit", webHandler.EditMetadataForm).Methods("GET")
	webRouter.HandleFunc("/metadata/update", webHandler.UpdateMetadata).Methods("PUT")
	webRouter.HandleFunc("/metadata/delete", webHandler.DeleteMetadata).Methods("DELETE")

    // Storage routes
    webRouter.HandleFunc("/storage", webHandler.ListStorage).Methods("GET")
    webRouter.HandleFunc("/storage/buckets/new", webHandler.NewBucketForm).Methods("GET")
    webRouter.HandleFunc("/storage/buckets", webHandler.CreateBucket).Methods("POST")
    webRouter.HandleFunc("/storage/buckets/{name}/objects", webHandler.ListBucketObjects).Methods("GET")
    webRouter.HandleFunc("/storage/buckets/{name}/objects/new", webHandler.NewObjectForm).Methods("GET")
    webRouter.HandleFunc("/storage/buckets/{name}/objects", webHandler.CreateObject).Methods("POST")
    webRouter.HandleFunc("/storage/buckets/{name}/objects/{objid}", webHandler.ViewObject).Methods("GET")

	// API prefix
	api := router.PathPrefix("/v1").Subrouter()

	// Project routes
	api.HandleFunc("/projects", handler.CreateProject).Methods("POST")
	api.HandleFunc("/projects", handler.ListProjects).Methods("GET")
	api.HandleFunc("/projects/{id}", handler.GetProject).Methods("GET")
	api.HandleFunc("/projects/{id}", handler.UpdateProject).Methods("PATCH")
	api.HandleFunc("/projects/{id}", handler.DeleteProject).Methods("DELETE")

	// Instance routes
	api.HandleFunc("/instances", handler.CreateInstance).Methods("POST")
	api.HandleFunc("/instances", handler.ListInstances).Methods("GET")
	api.HandleFunc("/instances/{id}", handler.GetInstance).Methods("GET")
	api.HandleFunc("/instances/{id}", handler.UpdateInstance).Methods("PATCH")
	api.HandleFunc("/instances/{id}", handler.DeleteInstance).Methods("DELETE")

	// Metadata routes
	api.HandleFunc("/metadata", handler.CreateMetadata).Methods("POST")
	api.HandleFunc("/metadata", handler.ListMetadata).Methods("GET").Queries("prefix", "")
	api.HandleFunc("/metadata", handler.ListMetadata).Methods("GET")
	api.HandleFunc("/metadata/{id}", handler.GetMetadata).Methods("GET")
	api.HandleFunc("/metadata/{id}", handler.UpdateMetadata).Methods("PATCH")
	api.HandleFunc("/metadata/{id}", handler.DeleteMetadata).Methods("DELETE")

	// Storage bucket routes
	api.HandleFunc("/buckets", handler.CreateBucket).Methods("POST")
	api.HandleFunc("/buckets", handler.ListBuckets).Methods("GET")
	api.HandleFunc("/buckets/{id}", handler.GetBucket).Methods("GET")
	api.HandleFunc("/buckets/{id}", handler.UpdateBucket).Methods("PATCH")
	api.HandleFunc("/buckets/{id}", handler.DeleteBucket).Methods("DELETE")

	// Bucket-scoped object routes
	api.HandleFunc("/bucket/{bucket_id}/objects", handler.CreateObject).Methods("POST")
	api.HandleFunc("/bucket/{bucket_id}/objects", handler.ListObjects).Methods("GET") // optional: prefix
	api.HandleFunc("/bucket/{bucket_id}/objects/{id}", handler.GetObject).Methods("GET")
	api.HandleFunc("/bucket/{bucket_id}/objects/{id}", handler.UpdateObject).Methods("PATCH")
	api.HandleFunc("/bucket/{bucket_id}/objects/{id}", handler.DeleteObject).Methods("DELETE")

	// Terraform state routes
	api.HandleFunc("/tfstate/{id}", handler.TFStateGet).Methods("GET")
	api.HandleFunc("/tfstate/{id}", handler.TFStatePost).Methods("POST")
	api.HandleFunc("/tfstate/{id}", handler.TFStateDelete).Methods("DELETE")
	api.HandleFunc("/tfstate/{id}", handler.TFStateLock).Methods("LOCK")
	api.HandleFunc("/tfstate/{id}", handler.TFStateUnlock).Methods("UNLOCK")

	// Add CORS middleware for development
	router.Use(corsMiddleware)

	// Add logging middleware
	router.Use(loggingMiddleware)

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Nah-No-Chaos, X-Nah-Latency")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware adds basic request logging
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add proper structured logging here
		// For now, we'll let the main server handle logging
		next.ServeHTTP(w, r)
	})
}