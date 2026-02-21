package repository

import (
	"context"

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
