package models

// ProductRequest ใช้ทั้งตอนสร้าง (POST) และแก้ไข (PUT)
// tag `validate` ถูกตรวจโดย c.Validate() ใน router (ดู internal/middleware/validator.go)
type ProductRequest struct {
	Name  string  `json:"name"  validate:"required"`
	Price float64 `json:"price" validate:"required,gt=0"`
}
