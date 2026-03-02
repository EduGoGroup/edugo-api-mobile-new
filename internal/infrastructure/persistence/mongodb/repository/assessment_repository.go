package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
)

// MongoAssessmentRepository implements repository.MongoAssessmentRepository.
type MongoAssessmentRepository struct {
	collection *mongo.Collection
}

// NewMongoAssessmentRepository creates a new MongoAssessmentRepository.
func NewMongoAssessmentRepository(db *mongo.Database) *MongoAssessmentRepository {
	return &MongoAssessmentRepository{
		collection: db.Collection("material_assessment_worker"),
	}
}

// GetByMaterialID retrieves the assessment document for a material from MongoDB.
func (r *MongoAssessmentRepository) GetByMaterialID(ctx context.Context, materialID string) (*mongoentities.MaterialAssessment, error) {
	filter := bson.M{"material_id": materialID}

	var result mongoentities.MaterialAssessment
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.NewNotFoundError("material assessment")
		}
		return nil, errors.NewDatabaseError("get mongo assessment", err)
	}
	return &result, nil
}

// GetByObjectID retrieves the assessment document by its MongoDB _id.
func (r *MongoAssessmentRepository) GetByObjectID(ctx context.Context, objectID string) (*mongoentities.MaterialAssessment, error) {
	oid, err := bson.ObjectIDFromHex(objectID)
	if err != nil {
		return nil, errors.NewValidationError("invalid mongo object id")
	}

	var result mongoentities.MaterialAssessment
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.NewNotFoundError("material assessment")
		}
		return nil, errors.NewDatabaseError("get mongo assessment by id", err)
	}
	return &result, nil
}

// Create inserts a new assessment document into MongoDB and returns the ObjectID hex string.
func (r *MongoAssessmentRepository) Create(ctx context.Context, doc *mongoentities.MaterialAssessment) (string, error) {
	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return "", errors.NewDatabaseError("create mongo assessment", err)
	}
	oid, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return "", errors.NewInternalError("unexpected inserted id type", nil)
	}
	return oid.Hex(), nil
}

// ReplaceQuestions replaces the entire questions array and updates metadata.
func (r *MongoAssessmentRepository) ReplaceQuestions(ctx context.Context, objectID string, questions []mongoentities.Question, totalPoints int) error {
	oid, err := bson.ObjectIDFromHex(objectID)
	if err != nil {
		return errors.NewValidationError("invalid mongo object id")
	}

	update := bson.M{
		"$set": bson.M{
			"questions":       questions,
			"total_questions": len(questions),
			"total_points":    totalPoints,
			"updated_at":      time.Now(),
		},
	}

	result, err := r.collection.UpdateByID(ctx, oid, update)
	if err != nil {
		return errors.NewDatabaseError("replace questions", err)
	}
	if result.MatchedCount == 0 {
		return errors.NewNotFoundError("material assessment")
	}
	return nil
}
