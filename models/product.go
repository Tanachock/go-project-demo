package models

import "errors"

// ErrNotFound ใช้ร่วมกันทุก layer: storage คืนค่า → service ส่งต่อ → router แปลงเป็น 404
var ErrNotFound = errors.New("product not found")

// Product คือข้อมูลที่เก็บใน collection "products"
type Product struct {
	ID    string  `json:"id"    bson:"_id"`
	Name  string  `json:"name"  bson:"name"`
	Price float64 `json:"price" bson:"price"`
}
