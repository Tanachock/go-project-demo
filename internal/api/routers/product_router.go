package routers

import (
	"errors"
	"log/slog"
	"net/http"

	"go-crud-demo/models"

	"github.com/labstack/echo/v4"
)

// POST /products
func (ar APIRouter) CreateProduct(c echo.Context) error {
	var request models.ProductRequest
	if err := c.Bind(&request); err != nil {
		return badRequest(c, "invalid request body")
	}
	if err := c.Validate(&request); err != nil {
		return badRequest(c, err.Error())
	}

	product, err := ar.ProductService.CreateProduct(c.Request().Context(), request)
	if err != nil {
		return handleError(c, "CreateProduct", err)
	}
	return c.JSON(http.StatusCreated, product)
}

// GET /products
func (ar APIRouter) GetAllProducts(c echo.Context) error {
	products, err := ar.ProductService.GetAllProducts(c.Request().Context())
	if err != nil {
		return handleError(c, "GetAllProducts", err)
	}
	return c.JSON(http.StatusOK, products)
}

// GET /products/:id
func (ar APIRouter) GetProductByID(c echo.Context) error {
	product, err := ar.ProductService.GetProductByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return handleError(c, "GetProductByID", err)
	}
	return c.JSON(http.StatusOK, product)
}

// PUT /products/:id
func (ar APIRouter) UpdateProduct(c echo.Context) error {
	var request models.ProductRequest
	if err := c.Bind(&request); err != nil {
		return badRequest(c, "invalid request body")
	}
	if err := c.Validate(&request); err != nil {
		return badRequest(c, err.Error())
	}

	product, err := ar.ProductService.UpdateProduct(c.Request().Context(), c.Param("id"), request)
	if err != nil {
		return handleError(c, "UpdateProduct", err)
	}
	return c.JSON(http.StatusOK, product)
}

// DELETE /products/:id
func (ar APIRouter) DeleteProduct(c echo.Context) error {
	if err := ar.ProductService.DeleteProduct(c.Request().Context(), c.Param("id")); err != nil {
		return handleError(c, "DeleteProduct", err)
	}
	return c.NoContent(http.StatusNoContent)
}

// handleError แปลง error เป็น HTTP status
// error ภายใน (เช่น DB ล่ม) → log ไว้ แต่ตอบ client แค่ "internal error"
func handleError(c echo.Context, action string, err error) error {
	if errors.Is(err, models.ErrNotFound) {
		return c.JSON(http.StatusNotFound, models.GenericResponse{Code: http.StatusNotFound, Message: err.Error()})
	}
	slog.Error(action+" failed", "error", err)
	return c.JSON(http.StatusInternalServerError, models.GenericResponse{Code: http.StatusInternalServerError, Message: "internal error"})
}

func badRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, models.GenericResponse{Code: http.StatusBadRequest, Message: message})
}
