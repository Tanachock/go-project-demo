package environment

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadProjectEnv โหลด .env จาก root ของโปรเจกต์ (โฟลเดอร์ที่มี go.mod)
// ทำให้รันได้ทั้งจาก root (go run ./cmd/cmd-api) และจากใน cmd/cmd-api (go run .)
func LoadProjectEnv() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := findProjectRoot(wd)
	if err != nil {
		return err
	}
	return godotenv.Load(filepath.Join(root, ".env"))
}

// findProjectRoot ไล่ขึ้นไปทีละโฟลเดอร์จนกว่าจะเจอ go.mod
func findProjectRoot(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir { // ถึง root ของเครื่องแล้ว ยังไม่เจอ
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}
