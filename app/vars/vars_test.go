package vars

import "testing"

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"orange", true},
		{"_pineapple", true},
		{"Grape_2", true},
		{"5raspberry", false},
		{"pear-grape", false},
		{"", false},
		{"a", true},
		{"_", true},
	}
	for _, tt := range tests {
		if got := IsValidIdentifier(tt.name); got != tt.want {
			t.Errorf("IsValidIdentifier(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestStore_SetGet(t *testing.T) {
	s := New()
	if _, ok := s.Get("x"); ok {
		t.Fatal("expected missing")
	}
	s.Set("x", "1")
	if got, ok := s.Get("x"); !ok || got != "1" {
		t.Fatalf("Get = %q %v, want 1 true", got, ok)
	}
	s.Set("x", "2")
	if got, ok := s.Get("x"); !ok || got != "2" {
		t.Fatalf("Get after overwrite = %q %v, want 2 true", got, ok)
	}
}
