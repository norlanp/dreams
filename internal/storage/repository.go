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

func (r *Repository) SearchDreams(ctx context.Context, query string) ([]model.Dream, error) {
	rows, err := r.queries.SearchDreams(ctx, sql.NullString{String: query, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to search dreams: %w", err)
	}

	dreams := make([]model.Dream, len(rows))
	for i, d := range rows {
		dreams[i] = *toModel(d)
	}

	return dreams, nil
}

func toModel(d sqlc.Dream) *model.Dream {
	return &model.Dream{
		ID:        d.ID,
		Content:   d.Content,
		CreatedAt: d.CreatedAt.Time,
		UpdatedAt: d.UpdatedAt.Time,
	}
}

func (r *Repository) SaveAnalysis(ctx context.Context, analysisDate time.Time, dreamCount, nClusters int64, resultsJSON string) (*model.Analysis, error) {
	params := sqlc.CreateAnalysisParams{
		AnalysisDate: analysisDate.Format(time.RFC3339),
		DreamCount:   dreamCount,
		NClusters:    nClusters,
		ResultsJson:  resultsJSON,
		CreatedAt:    sql.NullTime{Time: time.Now().UTC(), Valid: true},
	}

	a, err := r.queries.CreateAnalysis(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create analysis: %w", err)
	}

	return toAnalysisModel(a), nil
}

func (r *Repository) SaveCluster(ctx context.Context, analysisID, clusterID, dreamCount int64, topTerms, dreamIDs string) (*model.Cluster, error) {
	params := sqlc.CreateClusterParams{
		AnalysisID: analysisID,
		ClusterID:  clusterID,
		DreamCount: dreamCount,
		TopTerms:   topTerms,
		DreamIds:   dreamIDs,
		CreatedAt:  sql.NullTime{Time: time.Now().UTC(), Valid: true},
	}

	c, err := r.queries.CreateCluster(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	return toClusterModel(c), nil
}

func (r *Repository) GetLatestAnalysis(ctx context.Context) (*model.Analysis, error) {
	a, err := r.queries.GetLatestAnalysis(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest analysis: %w", err)
	}

	return toAnalysisModel(a), nil
}

func (r *Repository) GetAnalysisClusters(ctx context.Context, analysisID int64) ([]model.Cluster, error) {
	rows, err := r.queries.GetAnalysisClusters(ctx, analysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis clusters: %w", err)
	}

	clusters := make([]model.Cluster, len(rows))
	for i, c := range rows {
		clusters[i] = *toClusterModel(c)
	}

	return clusters, nil
}

func (r *Repository) ListAnalysisHistory(ctx context.Context) ([]model.Analysis, error) {
	rows, err := r.queries.ListAnalysisHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list analysis history: %w", err)
	}

	analyses := make([]model.Analysis, len(rows))
	for i, a := range rows {
		analyses[i] = *toAnalysisModel(a)
	}

	return analyses, nil
}

func toAnalysisModel(a sqlc.DreamAnalysis) *model.Analysis {
	analysisDate, _ := time.Parse(time.RFC3339, a.AnalysisDate)
	return &model.Analysis{
		ID:           a.ID,
		AnalysisDate: analysisDate,
		DreamCount:   a.DreamCount,
		NClusters:    a.NClusters,
		ResultsJSON:  a.ResultsJson,
		CreatedAt:    a.CreatedAt.Time,
	}
}

func toClusterModel(c sqlc.DreamCluster) *model.Cluster {
	return &model.Cluster{
		ID:         c.ID,
		AnalysisID: c.AnalysisID,
		ClusterID:  c.ClusterID,
		DreamCount: c.DreamCount,
		CreatedAt:  c.CreatedAt.Time,
	}
}
