package storage

import (
	"context"
	"database/sql"
	"encoding/json"
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

	analysis, err := toAnalysisModel(a)
	if err != nil {
		return nil, fmt.Errorf("failed to map analysis: %w", err)
	}

	return analysis, nil
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

	cluster, err := toClusterModel(c)
	if err != nil {
		return nil, fmt.Errorf("failed to map cluster: %w", err)
	}

	return cluster, nil
}

func (r *Repository) SaveAnalysisWithClusters(ctx context.Context, analysisDate time.Time, dreamCount, nClusters int64, resultsJSON string, clusters []model.Cluster) (*model.Analysis, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start analysis transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)
	analysisRow, err := qtx.CreateAnalysis(ctx, sqlc.CreateAnalysisParams{
		AnalysisDate: analysisDate.Format(time.RFC3339),
		DreamCount:   dreamCount,
		NClusters:    nClusters,
		ResultsJson:  resultsJSON,
		CreatedAt:    sql.NullTime{Time: time.Now().UTC(), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create analysis: %w", err)
	}

	analysis, err := toAnalysisModel(analysisRow)
	if err != nil {
		return nil, fmt.Errorf("failed to map analysis: %w", err)
	}

	for _, cluster := range clusters {
		topTerms, err := json.Marshal(cluster.TopTerms)
		if err != nil {
			return nil, fmt.Errorf("failed to encode cluster top terms: %w", err)
		}

		dreamIDs, err := json.Marshal(cluster.DreamIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to encode cluster dream ids: %w", err)
		}

		_, err = qtx.CreateCluster(ctx, sqlc.CreateClusterParams{
			AnalysisID: analysis.ID,
			ClusterID:  cluster.ClusterID,
			DreamCount: cluster.DreamCount,
			TopTerms:   string(topTerms),
			DreamIds:   string(dreamIDs),
			CreatedAt:  sql.NullTime{Time: time.Now().UTC(), Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create cluster: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit analysis transaction: %w", err)
	}

	committed = true
	return analysis, nil
}

func (r *Repository) GetLatestAnalysis(ctx context.Context) (*model.Analysis, error) {
	a, err := r.queries.GetLatestAnalysis(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest analysis: %w", err)
	}

	analysis, err := toAnalysisModel(a)
	if err != nil {
		return nil, fmt.Errorf("failed to map latest analysis: %w", err)
	}

	return analysis, nil
}

func (r *Repository) GetAnalysisClusters(ctx context.Context, analysisID int64) ([]model.Cluster, error) {
	rows, err := r.queries.GetAnalysisClusters(ctx, analysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis clusters: %w", err)
	}

	clusters := make([]model.Cluster, len(rows))
	for i, c := range rows {
		cluster, err := toClusterModel(c)
		if err != nil {
			return nil, fmt.Errorf("failed to map analysis cluster: %w", err)
		}
		clusters[i] = *cluster
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
		analysis, err := toAnalysisModel(a)
		if err != nil {
			return nil, fmt.Errorf("failed to map analysis history row: %w", err)
		}
		analyses[i] = *analysis
	}

	return analyses, nil
}

func (r *Repository) GetFreshPrimingCache(ctx context.Context, source string, now time.Time, ttl time.Duration) (*model.PrimingCache, error) {
	if ttl <= 0 {
		return nil, fmt.Errorf("ttl must be greater than zero")
	}

	cache, err := r.queries.GetPrimingCache(ctx, source)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get priming cache: %w", err)
	}

	if cache.FetchedAt.Before(now.Add(-ttl)) {
		return nil, nil
	}

	var payload []string
	if err := json.Unmarshal([]byte(cache.PayloadJson), &payload); err != nil {
		return nil, fmt.Errorf("failed to decode priming cache payload: %w", err)
	}

	return &model.PrimingCache{
		Source:    cache.Source,
		Payload:   payload,
		FetchedAt: cache.FetchedAt,
	}, nil
}

func (r *Repository) SavePrimingCache(ctx context.Context, source string, payload []string, fetchedAt time.Time) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode priming cache payload: %w", err)
	}

	err = r.queries.UpsertPrimingCache(ctx, sqlc.UpsertPrimingCacheParams{
		Source:      source,
		PayloadJson: string(data),
		FetchedAt:   fetchedAt,
		UpdatedAt:   sql.NullTime{Time: time.Now().UTC(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to save priming cache: %w", err)
	}

	return nil
}

func (r *Repository) SavePrimingLog(ctx context.Context, source, outcome, detail, content string, createdAt time.Time) error {
	err := r.queries.InsertPrimingLog(ctx, sqlc.InsertPrimingLogParams{
		CreatedAt: createdAt,
		Source:    source,
		Outcome:   outcome,
		Detail:    detail,
		Content:   content,
	})
	if err != nil {
		return fmt.Errorf("failed to save priming log: %w", err)
	}

	return nil
}

func (r *Repository) ListPrimingLogs(ctx context.Context) ([]model.PrimingLog, error) {
	rows, err := r.queries.ListPrimingLogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list priming logs: %w", err)
	}

	logs := make([]model.PrimingLog, len(rows))
	for i, row := range rows {
		logs[i] = model.PrimingLog{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			Source:    row.Source,
			Outcome:   row.Outcome,
			Detail:    row.Detail,
			Content:   row.Content,
		}
	}

	return logs, nil
}

func toAnalysisModel(a sqlc.DreamAnalysis) (*model.Analysis, error) {
	analysisDate, err := time.Parse(time.RFC3339, a.AnalysisDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse analysis timestamp %q: %w", a.AnalysisDate, err)
	}

	return &model.Analysis{
		ID:           a.ID,
		AnalysisDate: analysisDate,
		DreamCount:   a.DreamCount,
		NClusters:    a.NClusters,
		ResultsJSON:  a.ResultsJson,
		CreatedAt:    a.CreatedAt.Time,
	}, nil
}

func toClusterModel(c sqlc.DreamCluster) (*model.Cluster, error) {
	cluster := &model.Cluster{
		ID:         c.ID,
		AnalysisID: c.AnalysisID,
		ClusterID:  c.ClusterID,
		DreamCount: c.DreamCount,
		CreatedAt:  c.CreatedAt.Time,
	}

	err := cluster.SetTopTermsFromJSON(c.TopTerms)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cluster top terms: %w", err)
	}

	err = cluster.SetDreamIDsFromJSON(c.DreamIds)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cluster dream ids: %w", err)
	}

	return cluster, nil
}
