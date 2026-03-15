package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"dreams/internal/model"
	"dreams/internal/storage/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	r := &Repository{
		db:      db,
		queries: sqlc.New(db),
	}

	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return r, nil
}

func (r *Repository) migrate() error {
	driver, err := sqlite.WithInstance(r.db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/storage/migrations",
		"sqlite",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) CreateDream(ctx context.Context, content string) (*model.Dream, error) {
	now := sql.NullTime{Time: time.Now().UTC(), Valid: true}
	params := sqlc.CreateDreamParams{
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	d, err := r.queries.CreateDream(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create dream: %w", err)
	}

	return toModel(d), nil
}

func (r *Repository) ListDreams(ctx context.Context) ([]model.Dream, error) {
	rows, err := r.queries.ListDreams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list dreams: %w", err)
	}

	dreams := make([]model.Dream, len(rows))
	for i, d := range rows {
		dreams[i] = *toModel(d)
	}

	return dreams, nil
}

func (r *Repository) GetDream(ctx context.Context, id int64) (*model.Dream, error) {
	d, err := r.queries.GetDream(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get dream: %w", err)
	}

	return toModel(d), nil
}

func (r *Repository) UpdateDream(ctx context.Context, id int64, content string) (*model.Dream, error) {
	params := sqlc.UpdateDreamParams{
		ID:        id,
		Content:   content,
		UpdatedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
	}

	d, err := r.queries.UpdateDream(ctx, params)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("dream not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update dream: %w", err)
	}

	return toModel(d), nil
}

func (r *Repository) DeleteDream(ctx context.Context, id int64) error {
	err := r.queries.DeleteDream(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete dream: %w", err)
	}
	return nil
}

func toModel(d sqlc.Dream) *model.Dream {
	return &model.Dream{
		ID:        d.ID,
		Content:   d.Content,
		CreatedAt: d.CreatedAt.Time,
		UpdatedAt: d.UpdatedAt.Time,
	}
}
