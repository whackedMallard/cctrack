package parser

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksred/cctrack/internal/calculator"
	"github.com/ksred/cctrack/internal/store"
)

func TestParseFile_SkipsNegativeTokens(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	p := New(s)

	// Create a JSONL file with one valid event and one with negative tokens
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-home-user-Github-test")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	jsonl := `{"type":"assistant","sessionId":"sess-valid","timestamp":"2026-03-19T10:00:00Z","message":{"model":"claude-sonnet-4-20250514","usage":{"input_tokens":100,"output_tokens":50,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}
{"type":"assistant","sessionId":"sess-negative","timestamp":"2026-03-19T10:01:00Z","message":{"model":"claude-sonnet-4-20250514","usage":{"input_tokens":-100,"output_tokens":50,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`

	testFile := filepath.Join(projectDir, "test-session.jsonl")
	if err := os.WriteFile(testFile, []byte(jsonl), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	sessions, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	// Only the valid session should be affected
	if len(sessions) != 1 {
		t.Errorf("expected 1 affected session, got %d", len(sessions))
	}

	// The negative-token session should not exist in the store
	_, err = s.GetSession("sess-negative")
	if err == nil {
		t.Error("sess-negative should not exist in store (negative tokens should be skipped)")
	}

	// The valid session should exist
	sess, err := s.GetSession("sess-valid")
	if err != nil {
		t.Fatalf("sess-valid should exist: %v", err)
	}
	if sess.TotalInput != 100 {
		t.Errorf("expected input=100, got %d", sess.TotalInput)
	}
}

func TestParseFile_ExtractsGitBranch(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	p := New(s)

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-home-user-Github-test")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Event WITH gitBranch and event WITHOUT gitBranch
	jsonl := `{"type":"assistant","sessionId":"sess-branch-test","gitBranch":"feat/test-branch","timestamp":"2026-03-19T10:00:00Z","message":{"model":"claude-sonnet-4-20250514","usage":{"input_tokens":100,"output_tokens":50,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}
{"type":"assistant","sessionId":"sess-no-branch","timestamp":"2026-03-19T10:01:00Z","message":{"model":"claude-sonnet-4-20250514","usage":{"input_tokens":200,"output_tokens":75,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`

	testFile := filepath.Join(projectDir, "branch-test.jsonl")
	if err := os.WriteFile(testFile, []byte(jsonl), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	sessions, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 affected sessions, got %d", len(sessions))
	}

	// Verify session with gitBranch has a branch row for "feat/test-branch"
	branches, _, err := s.ListSessionBranches(100, 0, "cost", "desc")
	if err != nil {
		t.Fatalf("ListSessionBranches: %v", err)
	}

	foundBranch := false
	foundNoRepo := false
	for _, b := range branches {
		if b.ID == "sess-branch-test" && b.Branch == "feat/test-branch" {
			foundBranch = true
			if b.TotalInput != 100 {
				t.Errorf("branch row input: got %d, want 100", b.TotalInput)
			}
			if b.TotalOutput != 50 {
				t.Errorf("branch row output: got %d, want 50", b.TotalOutput)
			}
		}
		if b.ID == "sess-no-branch" && b.Branch == "No repo" {
			foundNoRepo = true
			if b.TotalInput != 200 {
				t.Errorf("no-repo row input: got %d, want 200", b.TotalInput)
			}
		}
	}
	if !foundBranch {
		t.Error("expected branch row for sess-branch-test/feat/test-branch")
	}
	if !foundNoRepo {
		t.Error("expected branch row for sess-no-branch/No repo (default)")
	}
}

func TestParseFile_MixedModelCost(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	p := New(s)

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-home-user-Github-test")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// One Haiku event and one Sonnet event in the same session
	jsonl := `{"type":"assistant","sessionId":"sess-mixed","requestId":"req-haiku","timestamp":"2026-03-19T10:00:00Z","message":{"model":"claude-haiku-4-5-20251001","usage":{"input_tokens":1000,"output_tokens":500,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}
{"type":"assistant","sessionId":"sess-mixed","requestId":"req-sonnet","timestamp":"2026-03-19T10:01:00Z","message":{"model":"claude-sonnet-4-20250514","usage":{"input_tokens":1000,"output_tokens":500,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`

	testFile := filepath.Join(projectDir, "mixed-model.jsonl")
	if err := os.WriteFile(testFile, []byte(jsonl), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	sessions, err := p.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	sess, err := s.GetSession("sess-mixed")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	// Calculate expected cost: sum of per-event costs
	haikuCost := calculator.Calculate("claude-haiku-4-5-20251001", calculator.TokenUsage{
		InputTokens: 1000, OutputTokens: 500,
	})
	sonnetCost := calculator.Calculate("claude-sonnet-4-20250514", calculator.TokenUsage{
		InputTokens: 1000, OutputTokens: 500,
	})
	expectedCost := haikuCost.TotalCost + sonnetCost.TotalCost

	// The session cost should be the SUM of per-event costs, not all tokens at one model
	if math.Abs(sess.TotalCost-expectedCost) > 0.0001 {
		t.Errorf("session cost = %f, want %f (sum of per-event costs)", sess.TotalCost, expectedCost)
	}

	// Verify it's NOT priced at a single model (which would be wrong)
	wrongCost := calculator.Calculate("claude-sonnet-4-20250514", calculator.TokenUsage{
		InputTokens: 2000, OutputTokens: 1000,
	})
	if math.Abs(sess.TotalCost-wrongCost.TotalCost) < 0.0001 && haikuCost.TotalCost != sonnetCost.TotalCost {
		t.Error("session cost appears to use single-model pricing (bug not fixed)")
	}
}
