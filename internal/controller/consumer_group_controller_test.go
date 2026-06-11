package controller

import (
	"testing"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
)

func TestNormalizeOffsetResetRequestAcceptsType(t *testing.T) {
	req := &dto.OffsetResetRequest{Type: "beginning"}

	if err := normalizeOffsetResetRequest(req); err != nil {
		t.Fatalf("normalizeOffsetResetRequest() error = %v", err)
	}
	if req.Type != "beginning" {
		t.Fatalf("Type = %q, want beginning", req.Type)
	}
}

func TestNormalizeOffsetResetRequestAcceptsSeekAlias(t *testing.T) {
	req := &dto.OffsetResetRequest{Seek: "custom"}

	if err := normalizeOffsetResetRequest(req); err != nil {
		t.Fatalf("normalizeOffsetResetRequest() error = %v", err)
	}
	if req.Type != "offset" {
		t.Fatalf("Type = %q, want offset", req.Type)
	}
}

func TestNormalizeOffsetResetRequestRejectsInvalidType(t *testing.T) {
	req := &dto.OffsetResetRequest{Type: "middle"}

	if err := normalizeOffsetResetRequest(req); err == nil {
		t.Fatal("normalizeOffsetResetRequest() error = nil, want error")
	}
}
