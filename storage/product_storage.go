// Package storage คือ layer ที่คุยกับ MongoDB โดยตรง (มีแค่ query ไม่มี business logic)
package storage

import (
	"context"
	"errors"

	"go-crud-demo/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductStorage struct {
	ProductCollection *mongo.Collection
}

// NewProductStorage — Constructor: รับ db แล้วเลือก collection
func NewProductStorage(db *mongo.Database) *ProductStorage {
	return &ProductStorage{
		ProductCollection: db.Collection("products"),
	}
}

func (ps ProductStorage) Insert(ctx context.Context, product models.Product) (models.Product, error) {
	product.ID = primitive.NewObjectID().Hex()
	if _, err := ps.ProductCollection.InsertOne(ctx, product); err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (ps ProductStorage) FindAll(ctx context.Context) ([]models.Product, error) {
	cursor, err := ps.ProductCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	products := []models.Product{}
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (ps ProductStorage) FindByID(ctx context.Context, id string) (models.Product, error) {
	var product models.Product
	err := ps.ProductCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Product{}, models.ErrNotFound
	}
	return product, err
}

func (ps ProductStorage) Update(ctx context.Context, product models.Product) error {
	result, err := ps.ProductCollection.ReplaceOne(ctx, bson.M{"_id": product.ID}, product)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (ps ProductStorage) Delete(ctx context.Context, id string) error {
	result, err := ps.ProductCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}
