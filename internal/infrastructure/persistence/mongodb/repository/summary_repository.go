package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
)

// MongoSummaryRepository implements repository.MongoSummaryRepository.
type MongoSummaryRepository struct {
	collection *mongo.Collection
}

// NewMongoSummaryRepository creates a new MongoSummaryRepository.
func NewMongoSummaryRepository(db *mongo.Database) *MongoSummaryRepository {
	return &MongoSummaryRepository{
		collection: db.Collection("material_summary"),
	}
}

// GetByMaterialID retrieves the AI-generated summary for a material from MongoDB.
func (r *MongoSummaryRepository) GetByMaterialID(ctx context.Context, materialID string) (*mongoentities.MaterialSummary, error) {
	filter := bson.M{"material_id": materialID}

	var result mongoentities.MaterialSummary
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.NewNotFoundError("material summary")
		}
		return nil, errors.NewDatabaseError("get mongo summary", err)
	}
	return &result, nil
}
