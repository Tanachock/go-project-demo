// Package routers คือ HTTP layer: bind → validate → เรียก service → แปลง error เป็น HTTP
// factory.go เก็บ struct ของ router และการลงทะเบียน route ทั้งหมด
package routers

import (
	"net/http"

	"go-crud-demo/internal/api/services"
	"go-crud-demo/internal/middleware"
	"go-crud-demo/models"

	"github.com/labstack/echo/v4"
)

type APIRouter struct {
	ProductService services.IProductService
}

// NewAPIRouter — รับ service (interface) แล้วผูก route เข้ากับ echo
func NewAPIRouter(e *echo.Echo, productService services.IProductService) {
	ar := APIRouter{
		ProductService: productService,
	}

	e.Validator = middleware.NewCustomValidator()
	e.GET("/health", ar.HealthCheck)

	gProduct := e.Group("/products")
	gProduct.POST("", ar.CreateProduct)
	gProduct.GET("", ar.GetAllProducts)
	gProduct.GET("/:id", ar.GetProductByID)
	gProduct.PUT("/:id", ar.UpdateProduct)
	gProduct.DELETE("/:id", ar.DeleteProduct)
}

func (ar APIRouter) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, models.GenericResponse{Code: http.StatusOK, Message: "Health Check OK"})
}
