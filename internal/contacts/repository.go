package contacts

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		collection: db.Collection("contacts"),
	}
}

func (r *Repository) Create(ctx context.Context, contact *Contact) error {
	contact.CreatedAt = time.Now()
	contact.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, contact)
	return err
}

func (r *Repository) FindByUserID(ctx context.Context, userID string) ([]Contact, error) {
	var contacts []Contact
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &contacts); err != nil {
		return nil, err
	}

	return contacts, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Contact, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var contact Contact
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

func (r *Repository) Update(ctx context.Context, id string, contact *Contact) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	contact.UpdatedAt = time.Now()

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": contact})
	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
