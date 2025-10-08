package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

// HashPassword ใช้ Argon2id แปลงรหัสผ่านที่ client ส่งมา (หลัง SHA-256) ให้เป็น hash สำหรับเก็บในฐาน
func HashPassword(raw string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("สร้าง salt: %w", err)
	}

	key := argon2.IDKey([]byte(raw), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedKey := base64.RawStdEncoding.EncodeToString(key)

	hash := fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemory, argonTime, argonThreads, encodedSalt, encodedKey)
	return hash, nil
}

// CheckPassword เปรียบเทียบรหัสผ่าน (หลัง SHA-256 จาก client) กับ hash ที่เก็บในฐานข้อมูล
func CheckPassword(hash, raw string) error {
	parts := strings.Split(hash, "$")
	if len(parts) != 5 {
		return errors.New("รูปแบบ hash ไม่ถูกต้อง")
	}

	if parts[0] != "argon2id" || parts[1] != "v=19" {
		return errors.New("hash ไม่รองรับอัลกอริทึมนี้")
	}

	paramsPart := parts[2]
	paramValues := strings.Split(paramsPart, ",")
	if len(paramValues) != 3 {
		return errors.New("รูปแบบพารามิเตอร์ Argon2 ไม่ถูกต้อง")
	}

	var memory uint64
	var time uint64
	var threads uint64
	for _, param := range paramValues {
		keyValue := strings.SplitN(param, "=", 2)
		if len(keyValue) != 2 {
			return errors.New("รูปแบบพารามิเตอร์ Argon2 ไม่ถูกต้อง")
		}
		value, err := strconv.ParseUint(keyValue[1], 10, 32)
		if err != nil {
			return fmt.Errorf("อ่านค่า %s: %w", keyValue[0], err)
		}
		switch keyValue[0] {
		case "m":
			memory = value
		case "t":
			time = value
		case "p":
			threads = value
		default:
			return fmt.Errorf("พารามิเตอร์ไม่รองรับ: %s", keyValue[0])
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return fmt.Errorf("ถอดรหัส salt: %w", err)
	}

	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("ถอดรหัส key: %w", err)
	}

	derivedKey := argon2.IDKey([]byte(raw), salt, uint32(time), uint32(memory), uint8(threads), uint32(len(expectedKey)))

	if subtle.ConstantTimeCompare(derivedKey, expectedKey) != 1 {
		return errors.New("hash ไม่ตรง")
	}
	return nil
}
