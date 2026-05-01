package config

import "testing"

func TestRandomUserIDNotEmpty(t *testing.T) {
	id := randomUserID()
	if id == "" {
		t.Fatal("expected randomUserID to return non-empty value")
	}
}
