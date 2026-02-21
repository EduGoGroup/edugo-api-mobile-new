package services

import (
	"strings"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
)

// ScoreResult holds the result of grading a single answer.
type ScoreResult struct {
	IsCorrect    bool
	PointsEarned float64
	MaxPoints    float64
}

// GradeAnswer grades a student answer against the correct answer using
// the appropriate scoring strategy for the question type.
func GradeAnswer(question mongoentities.Question, studentAnswer string) ScoreResult {
	maxPoints := float64(question.Points)
	if maxPoints <= 0 {
		maxPoints = 1
	}

	switch question.QuestionType {
	case "multiple_choice":
		return gradeMultipleChoice(question, studentAnswer, maxPoints)
	case "true_false":
		return gradeTrueFalse(question, studentAnswer, maxPoints)
	case "short_answer":
		return gradeShortAnswer(question, studentAnswer, maxPoints)
	default:
		return ScoreResult{
			IsCorrect:    false,
			PointsEarned: 0,
			MaxPoints:    maxPoints,
		}
	}
}

func gradeMultipleChoice(q mongoentities.Question, answer string, maxPoints float64) ScoreResult {
	correct := strings.TrimSpace(strings.ToLower(q.CorrectAnswer))
	given := strings.TrimSpace(strings.ToLower(answer))
	isCorrect := correct == given
	earned := 0.0
	if isCorrect {
		earned = maxPoints
	}
	return ScoreResult{IsCorrect: isCorrect, PointsEarned: earned, MaxPoints: maxPoints}
}

func gradeTrueFalse(q mongoentities.Question, answer string, maxPoints float64) ScoreResult {
	correct := strings.TrimSpace(strings.ToLower(q.CorrectAnswer))
	given := strings.TrimSpace(strings.ToLower(answer))
	isCorrect := correct == given
	earned := 0.0
	if isCorrect {
		earned = maxPoints
	}
	return ScoreResult{IsCorrect: isCorrect, PointsEarned: earned, MaxPoints: maxPoints}
}

func gradeShortAnswer(q mongoentities.Question, answer string, maxPoints float64) ScoreResult {
	correct := strings.TrimSpace(strings.ToLower(q.CorrectAnswer))
	given := strings.TrimSpace(strings.ToLower(answer))
	isCorrect := correct == given
	earned := 0.0
	if isCorrect {
		earned = maxPoints
	}
	return ScoreResult{IsCorrect: isCorrect, PointsEarned: earned, MaxPoints: maxPoints}
}
