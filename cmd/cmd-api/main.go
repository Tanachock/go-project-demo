// main.go คือจุดเริ่มต้นของแอป และเป็นที่เดียวที่ประกอบทุก layer เข้าด้วยกัน
package main

import (
	"log/slog"
	"os"

	"go-crud-demo/internal/api/routers"
	"go-crud-demo/internal/api/services"
	"go-crud-demo/internal/middleware"
	"go-crud-demo/internal/pkg/connector"
	"go-crud-demo/internal/pkg/environment"
	"go-crud-demo/storage"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 1) โหลด .env (ถ้าไม่มีไฟล์ จะใช้ env ของเครื่องแทน)
	if err := environment.LoadProjectEnv(); err != nil {
		slog.Warn("no .env file, using system environment", "error", err)
	}
	if os.Getenv("MODE") == "" {
		os.Setenv("MODE", "dev")
	}
	slog.Info("starting", "mode", os.Getenv("MODE"))

	// 2) Connector: ต่อ MongoDB และ ping ให้แน่ใจว่าใช้ได้
	mongoURI := environment.GetMongoURLEnvironment()
	if mongoURI == "" {
		slog.Error("MongoDB URI is empty: set MONGODB_URI (or _UAT / _PROD) in .env")
		os.Exit(1)
	}
	mongodb, err := connector.NewMongoDBClient(mongoURI, 100)
	if err != nil {
		slog.Error("connect mongodb failed", "error", err)
		os.Exit(1)
	}
	defer mongodb.Disconnect()
	if err := mongodb.Ping(); err != nil {
		slog.Error("ping mongodb failed", "error", err)
		os.Exit(1)
	}
	db := mongodb.SelectDB(environment.GetDatabaseNameEnvironment())

	// 3) Storage: สร้างจาก db
	productStorage := storage.NewProductStorage(db)

	// 4) Service: ส่ง storage เข้าไป (Dependency Injection)
	productService := services.NewProductService(productStorage)

	// 5) Router: ผูก route กับ service
	e := middleware.EchoMiddleware()
	routers.NewAPIRouter(e, productService)

	// 6) Start server
	if err := e.Start(":" + environment.GetPort()); err != nil {
		slog.Error("server stopped", "error", err)
	}
}
