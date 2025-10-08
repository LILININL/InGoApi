# InGo Auth API

Go REST API สำหรับระบบยืนยันตัวตนและจัดการผู้ใช้ตัวเบา ๆ ที่เสริมด้วย Argon2, PostgreSQL และเอกสาร OpenAPI พร้อมใช้งานในไม่กี่ขั้นตอน

## Highlights

- Auth flow ครบ: สมัคร, ล็อกอิน, เปลี่ยนรหัสผ่าน พร้อมการจัดการ error ที่ชัดเจน
- Password hashing ด้วย Argon2id ผ่าน `pkg/password` ปลอดภัยสำหรับ production
- Repository pattern คั่นกลาง service / database ทดสอบง่าย ขยายสะดวก
- Static OpenAPI + Swagger UI เสิร์ฟผ่าน `/docs/` ให้ทีม Frontend ใช้งานทันที
- CORS middleware สำเร็จรูป รองรับการเรียกจาก client ฝั่ง browser

## Stack & Structure

- **Go 1.25**, standard library HTTP + `pgx/v5` connection pool
- **PostgreSQL 16** (หรือเวอร์ชันที่รองรับ pgx) สำหรับ persistence
- **Docker Compose** (ออปชัน) ตั้ง stack `api + db` ได้ทันที
- โครงสร้างหลัก
  ```
  cmd/server        # entrypoint ของ HTTP server
  internal/auth     # business logic การยืนยันตัวตน
  internal/user     # user service + repository
  internal/httpapi  # handler, router, middleware, DTO
  internal/db       # จัดการ connection pool
  docs              # OpenAPI + Swagger UI static files
  pkg/password      # Argon2 helper สำหรับ hash/verify
  ```

## Quick Start

### 1. เตรียมฐานข้อมูล

ตั้งค่า environment variable ให้พอยท์ไปยังฐานข้อมูล PostgreSQL:

```bash
export DATABASE_URL="postgres://in:in@localhost:5432/lindb"
```

สร้างตาราง `users` (ตัวอย่าง schema ที่ใช้ใน repo นี้):

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 2. รันด้วย Go (โหมด dev)

```bash
go run ./cmd/server
```

เซิร์ฟเวอร์จะขึ้นที่ `http://localhost:8080` (เปลี่ยนพอร์ตได้ใน `cmd/server/main.go`).

### 3. รันด้วย Docker Compose (ออปชัน)

```bash
docker compose up --build
```

ระบบจะขึ้นครบทั้ง API (`http://localhost:8080`) และ PostgreSQL (`localhost:5432`).

## API Snapshot

| Method | Path                    | Description                    |
| ------ | ----------------------- | ------------------------------ |
| POST   | `/auth/register`        | สมัครสมาชิกใหม่                |
| POST   | `/auth/login`           | ล็อกอิน                        |
| POST   | `/auth/change-password` | เปลี่ยนรหัสผ่าน (ตรวจรหัสเดิม) |
| GET    | `/users`                | ดึงรายชื่อผู้ใช้ทั้งหมด        |

- รูปแบบ payload/response รายละเอียดเต็มดูที่ `/docs/` (Swagger UI) หรือไฟล์ `docs/openapi.yaml`.
- ทุก response เป็น JSON และมี CORS header ติดมาพร้อมใช้งานกับ frontends ต่างโดเมน

## Development Notes

- Argon2 helper อยู่ที่ `pkg/password/password.go` สามารถปรับ tuning parameters ได้
- Repository pattern (`internal/user/repository.go`) แยก DB ออกชัดเจน รองรับ mocking ในการทดสอบ
- Handler ชั้น `internal/httpapi` ใช้ DTO แยกต่างหาก ทำให้ reuse payload ได้หลายจุดและ parsing ง่ายขึ้น
- ค่า default ของ `DATABASE_URL` (กรณียังไม่ตั้ง env) ตั้งไว้ที่ `postgres://in:in@localhost:5432/lindb` ที่ `internal/db/postgres.go`

## Next Ideas

1. เพิ่ม unit test สำหรับ `auth.Service` และ `password` helpers
2. ติดตั้ง lint/formatter (เช่น golangci-lint) ให้เช็กรายละเอียด statically
3. ผูก JWT หรือ session layer เพิ่มเติมสำหรับ protected endpoints

พร้อมแล้วก็ git push ได้เลย 🚀
