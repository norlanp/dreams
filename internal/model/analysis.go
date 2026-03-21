package model

import (
	"encoding/json"
	"time"
)

type Analysis struct {
	ID           int64
	AnalysisDate time.Time
	DreamCount   int64
	NClusters    int64
	ResultsJSON  string
	CreatedAt    time.Time
}

type Cluster struct {
	ID         int64
	AnalysisID int64
	ClusterID  int64
	DreamCount int64
	TopTerms   []string
	DreamIDs   []int64
	CreatedAt  time.Time
}

func (c *Cluster) TopTermsJSON() (string, error) {
	data, err := json.Marshal(c.TopTerms)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Cluster) DreamIDsJSON() (string, error) {
	data, err := json.Marshal(c.DreamIDs)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Cluster) SetTopTermsFromJSON(data string) error {
	return json.Unmarshal([]byte(data), &c.TopTerms)
}

func (c *Cluster) SetDreamIDsFromJSON(data string) error {
	return json.Unmarshal([]byte(data), &c.DreamIDs)
}
