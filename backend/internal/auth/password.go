package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	mailaddress "net/mail"
	"strings"

	"soulink-backend/internal/httpapi"
	"soulink-backend/internal/school"
)

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ValidateEmail(email string) error {
	normalized := NormalizeEmail(email)
	if normalized == "" {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidEmail, "邮箱格式错误")
	}
	if _, err := mailaddress.ParseAddress(normalized); err != nil {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidEmail, "邮箱格式错误")
	}
	if !school.SupportsEmail(normalized) {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidEmail, "邮箱格式错误")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidPassword, "密码格式不合法")
	}
	return nil
}

func ValidateVerificationCode(code string) error {
	if len(strings.TrimSpace(code)) != 6 {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidVerificationCode, "验证码错误")
	}
	return nil
}

func ValidateRequiredText(value string) error {
	if strings.TrimSpace(value) == "" {
		return httpapi.NewError(http.StatusBadRequest, httpapi.CodeInvalidParams, "参数错误")
	}
	return nil
}

func GenerateVerificationCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generate verification code: %w", err)
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func HashVerificationCode(purpose, email, code, secret string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(email) + "|" + purpose + "|" + code + "|" + secret))
	return hex.EncodeToString(sum[:])
}
