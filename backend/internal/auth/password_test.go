package auth

import (
	"regexp"
	"testing"
)

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	got := NormalizeEmail("  Test@Example.COM ")
	if got != "test@example.com" {
		t.Fatalf("NormalizeEmail() = %q, want %q", got, "test@example.com")
	}
}

func TestValidateEmail(t *testing.T) {
	t.Parallel()

	if err := ValidateEmail("test@mails.tsinghua.edu.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error: %v", err)
	}
	if err := ValidateEmail("test@tsinghua.org.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for tsinghua.org.cn email: %v", err)
	}
	if err := ValidateEmail("test@mail.tsinghua.org.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for mail.tsinghua.org.cn email: %v", err)
	}
	if err := ValidateEmail("test@tsinghua.edu.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for tsinghua.edu.cn email: %v", err)
	}
	if err := ValidateEmail("test@student.pumc.edu.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for student.pumc.edu.cn email: %v", err)
	}
	if err := ValidateEmail("test@pku.edu.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for PKU email: %v", err)
	}
	if err := ValidateEmail("test@stu.pku.edu.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for stu.pku.edu.cn email: %v", err)
	}
	if err := ValidateEmail("test@bimba.pku.edu.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for bimba.pku.edu.cn email: %v", err)
	}
	if err := ValidateEmail("test@bfsu.edu.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for bfsu.edu.cn email: %v", err)
	}
	if err := ValidateEmail("test@stu.zuel.edu.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for ZUEL email: %v", err)
	}
	if err := ValidateEmail("test@u.nus.edu"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for NUS email: %v", err)
	}
	if err := ValidateEmail("test@e.ntu.edu.sg"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for NTU email: %v", err)
	}
	if err := ValidateEmail("test@m.fudan.edu.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for Shanghai Jiao Tong email: %v", err)
	}
	if err := ValidateEmail("test@sjtu.edu.cn"); err != nil {
		t.Fatalf("ValidateEmail() unexpected error for Fudan email: %v", err)
	}
	if err := ValidateEmail("bad-email"); err == nil {
		t.Fatal("ValidateEmail() expected error for invalid email")
	}
	if err := ValidateEmail("test@example.com"); err == nil {
		t.Fatal("ValidateEmail() expected error for unsupported domain")
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	if err := ValidatePassword("12345678"); err != nil {
		t.Fatalf("ValidatePassword() unexpected error: %v", err)
	}
	if err := ValidatePassword("123"); err == nil {
		t.Fatal("ValidatePassword() expected error for short password")
	}
}

func TestValidateVerificationCode(t *testing.T) {
	t.Parallel()

	if err := ValidateVerificationCode("123456"); err != nil {
		t.Fatalf("ValidateVerificationCode() unexpected error: %v", err)
	}
	if err := ValidateVerificationCode("12345"); err == nil {
		t.Fatal("ValidateVerificationCode() expected error for invalid code")
	}
}

func TestGenerateVerificationCode(t *testing.T) {
	t.Parallel()

	code, err := GenerateVerificationCode()
	if err != nil {
		t.Fatalf("GenerateVerificationCode() unexpected error: %v", err)
	}
	if !regexp.MustCompile(`^\d{6}$`).MatchString(code) {
		t.Fatalf("GenerateVerificationCode() = %q, want 6 digits", code)
	}
}

func TestHashVerificationCode(t *testing.T) {
	t.Parallel()

	first := HashVerificationCode("register", "test@example.com", "123456", "secret")
	second := HashVerificationCode("register", "test@example.com", "123456", "secret")
	third := HashVerificationCode("register", "test@example.com", "654321", "secret")

	if first != second {
		t.Fatal("HashVerificationCode() should be deterministic")
	}
	if first == third {
		t.Fatal("HashVerificationCode() should change when code changes")
	}
}
