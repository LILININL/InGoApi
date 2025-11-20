package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"fristGoproject/internal/user"
	"fristGoproject/pkg/password"
)

var (
	// ErrEmailInUse จะถูกส่งกลับเมื่อพยายามสมัครซ้ำอีเมลเดิม
	ErrEmailInUse = errors.New("email นี้มีผู้ใช้งานแล้ว")
	// ErrInvalidCredentials ใช้กับ logic login หรือ change password
	ErrInvalidCredentials = errors.New("อีเมลหรือรหัสผ่านไม่ถูกต้อง")
)

// Service คือชั้นกลางที่เก็บ business logic ของ auth ทั้งหมด
type Service struct {
	users user.Repository
}

// NewService คืน service พร้อมใช้งาน
func NewService(repo user.Repository) *Service {
	return &Service{users: repo}
}

// Register สมัครสมาชิกใหม่และคืนข้อมูล user (ไม่รวม password hash)
// rawPassword ควรเป็นสตริง SHA-256 hex ที่ client แปลงมาก่อน (หรือรูปแบบที่เตรียมไว้)
func (s *Service) Register(ctx context.Context, email, rawPassword, name string) (user.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	rawPassword = strings.TrimSpace(rawPassword)
	name = strings.TrimSpace(name)
	if email == "" || rawPassword == "" {
		return user.User{}, errors.New("email และ password ต้องไม่ว่าง")
	}

	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return user.User{}, ErrEmailInUse
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, fmt.Errorf("ตรวจสอบอีเมลซ้ำ: %w", err)
	}

	hash, err := password.HashPassword(rawPassword)
	if err != nil {
		return user.User{}, fmt.Errorf("hash password: %w", err)
	}

	newUser := user.User{
		Email:        email,
		PasswordHash: hash,
		Name:         name,
	}



	if err := s.users.Create(ctx, newUser); err != nil {
		return user.User{}, fmt.Errorf("สร้างผู้ใช้: %w", err)
	}

	// ดึงข้อมูลกลับจากฐาน เพื่อให้ได้ id & created_at
	created, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return user.User{}, fmt.Errorf("ดึงข้อมูลผู้ใช้: %w", err)
	}
	created.PasswordHash = "" // ไม่ส่ง hash กลับไปยัง handler
	return created, nil
}

// Login ตรวจสอบ email/password ที่ client ส่ง (หลังเข้ารหัส SHA-256) แล้วคืนข้อมูลผู้ใช้หากสำเร็จ
func (s *Service) Login(ctx context.Context, email, rawPassword string) (user.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	rawPassword = strings.TrimSpace(rawPassword)
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, ErrInvalidCredentials
		}
		return user.User{}, fmt.Errorf("ค้นหาผู้ใช้: %w", err)
	}

	if err := password.CheckPassword(u.PasswordHash, rawPassword); err != nil {
		return user.User{}, ErrInvalidCredentials
	}

	u.PasswordHash = ""
	return u, nil
}

// ChangePassword ตรวจสอบรหัสเดิม (รูปแบบเดียวกับที่ client ส่งให้ เช่น SHA-256) ก่อนบันทึกรหัสใหม่
func (s *Service) ChangePassword(ctx context.Context, email, oldPassword, newPassword string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	oldPassword = strings.TrimSpace(oldPassword)
	newPassword = strings.TrimSpace(newPassword)
	if newPassword == "" {
		return errors.New("รหัสผ่านใหม่ต้องไม่ว่าง")
	}

	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("ค้นหาผู้ใช้: %w", err)
	}

	if err := password.CheckPassword(u.PasswordHash, oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := password.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.users.UpdatePassword(ctx, u.ID, hash); err != nil {
		return fmt.Errorf("อัปเดตรหัสผ่าน: %w", err)
	}
	return nil
}
