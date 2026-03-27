package storage

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"dreams/internal/model"
	"dreams/internal/storage/sqlc"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

const maxDreamContentLength = 100000 // 100KB limit

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

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

	source, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
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
	if err := validateDreamContent(content); err != nil {
		return nil, err
	}

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

func validateDreamContent(content string) error {
	if len(content) == 0 {
		return fmt.Errorf("dream content cannot be empty")
	}
	if len(content) > maxDreamContentLength {
		return fmt.Errorf("dream content exceeds maximum length of %d bytes", maxDreamContentLength)
	}
	return nil
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

func (r *Repository) CountDreams(ctx context.Context) (int64, error) {
	count, err := r.queries.CountDreams(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count dreams: %w", err)
	}
	return count, nil
}

func (r *Repository) GetRandomDream(ctx context.Context) (*model.Dream, error) {
	d, err := r.queries.GetRandomDream(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get random dream: %w", err)
	}
	return toModel(d), nil
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
	if err := validateDreamContent(content); err != nil {
		return nil, err
	}

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

func (r *Repository) ListPrimingContent(ctx context.Context) ([]model.PrimingContent, error) {
	rows, err := r.queries.ListPrimingContent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list priming content: %w", err)
	}

	content := make([]model.PrimingContent, len(rows))
	for i, row := range rows {
		content[i] = model.PrimingContent{
			ID:        row.ID,
			Source:    row.Source,
			Title:     row.Title,
			Content:   row.Content,
			Category:  row.Category.String,
			URL:       row.Url.String,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		}
	}

	return content, nil
}

func (r *Repository) GetPrimingContentByCategory(ctx context.Context, category string) ([]model.PrimingContent, error) {
	rows, err := r.queries.GetPrimingContentByCategory(ctx, sql.NullString{String: category, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get priming content by category: %w", err)
	}

	content := make([]model.PrimingContent, len(rows))
	for i, row := range rows {
		content[i] = model.PrimingContent{
			ID:        row.ID,
			Source:    row.Source,
			Title:     row.Title,
			Content:   row.Content,
			Category:  row.Category.String,
			URL:       row.Url.String,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		}
	}

	return content, nil
}

func (r *Repository) SeedPrimingContent(ctx context.Context) error {
	count, err := r.queries.CountPrimingContent(ctx)
	if err != nil {
		return fmt.Errorf("failed to check priming content count: %w", err)
	}

	if count > 0 {
		return nil
	}

	return r.InsertDefaultPrimingContent(ctx)
}

func (r *Repository) InsertDefaultPrimingContent(ctx context.Context) error {
	now := sql.NullTime{Time: time.Now().UTC(), Valid: true}

	for _, item := range defaultPrimingContent {
		params := sqlc.InsertPrimingContentParams{
			Source:    item.Source,
			Title:     item.Title,
			Content:   item.Content,
			Category:  sql.NullString{String: item.Category, Valid: true},
			Url:       sql.NullString{String: item.URL, Valid: true},
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := r.queries.InsertPrimingContent(ctx, params); err != nil {
			return fmt.Errorf("failed to insert priming content %s: %w", item.Title, err)
		}
	}

	return nil
}

type primingContentItem struct {
	Source   string
	Title    string
	Content  string
	Category string
	URL      string
}

var defaultPrimingContent = []primingContentItem{
	{
		Source:   "reddit_wiki",
		Title:    "Lucid Dreaming Wiki",
		Category: "beginner",
		URL:      "https://www.reddit.com/r/LucidDreaming/wiki/index",
		Content: `The Lucid Dreaming Wiki is a comprehensive resource covering everything from basic techniques to advanced practices.

Key Topics:
- Reality Testing: Learn to question your waking state throughout the day so it becomes habit in dreams
- Dream Recall: Keep a dream journal and write immediately upon waking, even if just fragments
- MILD (Mnemonic Induction): As you fall asleep, repeat "I will recognize I'm dreaming" while visualizing a recent dream
- WILD (Wake Induced): Transition directly from wakefulness to lucidity while maintaining consciousness
- WBTB (Wake Back to Bed): Wake after 5-6 hours, stay awake briefly, then return to sleep with intention
- Most beginners see results within 2-4 weeks of consistent practice.`,
	},
	{
		Source:   "reddit_faq",
		Title:    "Frequently Asked Questions",
		Category: "beginner",
		URL:      "https://www.reddit.com/r/LucidDreaming/wiki/faq",
		Content: `Common Questions Answered:

Q: How long does it take to have my first lucid dream?
A: Most people report their first lucid dream within 2-6 weeks of consistent practice.

Q: Is lucid dreaming safe?
A: Yes. It's a natural state of consciousness that occurs spontaneously in about 55% of people at least once in their lifetime.

Q: Can I get stuck in a lucid dream?
A: No. Your body will naturally wake or transition to non-lucid sleep.

Q: What's the best technique for beginners?
A: Start with reality testing combined with dream journaling. These foundational practices make other techniques more effective.`,
	},
	{
		Source:   "reddit_beginners_qa",
		Title:    "Beginner Q&A Part 1",
		Category: "beginner",
		URL:      "https://www.reddit.com/r/LucidDreaming/comments/3iplpa/beginners_qa/",
		Content: `Getting Started: Essential First Steps

1. Start a Dream Journal
   Keep it by your bed. Write anything you remember immediately upon waking. Even fragments strengthen recall.

2. Perform Reality Checks
   Ask "Am I dreaming?" 10-20 times daily. Check your hands, count fingers, try to push finger through palm.

3. Set Intent Before Sleep
   As you drift off, firmly intend to recognize when you're dreaming. Visualize yourself becoming lucid.

4. Wake Back to Bed (WBTB)
   Set alarm for 5-6 hours after bedtime. Stay awake 15-30 minutes, then return to sleep with strong intention.

5. Be Patient and Consistent
   Results compound. Missing one day isn't failure—just resume the next.`,
	},
	{
		Source:   "reddit_beginners_faq_extended",
		Title:    "Beginner FAQ Extended Part 2",
		Category: "beginner",
		URL:      "https://www.reddit.com/r/LucidDreaming/comments/4cpb6o/beginners_faq_extended/",
		Content: `Advanced Beginner Concepts

Dream Signs:
Pay attention to recurring elements in your dreams—these become triggers for lucidity. Common signs: impossible places, seeing deceased people, impossible physics, old homes/schools.

Stabilization Techniques:
When you become lucid, the dream may start fading. To stabilize:
- Rub your hands together
- Spin around slowly
- Touch objects in the dream
- Remind yourself "I'm dreaming" calmly

False Awakenings:
You may "wake up" within a dream. Always do a reality check upon waking! This is a prime opportunity for lucidity.

Sleep Paralysis:
Sometimes occurs during WILD attempts. It's harmless but can be frightening. Stay calm, focus on breathing, know it will pass.`,
	},
	{
		Source:   "reddit_myths",
		Title:    "Myths and Misconceptions",
		Category: "education",
		URL:      "https://www.reddit.com/r/LucidDreaming/comments/2o22rm/myths_and_misconceptions_about_lucid_dreaming/",
		Content: `Debunking Common Myths

MYTH: Lucid dreaming is unnatural or dangerous.
TRUTH: It's a well-documented, natural state studied at Stanford and other institutions.

MYTH: You need supplements or drugs to lucid dream.
TRUTH: While some supplements may help, they're not necessary. Most achieve lucidity through mental techniques alone.

MYTH: Lucid dreaming makes you tired.
TRUTH: You get the same rest. Lucid dreams occur during REM sleep, part of normal sleep architecture.

MYTH: You can practice skills in lucid dreams.
TRUTH: Research shows motor skills can actually be improved through mental rehearsal in lucid dreams.

MYTH: Everyone can lucid dream easily.
TRUTH: While most people can learn, it requires consistent practice and varies by individual.`,
	},
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
