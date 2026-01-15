package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/hypertf/nahcloud/domain"
)

// BucketRepository handles bucket data operations
type BucketRepository struct {
	db *DB
}

// NewBucketRepository creates a new bucket repository
func NewBucketRepository(db *DB) *BucketRepository {
	return &BucketRepository{db: db}
}

// Create creates a new bucket
func (r *BucketRepository) Create(bucket *domain.Bucket) error {
	now := time.Now()
	bucket.CreatedAt = now
	bucket.UpdatedAt = now

	query := `INSERT INTO buckets (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, bucket.ID, bucket.Name, bucket.CreatedAt, bucket.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: buckets.name") {
			return domain.AlreadyExistsError("bucket", "name", bucket.Name)
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed: buckets.id") || strings.Contains(strings.ToLower(err.Error()), "primary key constraint failed") {
			return domain.AlreadyExistsError("bucket", "id", bucket.ID)
		}
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

// GetByID retrieves a bucket by ID
func (r *BucketRepository) GetByID(id string) (*domain.Bucket, error) {
	bucket := &domain.Bucket{}
	query := `SELECT id, name, created_at, updated_at FROM buckets WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&bucket.ID, &bucket.Name, &bucket.CreatedAt, &bucket.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NotFoundError("bucket", id)
		}
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}
	return bucket, nil
}

// GetByName retrieves a bucket by name
func (r *BucketRepository) GetByName(name string) (*domain.Bucket, error) {
	bucket := &domain.Bucket{}
	query := `SELECT id, name, created_at, updated_at FROM buckets WHERE name = ?`
	err := r.db.QueryRow(query, name).Scan(&bucket.ID, &bucket.Name, &bucket.CreatedAt, &bucket.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NotFoundError("bucket", name)
		}
		return nil, fmt.Errorf("failed to get bucket by name: %w", err)
	}
	return bucket, nil
}

// List retrieves buckets with optional filtering
func (r *BucketRepository) List(opts domain.BucketListOptions) ([]*domain.Bucket, error) {
	var buckets []*domain.Bucket
	var args []interface{}
	query := `SELECT id, name, created_at, updated_at FROM buckets`
	var conditions []string
	if opts.Name != "" {
		conditions = append(conditions, "name = ?")
		args = append(args, opts.Name)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY name"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		b := &domain.Bucket{}
		if err := rows.Scan(&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan bucket: %w", err)
		}
		buckets = append(buckets, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating buckets: %w", err)
	}
	return buckets, nil
}

// Update updates an existing bucket
func (r *BucketRepository) Update(id string, req domain.UpdateBucketRequest) (*domain.Bucket, error) {
	b, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	b.Name = req.Name
	b.UpdatedAt = time.Now()
	query := `UPDATE buckets SET name = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.Exec(query, b.Name, b.UpdatedAt, id)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: buckets.name") {
			return nil, domain.AlreadyExistsError("bucket", "name", b.Name)
		}
		return nil, fmt.Errorf("failed to update bucket: %w", err)
	}
	return b, nil
}

// Delete deletes a bucket by ID (and cascades to delete its objects)
func (r *BucketRepository) Delete(id string) error {
	// Ensure bucket exists
	_, err := r.GetByID(id)
	if err != nil {
		return err
	}
	// Rely on FK ON DELETE CASCADE to remove objects
	_, err = r.db.Exec(`DELETE FROM buckets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	return nil
}
