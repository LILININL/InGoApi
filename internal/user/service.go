package user

import (
	"context"
	"fmt"
)

// Service เก็บ logic เพิ่มเติมเกี่ยวกับข้อมูลผู้ใช้ (นอกเหนือจาก auth)
type Service struct {
	repo Repository
}

// NewService คืน service ที่ใช้ repository เดิม
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List ดึงผู้ใช้ทั้งหมดจากฐานข้อมูล
func (s *Service) List(ctx context.Context) ([]User, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("ดึงรายชื่อผู้ใช้: %w", err)
	}
	// ไม่ต้องเปิดเผย password hash
	for i := range users {
		users[i].PasswordHash = ""
	}
	return users, nil
}
