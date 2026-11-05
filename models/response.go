package models

// GenericResponse รูปแบบ response มาตรฐานสำหรับ error และข้อความทั่วไป
type GenericResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
