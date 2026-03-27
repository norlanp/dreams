package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDream_JSONMarshaling(t *testing.T) {
	now := time.Now().UTC()
	dream := Dream{
		ID:        123,
		Content:   "Test dream content",
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(dream)
	if err != nil {
		t.Fatalf("failed to marshal dream: %v", err)
	}

	var unmarshaled Dream
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal dream: %v", err)
	}

	if unmarshaled.ID != dream.ID {
		t.Errorf("expected ID %d, got %d", dream.ID, unmarshaled.ID)
	}
	if unmarshaled.Content != dream.Content {
		t.Errorf("expected content %q, got %q", dream.Content, unmarshaled.Content)
	}
}

func TestCluster_TopTermsJSON(t *testing.T) {
	cluster := Cluster{
		ID:         1,
		AnalysisID: 100,
		ClusterID:  5,
		DreamCount: 10,
		TopTerms:   []string{"flying", "water", "house"},
		DreamIDs:   []int64{1, 2, 3},
	}

	jsonStr, err := cluster.TopTermsJSON()
	if err != nil {
		t.Fatalf("failed to marshal top terms: %v", err)
	}

	var terms []string
	err = json.Unmarshal([]byte(jsonStr), &terms)
	if err != nil {
		t.Fatalf("failed to unmarshal top terms: %v", err)
	}

	if len(terms) != len(cluster.TopTerms) {
		t.Fatalf("expected %d terms, got %d", len(cluster.TopTerms), len(terms))
	}

	for i, term := range terms {
		if term != cluster.TopTerms[i] {
			t.Errorf("expected term %q at index %d, got %q", cluster.TopTerms[i], i, term)
		}
	}
}

func TestCluster_DreamIDsJSON(t *testing.T) {
	cluster := Cluster{
		ID:         1,
		AnalysisID: 100,
		ClusterID:  5,
		DreamCount: 10,
		TopTerms:   []string{"flying"},
		DreamIDs:   []int64{1, 2, 3, 4, 5},
	}

	jsonStr, err := cluster.DreamIDsJSON()
	if err != nil {
		t.Fatalf("failed to marshal dream IDs: %v", err)
	}

	var ids []int64
	err = json.Unmarshal([]byte(jsonStr), &ids)
	if err != nil {
		t.Fatalf("failed to unmarshal dream IDs: %v", err)
	}

	if len(ids) != len(cluster.DreamIDs) {
		t.Fatalf("expected %d IDs, got %d", len(cluster.DreamIDs), len(ids))
	}

	for i, id := range ids {
		if id != cluster.DreamIDs[i] {
			t.Errorf("expected ID %d at index %d, got %d", cluster.DreamIDs[i], i, id)
		}
	}
}

func TestCluster_SetTopTermsFromJSON(t *testing.T) {
	cluster := &Cluster{}
	jsonStr := `["term1", "term2", "term3"]`

	err := cluster.SetTopTermsFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("failed to set top terms: %v", err)
	}

	expected := []string{"term1", "term2", "term3"}
	if len(cluster.TopTerms) != len(expected) {
		t.Fatalf("expected %d terms, got %d", len(expected), len(cluster.TopTerms))
	}

	for i, term := range cluster.TopTerms {
		if term != expected[i] {
			t.Errorf("expected term %q at index %d, got %q", expected[i], i, term)
		}
	}
}

func TestCluster_SetDreamIDsFromJSON(t *testing.T) {
	cluster := &Cluster{}
	jsonStr := `[10, 20, 30, 40]`

	err := cluster.SetDreamIDsFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("failed to set dream IDs: %v", err)
	}

	expected := []int64{10, 20, 30, 40}
	if len(cluster.DreamIDs) != len(expected) {
		t.Fatalf("expected %d IDs, got %d", len(expected), len(cluster.DreamIDs))
	}

	for i, id := range cluster.DreamIDs {
		if id != expected[i] {
			t.Errorf("expected ID %d at index %d, got %d", expected[i], i, id)
		}
	}
}

func TestCluster_SetTopTermsFromJSON_Invalid(t *testing.T) {
	cluster := &Cluster{}
	invalidJSON := `["term1", "term2"`

	err := cluster.SetTopTermsFromJSON(invalidJSON)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestCluster_SetDreamIDsFromJSON_Invalid(t *testing.T) {
	cluster := &Cluster{}
	invalidJSON := `[10, 20, 30`

	err := cluster.SetDreamIDsFromJSON(invalidJSON)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestAnalysis_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC()
	analysis := Analysis{
		ID:           42,
		AnalysisDate: now,
		DreamCount:   100,
		NClusters:    5,
		ResultsJSON:  `{"key": "value"}`,
		CreatedAt:    now,
	}

	data, err := json.Marshal(analysis)
	if err != nil {
		t.Fatalf("failed to marshal analysis: %v", err)
	}

	var unmarshaled Analysis
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal analysis: %v", err)
	}

	if unmarshaled.ID != analysis.ID {
		t.Errorf("expected ID %d, got %d", analysis.ID, unmarshaled.ID)
	}
	if unmarshaled.DreamCount != analysis.DreamCount {
		t.Errorf("expected DreamCount %d, got %d", analysis.DreamCount, unmarshaled.DreamCount)
	}
	if unmarshaled.NClusters != analysis.NClusters {
		t.Errorf("expected NClusters %d, got %d", analysis.NClusters, unmarshaled.NClusters)
	}
}
