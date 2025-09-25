package service

import (
	"testing"
)

func TestServiceAddress(t *testing.T) {
	_, err := AddressFromString("aaa")
	if err == nil {
		t.Fatalf("should return an error")
	}

	sa, err := AddressFromString("8080")
	if err != nil {
		t.Errorf("should not return an error: %s", err)
	}
	if (sa.Addr != "[::1]" && sa.Addr != "127.0.0.1" && sa.Addr != "localhost") || sa.Port != 8080 {
		t.Errorf("expected not found, got: %s", sa)
	}

	sa, err = AddressFromString("0.0.0.0:8080")
	if err != nil {
		t.Errorf("should not return an error: %s", err)
	}
	if sa.Addr != "0.0.0.0" || sa.Port != 8080 {
		t.Errorf("expected not found, got: %s", sa)
	}

	sa, err = AddressFromString("skydive.network:8080")
	if err != nil {
		t.Errorf("should not return an error: %s", err)
	}
	if sa.Addr != "skydive.network" {
		t.Errorf("Expected domain, got: %s", sa.Addr)
	}
}
