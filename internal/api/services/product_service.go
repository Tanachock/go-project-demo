package services

import (
	"context"
	"strings"

	"go-crud-demo/models"
)

func (ps ProductService) CreateProduct(ctx context.Context, request models.ProductRequest) (models.Product, error) {
	product := models.Product{
		Name:  strings.TrimSpace(request.Name),
		Price: request.Price,
	}
	return ps.ProductStorage.Insert(ctx, product)
}

func (ps ProductService) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	return ps.ProductStorage.FindAll(ctx)
}

func (ps ProductService) GetProductByID(ctx context.Context, id string) (models.Product, error) {
	return ps.ProductStorage.FindByID(ctx, id)
}

func (ps ProductService) UpdateProduct(ctx context.Context, id string, request models.ProductRequest) (models.Product, error) {
	product := models.Product{
		ID:    id,
		Name:  strings.TrimSpace(request.Name),
		Price: request.Price,
	}
	if err := ps.ProductStorage.Update(ctx, product); err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (ps ProductService) DeleteProduct(ctx context.Context, id string) error {
	return ps.ProductStorage.Delete(ctx, id)
}
