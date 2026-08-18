package users

import "testing"

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
