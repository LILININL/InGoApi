package httpapi

import (
	"net/http"

	"fristGoproject/internal/user"
)

// UserHandler รวม endpoint ที่เกี่ยวกับข้อมูลผู้ใช้ทั่วไป
type UserHandler struct {
	service *user.Service
}

// NewUserHandler คืน handler ที่เชื่อมกับ user service เรียบร้อยแล้ว
func NewUserHandler(service *user.Service) *UserHandler {
	return &UserHandler{service: service}
}

// List คืนรายการผู้ใช้ทั้งหมดในระบบ
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "ไม่อนุญาตให้ใช้เมธอดนี้", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, users)
}
