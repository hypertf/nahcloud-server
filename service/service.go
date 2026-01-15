package service

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"github.com/hypertf/nahcloud/domain"
)

// Service provides business logic for NahCloud operations
type Service struct {
	projectRepo  ProjectRepository
	instanceRepo InstanceRepository
	metadataRepo MetadataRepository
	bucketRepo   BucketRepository
	objectRepo   ObjectRepository
}

// ProjectRepository defines the interface for project data operations
type ProjectRepository interface {
	Create(project *domain.Project) error
	GetByID(id string) (*domain.Project, error)
	GetByName(name string) (*domain.Project, error)
	List(opts domain.ProjectListOptions) ([]*domain.Project, error)
	Update(id string, req domain.UpdateProjectRequest) (*domain.Project, error)
	Delete(id string) error
}

// InstanceRepository defines the interface for instance data operations
type InstanceRepository interface {
	Create(instance *domain.Instance) error
	GetByID(id string) (*domain.Instance, error)
	List(opts domain.InstanceListOptions) ([]*domain.Instance, error)
	Update(id string, req domain.UpdateInstanceRequest) (*domain.Instance, error)
	Delete(id string) error
}

// MetadataRepository defines the interface for metadata data operations
type MetadataRepository interface {
	Create(req domain.CreateMetadataRequest) (*domain.Metadata, error)
	GetByID(id string) (*domain.Metadata, error)
	Update(id string, req domain.UpdateMetadataRequest) (*domain.Metadata, error)
	List(opts domain.MetadataListOptions) ([]*domain.Metadata, error)
	Delete(id string) error
}

// BucketRepository defines the interface for bucket data operations
type BucketRepository interface {
	Create(bucket *domain.Bucket) error
	GetByID(id string) (*domain.Bucket, error)
	GetByName(name string) (*domain.Bucket, error)
	List(opts domain.BucketListOptions) ([]*domain.Bucket, error)
	Update(id string, req domain.UpdateBucketRequest) (*domain.Bucket, error)
	Delete(id string) error
}

// ObjectRepository defines the interface for object data operations
type ObjectRepository interface {
	Create(req domain.CreateObjectRequest) (*domain.Object, error)
	GetByID(id string) (*domain.Object, error)
	Update(id string, req domain.UpdateObjectRequest) (*domain.Object, error)
	List(opts domain.ObjectListOptions) ([]*domain.Object, error)
	Delete(id string) error
}

// NewService creates a new service instance
func NewService(projectRepo ProjectRepository, instanceRepo InstanceRepository, metadataRepo MetadataRepository, bucketRepo BucketRepository, objectRepo ObjectRepository) *Service {
	return &Service{
		projectRepo:  projectRepo,
		instanceRepo: instanceRepo,
		metadataRepo: metadataRepo,
		bucketRepo:   bucketRepo,
		objectRepo:   objectRepo,
	}
}

// generateID generates a random hex ID
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// validateProjectName validates a project name
func validateProjectName(name string) error {
	if name == "" {
		return domain.InvalidInputError("project name cannot be empty", nil)
	}
	if len(name) > 255 {
		return domain.InvalidInputError("project name too long", map[string]interface{}{
			"max_length": 255,
			"actual":     len(name),
		})
	}
	// Simple alphanumeric + dash/underscore validation
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
		return domain.InvalidInputError("project name can only contain alphanumeric characters, dashes, and underscores", nil)
	}
	return nil
}

// validateBucketName validates a bucket name
func validateBucketName(name string) error {
	if name == "" {
		return domain.InvalidInputError("bucket name cannot be empty", nil)
	}
	if len(name) > 255 {
		return domain.InvalidInputError("bucket name too long", map[string]interface{}{
			"max_length": 255,
			"actual":     len(name),
		})
	}
	// Simple alphanumeric + dash/underscore validation
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
		return domain.InvalidInputError("bucket name can only contain alphanumeric characters, dashes, and underscores", nil)
	}
	return nil
}

// validateInstanceName validates an instance name
func validateInstanceName(name string) error {
	if name == "" {
		return domain.InvalidInputError("instance name cannot be empty", nil)
	}
	if len(name) > 255 {
		return domain.InvalidInputError("instance name too long", map[string]interface{}{
			"max_length": 255,
			"actual":     len(name),
		})
	}
	// Simple alphanumeric + dash/underscore validation
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
		return domain.InvalidInputError("instance name can only contain alphanumeric characters, dashes, and underscores", nil)
	}
	return nil
}

// validateInstanceSpecs validates instance specifications
func validateInstanceSpecs(cpu int, memoryMB int, image string) error {
	if cpu <= 0 {
		return domain.InvalidInputError("CPU must be positive", map[string]interface{}{"cpu": cpu})
	}
	if cpu > 64 {
		return domain.InvalidInputError("CPU too high", map[string]interface{}{
			"max_cpu": 64,
			"actual":  cpu,
		})
	}
	if memoryMB <= 0 {
		return domain.InvalidInputError("memory must be positive", map[string]interface{}{"memory_mb": memoryMB})
	}
	if memoryMB > 512*1024 { // 512GB
		return domain.InvalidInputError("memory too high", map[string]interface{}{
			"max_memory_mb": 512 * 1024,
			"actual":        memoryMB,
		})
	}
	if image == "" {
		return domain.InvalidInputError("image cannot be empty", nil)
	}
	if len(image) > 255 {
		return domain.InvalidInputError("image name too long", map[string]interface{}{
			"max_length": 255,
			"actual":     len(image),
		})
	}
	return nil
}

// validateInstanceStatus validates instance status
func validateInstanceStatus(status string) error {
	if status != domain.StatusRunning && status != domain.StatusStopped {
		return domain.InvalidInputError("invalid status", map[string]interface{}{
			"valid_statuses": []string{domain.StatusRunning, domain.StatusStopped},
			"actual":         status,
		})
	}
	return nil
}

// validateObjectPath validates an object path
func validateObjectPath(path string) error {
	if path == "" {
		return domain.InvalidInputError("object path cannot be empty", nil)
	}
	if len(path) > 1024 {
		return domain.InvalidInputError("object path too long", map[string]interface{}{"max_length": 1024, "actual": len(path)})
	}
	return nil
}

// Project operations

// CreateProject creates a new project
func (s *Service) CreateProject(req domain.CreateProjectRequest) (*domain.Project, error) {
	if err := validateProjectName(req.Name); err != nil {
		return nil, err
	}

	id, err := generateID()
	if err != nil {
		return nil, domain.InternalError("failed to generate ID")
	}

	project := &domain.Project{
		ID:   id,
		Name: req.Name,
	}

	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject retrieves a project by ID
func (s *Service) GetProject(id string) (*domain.Project, error) {
	return s.projectRepo.GetByID(id)
}

// ListProjects lists projects with optional filtering
func (s *Service) ListProjects(opts domain.ProjectListOptions) ([]*domain.Project, error) {
	return s.projectRepo.List(opts)
}

// UpdateProject updates an existing project
func (s *Service) UpdateProject(id string, req domain.UpdateProjectRequest) (*domain.Project, error) {
	if err := validateProjectName(req.Name); err != nil {
		return nil, err
	}

	return s.projectRepo.Update(id, req)
}

// DeleteProject deletes a project
func (s *Service) DeleteProject(id string) error {
	return s.projectRepo.Delete(id)
}

// Instance operations

// CreateInstance creates a new instance
func (s *Service) CreateInstance(req domain.CreateInstanceRequest) (*domain.Instance, error) {
	if err := validateInstanceName(req.Name); err != nil {
		return nil, err
	}

	if err := validateInstanceSpecs(req.CPU, req.MemoryMB, req.Image); err != nil {
		return nil, err
	}

	status := req.Status
	if status == "" {
		status = domain.StatusRunning
	}
	if err := validateInstanceStatus(status); err != nil {
		return nil, err
	}

	// Verify project exists
	_, err := s.projectRepo.GetByID(req.ProjectID)
	if err != nil {
		if domain.IsNotFound(err) {
			return nil, domain.ForeignKeyViolationError("project", "id", req.ProjectID)
		}
		return nil, err
	}

	id, err := generateID()
	if err != nil {
		return nil, domain.InternalError("failed to generate ID")
	}

	instance := &domain.Instance{
		ID:        id,
		ProjectID: req.ProjectID,
		Name:      req.Name,
		CPU:       req.CPU,
		MemoryMB:  req.MemoryMB,
		Image:     req.Image,
		Status:    status,
	}

	if err := s.instanceRepo.Create(instance); err != nil {
		return nil, err
	}

	return instance, nil
}

// GetInstance retrieves an instance by ID
func (s *Service) GetInstance(id string) (*domain.Instance, error) {
	return s.instanceRepo.GetByID(id)
}

// ListInstances lists instances with optional filtering
func (s *Service) ListInstances(opts domain.InstanceListOptions) ([]*domain.Instance, error) {
	return s.instanceRepo.List(opts)
}

// UpdateInstance updates an existing instance
func (s *Service) UpdateInstance(id string, req domain.UpdateInstanceRequest) (*domain.Instance, error) {
	// Get current instance to check for immutable field changes
	current, err := s.instanceRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Image changes require instance recreation - only error if actually changing
	if req.Image != nil && *req.Image != current.Image {
		return nil, domain.InvalidInputError(
			"Cannot change instance image from '"+current.Image+"' to '"+*req.Image+"'. The image field is immutable - you must destroy and recreate the instance to change the image.",
			map[string]interface{}{
				"field":         "image",
				"current_value": current.Image,
				"requested_value": *req.Image,
				"solution":      "Remove and re-add the resource, or use 'terraform taint' to force recreation",
			},
		)
	}

	if req.Name != nil {
		if err := validateInstanceName(*req.Name); err != nil {
			return nil, err
		}
	}

	if req.CPU != nil || req.MemoryMB != nil {
		// Validate complete specs using current values as defaults
		cpu := current.CPU
		memory := current.MemoryMB
		image := current.Image

		if req.CPU != nil {
			cpu = *req.CPU
		}
		if req.MemoryMB != nil {
			memory = *req.MemoryMB
		}

		if err := validateInstanceSpecs(cpu, memory, image); err != nil {
			return nil, err
		}
	}

	if req.Status != nil {
		if err := validateInstanceStatus(*req.Status); err != nil {
			return nil, err
		}
	}

	return s.instanceRepo.Update(id, req)
}

// DeleteInstance deletes an instance
func (s *Service) DeleteInstance(id string) error {
	return s.instanceRepo.Delete(id)
}

// Metadata operations

// CreateMetadata creates new metadata
func (s *Service) CreateMetadata(req domain.CreateMetadataRequest) (*domain.Metadata, error) {
	if req.Path == "" {
		return nil, domain.InvalidInputError("metadata path cannot be empty", nil)
	}

	return s.metadataRepo.Create(req)
}

// GetMetadata retrieves metadata by ID
func (s *Service) GetMetadata(id string) (*domain.Metadata, error) {
	if id == "" {
		return nil, domain.InvalidInputError("metadata ID cannot be empty", nil)
	}

	return s.metadataRepo.GetByID(id)
}

// UpdateMetadata updates existing metadata
func (s *Service) UpdateMetadata(id string, req domain.UpdateMetadataRequest) (*domain.Metadata, error) {
	if id == "" {
		return nil, domain.InvalidInputError("metadata ID cannot be empty", nil)
	}

	return s.metadataRepo.Update(id, req)
}

// ListMetadata lists metadata with optional prefix filtering
func (s *Service) ListMetadata(opts domain.MetadataListOptions) ([]*domain.Metadata, error) {
	return s.metadataRepo.List(opts)
}

// DeleteMetadata deletes metadata by ID
func (s *Service) DeleteMetadata(id string) error {
	if id == "" {
		return domain.InvalidInputError("metadata ID cannot be empty", nil)
	}

	return s.metadataRepo.Delete(id)
}

// Bucket operations

// CreateBucket creates a new bucket
// Bucket IDs are now equal to their names to provide a stable, user-defined identifier.
func (s *Service) CreateBucket(req domain.CreateBucketRequest) (*domain.Bucket, error) {
	if err := validateBucketName(req.Name); err != nil {
		return nil, err
	}
	// Use name as the stable identifier (ID)
	b := &domain.Bucket{ID: req.Name, Name: req.Name}
	if err := s.bucketRepo.Create(b); err != nil {
		return nil, err
	}
	return b, nil
}

// GetBucket retrieves a bucket by ID (or name, if name is the identifier)
func (s *Service) GetBucket(id string) (*domain.Bucket, error) {
	return s.bucketRepo.GetByID(id)
}

// GetBucketByName retrieves a bucket by name
func (s *Service) GetBucketByName(name string) (*domain.Bucket, error) {
	return s.bucketRepo.GetByName(name)
}

// ListBuckets lists buckets with optional filtering
func (s *Service) ListBuckets(opts domain.BucketListOptions) ([]*domain.Bucket, error) {
	return s.bucketRepo.List(opts)
}

// UpdateBucket updates an existing bucket
// With IDs equal to names, bucket name is immutable. Attempting to change it will return an error.
func (s *Service) UpdateBucket(id string, req domain.UpdateBucketRequest) (*domain.Bucket, error) {
	if err := validateBucketName(req.Name); err != nil {
		return nil, err
	}
	// Get current bucket to enforce immutability
	current, err := s.bucketRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != current.Name {
		return nil, domain.InvalidInputError(
			"Cannot change bucket name from '"+current.Name+"' to '"+req.Name+"'. The name is immutable because it is used as the bucket ID. Destroy and recreate the bucket to change the name.",
			map[string]interface{}{
				"field":           "name",
				"current_value":   current.Name,
				"requested_value": req.Name,
				"solution":        "Remove and re-add the resource, or recreate the bucket with the desired name",
			},
		)
	}
	// No-op update (name unchanged)
	return current, nil
}

// DeleteBucket deletes a bucket
func (s *Service) DeleteBucket(id string) error {
	return s.bucketRepo.Delete(id)
}

// Object operations

// CreateObject creates a new object under a bucket
func (s *Service) CreateObject(req domain.CreateObjectRequest) (*domain.Object, error) {
	if err := validateObjectPath(req.Path); err != nil {
		return nil, err
	}
	if req.BucketID == "" {
		return nil, domain.InvalidInputError("bucket_id cannot be empty", nil)
	}
	if req.Content == "" {
		return nil, domain.InvalidInputError("content cannot be empty", nil)
	}
	// Verify bucket exists
	if _, err := s.bucketRepo.GetByID(req.BucketID); err != nil {
		if domain.IsNotFound(err) {
			return nil, domain.ForeignKeyViolationError("bucket", "id", req.BucketID)
		}
		return nil, err
	}
	return s.objectRepo.Create(req)
}

// GetObject retrieves an object by ID
func (s *Service) GetObject(id string) (*domain.Object, error) {
	return s.objectRepo.GetByID(id)
}

// ListObjects lists objects with optional filtering
func (s *Service) ListObjects(opts domain.ObjectListOptions) ([]*domain.Object, error) {
	return s.objectRepo.List(opts)
}

// UpdateObject updates an existing object
func (s *Service) UpdateObject(id string, req domain.UpdateObjectRequest) (*domain.Object, error) {
	if req.Path != nil {
		if err := validateObjectPath(*req.Path); err != nil {
			return nil, err
		}
	}
	return s.objectRepo.Update(id, req)
}

// DeleteObject deletes an object
func (s *Service) DeleteObject(id string) error {
	return s.objectRepo.Delete(id)
}
