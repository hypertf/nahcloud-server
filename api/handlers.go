package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hypertf/nahcloud/domain"
	"github.com/hypertf/nahcloud/service"
	"github.com/hypertf/nahcloud/service/chaos"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	service      *service.Service
	chaosService *chaos.ChaosService
	token        string
}

// NewHandler creates a new HTTP handler
func NewHandler(svc *service.Service, chaosService *chaos.ChaosService, token string) *Handler {
	return &Handler{
		service:      svc,
		chaosService: chaosService,
		token:        token,
	}
}

// authenticate checks bearer token authentication
func (h *Handler) authenticate(r *http.Request) error {
	if h.token == "" {
		return nil // No authentication required
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return domain.UnauthorizedError("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return domain.UnauthorizedError("invalid authorization header format")
	}

	if parts[1] != h.token {
		return domain.UnauthorizedError("invalid token")
	}

	return nil
}

// writeError writes a domain error as JSON response
func (h *Handler) writeError(w http.ResponseWriter, err error) {
	var statusCode int
	var nahErr *domain.NahError

	if de, ok := err.(*domain.NahError); ok {
		nahErr = de
		switch de.Code {
		case domain.ErrorCodeNotFound:
			statusCode = http.StatusNotFound
		case domain.ErrorCodeAlreadyExists:
			statusCode = http.StatusConflict
		case domain.ErrorCodeInvalidInput:
			statusCode = http.StatusBadRequest
		case domain.ErrorCodeForeignKeyViolation:
			statusCode = http.StatusBadRequest
		case domain.ErrorCodeUnauthorized:
			statusCode = http.StatusUnauthorized
		case domain.ErrorCodeTooManyRequests:
			statusCode = http.StatusTooManyRequests
		case domain.ErrorCodeServiceUnavailable:
			statusCode = http.StatusServiceUnavailable
		default:
			statusCode = http.StatusInternalServerError
		}
	} else {
		statusCode = http.StatusInternalServerError
		nahErr = domain.InternalError(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(nahErr)
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeText writes a plain text response
func (h *Handler) writeText(w http.ResponseWriter, statusCode int, text string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(text))
}

// Project handlers

// CreateProject handles POST /v1/projects
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyProjectsChaos(r.Context(), r, "POST"); err != nil {
		h.writeError(w, err)
		return
	}

	var req domain.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}

	project, err := h.service.CreateProject(req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, project)
}

// GetProject handles GET /v1/projects/{id}
func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyProjectsChaos(r.Context(), r, "GET"); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	project, err := h.service.GetProject(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, project)
}

// ListProjects handles GET /v1/projects
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyProjectsChaos(r.Context(), r, "GET"); err != nil {
		h.writeError(w, err)
		return
	}

	opts := domain.ProjectListOptions{
		Name: r.URL.Query().Get("name"),
	}

	projects, err := h.service.ListProjects(opts)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, projects)
}

// UpdateProject handles PATCH /v1/projects/{id}
func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyProjectsChaos(r.Context(), r, "PATCH"); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var req domain.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}

	project, err := h.service.UpdateProject(id, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, project)
}

// DeleteProject handles DELETE /v1/projects/{id}
func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyProjectsChaos(r.Context(), r, "DELETE"); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	err := h.service.DeleteProject(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Instance handlers

// CreateInstance handles POST /v1/instances
func (h *Handler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyInstancesChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	var req domain.CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}

	instance, err := h.service.CreateInstance(req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, instance)
}

// GetInstance handles GET /v1/instances/{id}
func (h *Handler) GetInstance(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyInstancesChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	instance, err := h.service.GetInstance(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, instance)
}

// ListInstances handles GET /v1/instances
func (h *Handler) ListInstances(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyInstancesChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	opts := domain.InstanceListOptions{
		ProjectID: r.URL.Query().Get("project_id"),
		Name:      r.URL.Query().Get("name"),
		Status:    r.URL.Query().Get("status"),
	}

	instances, err := h.service.ListInstances(opts)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, instances)
}

// UpdateInstance handles PATCH /v1/instances/{id}
func (h *Handler) UpdateInstance(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyInstancesChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var req domain.UpdateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}

	instance, err := h.service.UpdateInstance(id, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, instance)
}

// DeleteInstance handles DELETE /v1/instances/{id}
func (h *Handler) DeleteInstance(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyInstancesChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	err := h.service.DeleteInstance(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Metadata handlers

// CreateMetadata handles POST /v1/metadata
func (h *Handler) CreateMetadata(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyMetadataChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	var req domain.CreateMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}

	metadata, err := h.service.CreateMetadata(req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, metadata)
}

// GetMetadata handles GET /v1/metadata/{id}
func (h *Handler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyMetadataChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	metadata, err := h.service.GetMetadata(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, metadata)
}

// ListMetadata handles GET /v1/metadata with prefix query parameter
func (h *Handler) ListMetadata(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyMetadataChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	opts := domain.MetadataListOptions{
		Prefix: r.URL.Query().Get("prefix"),
	}

	metadata, err := h.service.ListMetadata(opts)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, metadata)
}

// UpdateMetadata handles PATCH /v1/metadata/{id}
func (h *Handler) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyMetadataChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var req domain.UpdateMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}

	metadata, err := h.service.UpdateMetadata(id, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, metadata)
}

// DeleteMetadata handles DELETE /v1/metadata/{id}
func (h *Handler) DeleteMetadata(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}

	if err := h.chaosService.ApplyMetadataChaos(r.Context(), r); err != nil {
		h.writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	err := h.service.DeleteMetadata(id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Bucket handlers

// CreateBucket handles POST /v1/buckets
func (h *Handler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	var req domain.CreateBucketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}
	bucket, err := h.service.CreateBucket(req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, bucket)
}

// GetBucket handles GET /v1/buckets/{id}
func (h *Handler) GetBucket(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	bucket, err := h.service.GetBucket(id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, bucket)
}

// ListBuckets handles GET /v1/buckets
func (h *Handler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	opts := domain.BucketListOptions{ Name: r.URL.Query().Get("name") }
	buckets, err := h.service.ListBuckets(opts)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, buckets)
}

// UpdateBucket handles PATCH /v1/buckets/{id}
func (h *Handler) UpdateBucket(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	var req domain.UpdateBucketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}
	bucket, err := h.service.UpdateBucket(id, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, bucket)
}

// DeleteBucket handles DELETE /v1/buckets/{id}
func (h *Handler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	if err := h.service.DeleteBucket(id); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Object handlers

// CreateObject handles POST /v1/bucket/{bucket_id}/objects
func (h *Handler) CreateObject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	bucketID := vars["bucket_id"]
	var req domain.CreateObjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}
	// Force the bucket from the URL
	req.BucketID = bucketID
	obj, err := h.service.CreateObject(req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, obj)
}

// GetObject handles GET /v1/bucket/{bucket_id}/objects/{id}
func (h *Handler) GetObject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	bucketID := vars["bucket_id"]
	id := vars["id"]
	obj, err := h.service.GetObject(id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	// Enforce object belongs to the requested bucket
	if obj.BucketID != bucketID {
		h.writeError(w, domain.NotFoundError("object", id))
		return
	}
	h.writeJSON(w, http.StatusOK, obj)
}

// ListObjects handles GET /v1/bucket/{bucket_id}/objects
func (h *Handler) ListObjects(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	bucketID := vars["bucket_id"]
	opts := domain.ObjectListOptions{
		BucketID: bucketID,
		Prefix:   r.URL.Query().Get("prefix"),
	}
	objects, err := h.service.ListObjects(opts)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, objects)
}

// UpdateObject handles PATCH /v1/bucket/{bucket_id}/objects/{id}
func (h *Handler) UpdateObject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	bucketID := vars["bucket_id"]
	id := vars["id"]
	var req domain.UpdateObjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, domain.InvalidInputError("invalid JSON", nil))
		return
	}
	obj, err := h.service.UpdateObject(id, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	if obj.BucketID != bucketID {
		h.writeError(w, domain.NotFoundError("object", id))
		return
	}
	h.writeJSON(w, http.StatusOK, obj)
}

// DeleteObject handles DELETE /v1/bucket/{bucket_id}/objects/{id}
func (h *Handler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	bucketID := vars["bucket_id"]
	id := vars["id"]
	// Ensure object belongs to bucket before deleting
	obj, err := h.service.GetObject(id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	if obj.BucketID != bucketID {
		h.writeError(w, domain.NotFoundError("object", id))
		return
	}
	if err := h.service.DeleteObject(id); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
