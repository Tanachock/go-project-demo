// Package middleware สร้าง echo instance ที่ตั้งค่า middleware พื้นฐานไว้แล้ว
package middleware

import (
	"net/http"

	"go-crud-demo/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func EchoMiddleware() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())    // log ทุก request
	e.Use(middleware.Recover())   // panic → 500 แทนการทำให้ server ล่ม
	e.Use(middleware.RequestID()) // ใส่ X-Request-Id ให้ทุก request
	e.Use(middleware.CORS())

	// ตอบ 404 / 405 ในรูปแบบ GenericResponse เหมือน endpoint อื่น
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok {
			switch he.Code {
			case http.StatusNotFound:
				_ = c.JSON(http.StatusNotFound, models.GenericResponse{Code: 404, Message: "Not Found"})
				return
			case http.StatusMethodNotAllowed:
				_ = c.JSON(http.StatusMethodNotAllowed, models.GenericResponse{Code: 405, Message: "Method Not Allowed"})
				return
			}
		}
		e.DefaultHTTPErrorHandler(err, c)
	}

	return e
}
