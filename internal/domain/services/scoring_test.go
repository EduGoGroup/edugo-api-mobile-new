package services

import (
	"testing"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
	"github.com/stretchr/testify/assert"
)

func TestGradeAnswer(t *testing.T) {
	tests := []struct {
		name            string
		question        mongoentities.Question
		studentAnswer   string
		expectedCorrect bool
		expectedEarned  float64
		expectedMax     float64
	}{
		// --- multiple_choice ---
		{
			name: "multiple_choice correct",
			question: mongoentities.Question{
				QuestionType:  "multiple_choice",
				CorrectAnswer: "B",
				Points:        2,
			},
			studentAnswer:   "B",
			expectedCorrect: true,
			expectedEarned:  2.0,
			expectedMax:     2.0,
		},
		{
			name: "multiple_choice incorrect",
			question: mongoentities.Question{
				QuestionType:  "multiple_choice",
				CorrectAnswer: "B",
				Points:        2,
			},
			studentAnswer:   "A",
			expectedCorrect: false,
			expectedEarned:  0.0,
			expectedMax:     2.0,
		},
		{
			name: "multiple_choice case insensitive",
			question: mongoentities.Question{
				QuestionType:  "multiple_choice",
				CorrectAnswer: "B",
				Points:        1,
			},
			studentAnswer:   "b",
			expectedCorrect: true,
			expectedEarned:  1.0,
			expectedMax:     1.0,
		},
		{
			name: "multiple_choice with whitespace",
			question: mongoentities.Question{
				QuestionType:  "multiple_choice",
				CorrectAnswer: " B ",
				Points:        1,
			},
			studentAnswer:   "B",
			expectedCorrect: true,
			expectedEarned:  1.0,
			expectedMax:     1.0,
		},
		// --- true_false ---
		{
			name: "true_false correct true",
			question: mongoentities.Question{
				QuestionType:  "true_false",
				CorrectAnswer: "true",
				Points:        1,
			},
			studentAnswer:   "true",
			expectedCorrect: true,
			expectedEarned:  1.0,
			expectedMax:     1.0,
		},
		{
			name: "true_false incorrect",
			question: mongoentities.Question{
				QuestionType:  "true_false",
				CorrectAnswer: "true",
				Points:        1,
			},
			studentAnswer:   "false",
			expectedCorrect: false,
			expectedEarned:  0.0,
			expectedMax:     1.0,
		},
		{
			name: "true_false correct false",
			question: mongoentities.Question{
				QuestionType:  "true_false",
				CorrectAnswer: "false",
				Points:        3,
			},
			studentAnswer:   "false",
			expectedCorrect: true,
			expectedEarned:  3.0,
			expectedMax:     3.0,
		},
		{
			name: "true_false case insensitive",
			question: mongoentities.Question{
				QuestionType:  "true_false",
				CorrectAnswer: "True",
				Points:        1,
			},
			studentAnswer:   "true",
			expectedCorrect: true,
			expectedEarned:  1.0,
			expectedMax:     1.0,
		},
		// --- short_answer ---
		{
			name: "short_answer exact match",
			question: mongoentities.Question{
				QuestionType:  "short_answer",
				CorrectAnswer: "photosynthesis",
				Points:        5,
			},
			studentAnswer:   "photosynthesis",
			expectedCorrect: true,
			expectedEarned:  5.0,
			expectedMax:     5.0,
		},
		{
			name: "short_answer case insensitive",
			question: mongoentities.Question{
				QuestionType:  "short_answer",
				CorrectAnswer: "Photosynthesis",
				Points:        5,
			},
			studentAnswer:   "photosynthesis",
			expectedCorrect: true,
			expectedEarned:  5.0,
			expectedMax:     5.0,
		},
		{
			name: "short_answer wrong",
			question: mongoentities.Question{
				QuestionType:  "short_answer",
				CorrectAnswer: "photosynthesis",
				Points:        5,
			},
			studentAnswer:   "respiration",
			expectedCorrect: false,
			expectedEarned:  0.0,
			expectedMax:     5.0,
		},
		{
			name: "short_answer with whitespace trim",
			question: mongoentities.Question{
				QuestionType:  "short_answer",
				CorrectAnswer: " photosynthesis ",
				Points:        2,
			},
			studentAnswer:   "photosynthesis",
			expectedCorrect: true,
			expectedEarned:  2.0,
			expectedMax:     2.0,
		},
		// --- unknown question type ---
		{
			name: "unknown type always incorrect",
			question: mongoentities.Question{
				QuestionType:  "essay",
				CorrectAnswer: "anything",
				Points:        10,
			},
			studentAnswer:   "anything",
			expectedCorrect: false,
			expectedEarned:  0.0,
			expectedMax:     10.0,
		},
		// --- zero points fallback ---
		{
			name: "zero points defaults to 1",
			question: mongoentities.Question{
				QuestionType:  "multiple_choice",
				CorrectAnswer: "A",
				Points:        0,
			},
			studentAnswer:   "A",
			expectedCorrect: true,
			expectedEarned:  1.0,
			expectedMax:     1.0,
		},
		{
			name: "negative points defaults to 1",
			question: mongoentities.Question{
				QuestionType:  "multiple_choice",
				CorrectAnswer: "A",
				Points:        -5,
			},
			studentAnswer:   "A",
			expectedCorrect: true,
			expectedEarned:  1.0,
			expectedMax:     1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GradeAnswer(tt.question, tt.studentAnswer)
			assert.Equal(t, tt.expectedCorrect, result.IsCorrect, "IsCorrect mismatch")
			assert.InDelta(t, tt.expectedEarned, result.PointsEarned, 0.001, "PointsEarned mismatch")
			assert.InDelta(t, tt.expectedMax, result.MaxPoints, 0.001, "MaxPoints mismatch")
		})
	}
}
