package dto

// GlobalStatsResponse is the API response for global statistics.
type GlobalStatsResponse struct {
	TotalMaterials      int     `json:"total_materials"`
	CompletedProgress   int     `json:"completed_progress"`
	AverageAttemptScore float64 `json:"average_attempt_score"`
}

// MaterialStatsResponse is the API response for a single material's statistics.
type MaterialStatsResponse struct {
	TotalAttempts  int     `json:"total_attempts"`
	AverageScore   float64 `json:"average_score"`
	CompletionRate float64 `json:"completion_rate"`
	UniqueStudents int     `json:"unique_students"`
}
