// Package environment คือที่เดียวที่อ่าน environment variable
// ค่าจะเปลี่ยนตาม MODE: dev (ค่าเริ่มต้น) / uat / production
package environment

import "os"

func GetPort() string {
	if v := os.Getenv("PORT"); v != "" {
		return v
	}
	return "8080"
}

func GetMongoURLEnvironment() string {
	switch os.Getenv("MODE") {
	case "uat":
		return os.Getenv("MONGODB_URI_UAT")
	case "production":
		return os.Getenv("MONGODB_URI_PROD")
	default:
		return os.Getenv("MONGODB_URI")
	}
}

func GetDatabaseNameEnvironment() string {
	switch os.Getenv("MODE") {
	case "uat":
		return os.Getenv("DATABASE_NAME_UAT")
	case "production":
		return os.Getenv("DATABASE_NAME_PROD")
	default:
		return os.Getenv("DATABASE_NAME")
	}
}
