package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hypertf/nahcloud/domain"
)

// ObjectRepository handles object data operations
type ObjectRepository struct {
	db *DB
}

// NewObjectRepository creates a new object repository
func NewObjectRepository(db *DB) *ObjectRepository {
	return &ObjectRepository{db: db}
}

// Create creates a new object (assumes bucket existence validated by service)
func (r *ObjectRepository) Create(req domain.CreateObjectRequest) (*domain.Object, error) {
	// Ensure unique path within bucket
	id := uuid.New().String()
	now := time.Now()
	obj := &domain.Object{
		ID:       id,
		BucketID: req.BucketID,
		Path:     req.Path,
		Content:  req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	query := `INSERT INTO objects (id, bucket_id, path, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, obj.ID, obj.BucketID, obj.Path, obj.Content, obj.CreatedAt, obj.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: objects.bucket_id, objects.path") {
			return nil, domain.AlreadyExistsError("object", "path", obj.Path)
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return nil, domain.ForeignKeyViolationError("bucket", "id", obj.BucketID)
		}
		return nil, fmt.Errorf("failed to create object: %w", err)
	}
	return obj, nil
}

// GetByID retrieves an object by ID
func (r *ObjectRepository) GetByID(id string) (*domain.Object, error) {
	obj := &domain.Object{}
	query := `SELECT id, bucket_id, path, content, created_at, updated_at FROM objects WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&obj.ID, &obj.BucketID, &obj.Path, &obj.Content, &obj.CreatedAt, &obj.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NotFoundError("object", id)
		}
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return obj, nil
}

// Update updates an existing object
func (r *ObjectRepository) Update(id string, req domain.UpdateObjectRequest) (*domain.Object, error) {
	obj, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Path != nil {
		obj.Path = *req.Path
	}
	if req.Content != nil {
		obj.Content = *req.Content
	}
	obj.UpdatedAt = time.Now()

	query := `UPDATE objects SET path = ?, content = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.Exec(query, obj.Path, obj.Content, obj.UpdatedAt, id)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: objects.bucket_id, objects.path") {
			return nil, domain.AlreadyExistsError("object", "path", obj.Path)
		}
		return nil, fmt.Errorf("failed to update object: %w", err)
	}
	return obj, nil
}

// List retrieves objects with optional filtering
func (r *ObjectRepository) List(opts domain.ObjectListOptions) ([]*domain.Object, error) {
	var (
		objects []*domain.Object
		args    []interface{}
	)
	query := `SELECT id, bucket_id, path, content, created_at, updated_at FROM objects`
	var conditions []string
	if opts.BucketID != "" {
		conditions = append(conditions, "bucket_id = ?")
		args = append(args, opts.BucketID)
	}
	if opts.Prefix != "" {
		conditions = append(conditions, "path LIKE ?")
		args = append(args, opts.Prefix+"%")
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY path"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		o := &domain.Object{}
		if err := rows.Scan(&o.ID, &o.BucketID, &o.Path, &o.Content, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan object: %w", err)
		}
		objects = append(objects, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating objects: %w", err)
	}
	return objects, nil
}

// Delete deletes an object by ID
func (r *ObjectRepository) Delete(id string) error {
	// Ensure exists
	_, err := r.GetByID(id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM objects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}
