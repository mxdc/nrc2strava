package types

import "github.com/muktihari/fit/profile/filedef"

type Activity struct {
	ID              string            `json:"id"`
	Type            string            `json:"type"`
	AppID           string            `json:"app_id"`
	StartEpochMs    int64             `json:"start_epoch_ms"`
	EndEpochMs      int64             `json:"end_epoch_ms"`
	LastModified    int64             `json:"last_modified"`
	ActiveDuration  int64             `json:"active_duration_ms"`
	Status          string            `json:"status"`
	Session         bool              `json:"session"`
	DeleteIndicator bool              `json:"delete_indicator"`
	Summaries       []Summary         `json:"summaries"`
	Sources         []string          `json:"sources"`
	Tags            map[string]string `json:"tags"`
	ChangeTokens    []string          `json:"change_tokens"`
	MetricTypes     []string          `json:"metric_types"`
	Metrics         []Metric          `json:"metrics"`
	Moments         []Moment          `json:"moments"`
}

type Summary struct {
	Metric  string  `json:"metric"`
	Summary string  `json:"summary"`
	Source  string  `json:"source"`
	AppID   string  `json:"app_id"`
	Value   float64 `json:"value"`
}

type Metric struct {
	Type   string        `json:"type"`
	Unit   string        `json:"unit"`
	Source string        `json:"source"`
	AppID  string        `json:"appId"`
	Values []MetricValue `json:"values"`
}

type MetricValue struct {
	StartEpochMs int64   `json:"start_epoch_ms"`
	EndEpochMs   int64   `json:"end_epoch_ms"`
	Value        float64 `json:"value"`
}

type Moment struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	AppID     string `json:"app_id"`
	Source    string `json:"source"`
}

// PauseInterval represents a pause interval with start and end timestamps
type PauseInterval struct {
	StartEpochMs int64
	EndEpochMs   int64
}

type Run struct {
	Id       string
	Activity *filedef.Activity
}
