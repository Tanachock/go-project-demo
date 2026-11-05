// Package services คือ business logic (ห้าม import echo และไม่รู้จัก mongo)
// factory.go เก็บ struct, interface และ constructor ของ service
package services

import (
	"context"

	"go-crud-demo/models"
)

type ProductService struct {
	ProductStorage ProductStore
}

// IProductService คือสิ่งที่ router เห็น (router รู้จักแค่ interface นี้)
type IProductService interface {
	CreateProduct(ctx context.Context, request models.ProductRequest) (models.Product, error)
	GetAllProducts(ctx context.Context) ([]models.Product, error)
	GetProductByID(ctx context.Context, id string) (models.Product, error)
	UpdateProduct(ctx context.Context, id string, request models.ProductRequest) (models.Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

// ProductStore คือสิ่งที่ service ต้องการจาก storage
// อะไรก็ได้ที่มี method ครบ (Mongo, memory, fake ใน test) → ส่งเข้ามาได้หมด
type ProductStore interface {
	Insert(ctx context.Context, product models.Product) (models.Product, error)
	FindAll(ctx context.Context) ([]models.Product, error)
	FindByID(ctx context.Context, id string) (models.Product, error)
	Update(ctx context.Context, product models.Product) error
	Delete(ctx context.Context, id string) error
}

// NewProductService — Constructor + Dependency Injection
// รับ storage ที่สร้างไว้แล้วเข้ามา (ไม่สร้างเอง)
// → service ไม่ต้องรู้ว่าเบื้องหลังเป็น Mongo และ test ส่ง fake เข้ามาได้
func NewProductService(store ProductStore) IProductService {
	return ProductService{
		ProductStorage: store,
	}
}
