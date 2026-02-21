package dto

import "time"

// SummaryResponse is the API response for a material summary from MongoDB.
type SummaryResponse struct {
	MaterialID  string    `json:"material_id"`
	Summary     string    `json:"summary"`
	KeyPoints   []string  `json:"key_points"`
	Language    string    `json:"language"`
	WordCount   int       `json:"word_count"`
	Version     int       `json:"version"`
	AIModel     string    `json:"ai_model"`
	CreatedAt   time.Time `json:"created_at"`
}
