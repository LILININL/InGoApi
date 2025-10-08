package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"fristGoproject/internal/auth"
	"fristGoproject/internal/httpapi/dto"
)

// AuthHandler รับผิดชอบจัดการเส้นทางที่เกี่ยวกับการยืนยันตัวตนทั้งหมด
type AuthHandler struct {
	service *auth.Service
}

// NewAuthHandler คืนค่า handler ที่เชื่อมกับ service เรียบร้อยแล้ว
func NewAuthHandler(service *auth.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register รับคำขอสมัครสมาชิกใหม่และอ่านข้อมูลจาก JSON
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "ไม่อนุญาตให้ใช้เมธอดนี้", http.StatusMethodNotAllowed)
		return
	}

	var body dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "เนื้อหาไม่ใช่ JSON ที่ถูกต้อง", http.StatusBadRequest)
		return
	}

	u, err := h.service.Register(r.Context(), body.Email, body.Password, body.Name)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, auth.ErrEmailInUse) {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusCreated, u)
}

// Login ตรวจสอบอีเมลและรหัสผ่านจากคำขอ
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "ไม่อนุญาตให้ใช้เมธอดนี้", http.StatusMethodNotAllowed)
		return
	}

	var body dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "เนื้อหาไม่ใช่ JSON ที่ถูกต้อง", http.StatusBadRequest)
		return
	}

	u, err := h.service.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, auth.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		http.Error(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusOK, u)
}

// ChangePassword ตรวจสอบรหัสเดิมก่อนเปลี่ยนไปเป็นรหัสใหม่
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "ไม่อนุญาตให้ใช้เมธอดนี้", http.StatusMethodNotAllowed)
		return
	}

	var body dto.ChangePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "เนื้อหาไม่ใช่ JSON ที่ถูกต้อง", http.StatusBadRequest)
		return
	}

	if err := h.service.ChangePassword(r.Context(), body.Email, body.OldPassword, body.NewPassword); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, auth.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		http.Error(w, err.Error(), status)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "เปลี่ยนรหัสผ่านเรียบร้อย",
	})
}

// writeJSON เป็นฟังก์ชันช่วยเขียนผลลัพธ์เป็น JSON พร้อมตั้งค่า header ให้บริการ
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
