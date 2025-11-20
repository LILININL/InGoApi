package httpapi

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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

	email := strings.TrimSpace(body.Email)
	name := strings.TrimSpace(body.Name)
	if email == "" || name == "" {
		http.Error(w, "email และ name ต้องไม่ว่าง", http.StatusBadRequest)
		return
	}
	passwordHex, ok := normalizeSHA256Hex(body.Password)
	if !ok {
		http.Error(w, "password ต้องเป็น SHA-256 hex 64 ตัวอักษร", http.StatusBadRequest)
		return
	}

	u, err := h.service.Register(r.Context(), email, passwordHex, name)
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

	email := strings.TrimSpace(body.Email)
	if email == "" {
		http.Error(w, "email ต้องไม่ว่าง", http.StatusBadRequest)
		return
	}
	passwordHex, ok := normalizeSHA256Hex(body.Password)
	if !ok {
		http.Error(w, "password ต้องเป็น SHA-256 hex 64 ตัวอักษร", http.StatusBadRequest)
		return
	}

	u, err := h.service.Login(r.Context(), email, passwordHex)
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

	email := strings.TrimSpace(body.Email)
	if email == "" {
		http.Error(w, "email ต้องไม่ว่าง", http.StatusBadRequest)
		return
	}
	oldPassword, ok := normalizeSHA256Hex(body.OldPassword)
	if !ok {
		http.Error(w, "old_password ต้องเป็น SHA-256 hex 64 ตัวอักษร", http.StatusBadRequest)
		return
	}
	newPassword, ok := normalizeSHA256Hex(body.NewPassword)
	if !ok {
		http.Error(w, "new_password ต้องเป็น SHA-256 hex 64 ตัวอักษร", http.StatusBadRequest)
		return
	}

	if err := h.service.ChangePassword(r.Context(), email, oldPassword, newPassword); err != nil {
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

func normalizeSHA256Hex(input string) (string, bool) {
	value := strings.TrimSpace(input)
	value = strings.TrimPrefix(value, "0x")
	value = strings.TrimPrefix(value, "0X")
	value = strings.ToLower(value)
	if len(value) != 64 {
		return "", false
	}
	if _, err := hex.DecodeString(value); err != nil {
		return "", false
	}
	return value, true
}
