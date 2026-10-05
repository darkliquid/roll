package main

import (
	"testing"
)

func TestEvaluateRoll_Success(t *testing.T) {
	resp := EvaluateRoll("3d6+2")
	if !resp.OK {
		t.Fatalf("expected OK=true, got error: %s", resp.Error)
	}
	if resp.Input != "3d6+2" {
		t.Errorf("expected input '3d6+2', got '%s'", resp.Input)
	}
	if resp.Rendered != "3d6+2" {
		t.Errorf("expected rendered '3d6+2', got '%s'", resp.Rendered)
	}
	if len(resp.Results) != 3 {
		t.Fatalf("expected 3 dice results, got %d", len(resp.Results))
	}
	if resp.Total < 5 || resp.Total > 20 {
		t.Errorf("expected total between 5 and 20, got %d", resp.Total)
	}
	if resp.Text == "" {
		t.Errorf("expected non-empty text representation")
	}
}

func TestEvaluateRoll_FateDie(t *testing.T) {
	resp := EvaluateRoll("4dF")
	if !resp.OK {
		t.Fatalf("expected OK=true, got error: %s", resp.Error)
	}
	if len(resp.Results) != 4 {
		t.Fatalf("expected 4 dice results, got %d", len(resp.Results))
	}
	for _, r := range resp.Results {
		if r.Symbol != "⊞" && r.Symbol != "⊟" && r.Symbol != "☐" {
			t.Errorf("unexpected fate symbol: %q", r.Symbol)
		}
	}
}

func TestEvaluateRoll_Error(t *testing.T) {
	resp := EvaluateRoll("invalid_dice")
	if resp.OK {
		t.Fatalf("expected OK=false for invalid syntax")
	}
	if resp.Error == "" {
		t.Errorf("expected descriptive error message")
	}
}
