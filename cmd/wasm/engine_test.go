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

func TestEvaluateRoll_EventCounts(t *testing.T) {
	t.Run("drop only", func(t *testing.T) {
		resp := EvaluateRoll("4d6kh3")
		if !resp.OK {
			t.Fatalf("expected OK=true, got error: %s", resp.Error)
		}
		if !resp.HasDrops || resp.Drops != 1 {
			t.Errorf("expected HasDrops with 1 drop, got has=%v drops=%d", resp.HasDrops, resp.Drops)
		}
		if resp.HasRerolls || resp.HasExplosions {
			t.Errorf("unexpected reroll/explosion flags: %+v", resp)
		}
	})

	t.Run("explosion present but may not fire", func(t *testing.T) {
		resp := EvaluateRoll("4d6!")
		if !resp.OK {
			t.Fatalf("expected OK=true, got error: %s", resp.Error)
		}
		if !resp.HasExplosions {
			t.Errorf("expected HasExplosions for a roll containing explosions")
		}
	})

	t.Run("reroll flag", func(t *testing.T) {
		resp := EvaluateRoll("4d6ro1")
		if !resp.OK {
			t.Fatalf("expected OK=true, got error: %s", resp.Error)
		}
		if !resp.HasRerolls {
			t.Errorf("expected HasRerolls for a roll containing rerolls")
		}
	})

	t.Run("plain roll has no event flags", func(t *testing.T) {
		resp := EvaluateRoll("4d6")
		if !resp.OK {
			t.Fatalf("expected OK=true, got error: %s", resp.Error)
		}
		if resp.HasRerolls || resp.HasExplosions || resp.HasDrops {
			t.Errorf("unexpected flags for a plain roll: %+v", resp)
		}
	})
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
