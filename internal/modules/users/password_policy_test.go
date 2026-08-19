package users

import (
	"errors"
	"strings"
	"testing"
)

func TestPasswordPolicyHashAndVerify(t *testing.T) {
	policy := NewPasswordPolicy()
	password := "correct horse battery staple"

	hash, err := policy.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "" {
		t.Fatal("Hash() returned an empty hash")
	}
	if hash == password {
		t.Fatal("Hash() returned the plaintext password")
	}

	if err := policy.Verify(hash, password); err != nil {
		t.Fatalf("Verify() with the original password error = %v", err)
	}
	if err := policy.Verify(hash, "wrong password"); err == nil {
		t.Fatal("Verify() accepted an incorrect password")
	}
}

func TestPasswordPolicyEnforcesBcryptByteLimit(t *testing.T) {
	policy := NewPasswordPolicy()

	if _, err := policy.Hash(strings.Repeat("a", PasswordMaxBytes)); err != nil {
		t.Fatalf("Hash() at bcrypt byte limit error = %v", err)
	}
	if _, err := policy.Hash(strings.Repeat("a", PasswordMaxBytes+1)); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("Hash() above bcrypt byte limit error = %v, want ErrPasswordTooLong", err)
	}

	multibytePassword := strings.Repeat("я", PasswordMaxBytes/2+1)
	if _, err := policy.Hash(multibytePassword); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("Hash() overlong UTF-8 password error = %v, want ErrPasswordTooLong", err)
	}
}

func TestPasswordPolicyVerifyDistinguishesMismatchFromInvalidHash(t *testing.T) {
	policy := NewPasswordPolicy()
	password := "correct horse battery staple"
	hash, err := policy.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if err := policy.Verify(hash, "different password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Verify() mismatch error = %v, want ErrInvalidCredentials", err)
	}
	if err := policy.Verify("invalid stored hash", password); !errors.Is(err, ErrInternal) {
		t.Fatalf("Verify() invalid hash error = %v, want ErrInternal", err)
	}
}
