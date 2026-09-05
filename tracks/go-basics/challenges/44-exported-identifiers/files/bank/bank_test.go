package bank

import "testing"

// This file is inside the package, so it can reach the unexported helper. A
// test that needs the exported surface only would sit in `package bank_test`.

func TestNormalizeOwner(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Ada Lovelace", "Ada Lovelace"},
		{"  Ada Lovelace  ", "Ada Lovelace"},
		{"Ada    Lovelace", "Ada Lovelace"},
		{"\tAda\n Lovelace ", "Ada Lovelace"},
		{"   ", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := normalizeOwner(c.in); got != c.want {
			t.Errorf("normalizeOwner(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNewAccountNormalizesTheOwner(t *testing.T) {
	a, err := NewAccount("  ada   lovelace ", 0)
	if err != nil {
		t.Fatalf("NewAccount returned %v, want nil", err)
	}
	if a.owner != "ada lovelace" {
		t.Errorf("stored owner = %q, want %q", a.owner, "ada lovelace")
	}
	if a.balance != 0 {
		t.Errorf("stored balance = %d, want 0", a.balance)
	}
}
