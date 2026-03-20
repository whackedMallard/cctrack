package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_DirectoryPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	dbDir := filepath.Join(tmpDir, "newdir")
	dbPath := filepath.Join(dbDir, "test.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	s.Close()

	info, err := os.Stat(dbDir)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("db dir permissions = %o, want 0700", perm)
	}
}

func TestUpsertSessionBranch_Insert(t *testing.T) {
	s, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	// Must create the session first (FK constraint)
	err = s.UpsertSession(SessionDelta{
		ID: "sess-1", Project: "test-proj", Model: "claude-sonnet-4-20250514",
		Timestamp: "2026-03-19T10:00:00Z",
		DeltaInput: 100, DeltaOutput: 50,
	})
	if err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}

	// Insert a session-branch row
	err = s.UpsertSessionBranch(SessionDelta{
		ID: "sess-1", GitBranch: "feat/new-feature",
		Timestamp:  "2026-03-19T10:00:00Z",
		DeltaInput: 100, DeltaOutput: 50,
		DeltaCost: 0.01,
	}, "2026-03-19T10:00:00Z")
	if err != nil {
		t.Fatalf("UpsertSessionBranch: %v", err)
	}

	// Verify via ListSessionBranches
	rows, total, err := s.ListSessionBranches(10, 0, "cost", "desc")
	if err != nil {
		t.Fatalf("ListSessionBranches: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.ID != "sess-1" {
		t.Errorf("session_id: got %q, want %q", r.ID, "sess-1")
	}
	if r.Branch != "feat/new-feature" {
		t.Errorf("branch: got %q, want %q", r.Branch, "feat/new-feature")
	}
	if r.TotalInput != 100 {
		t.Errorf("total_input: got %d, want 100", r.TotalInput)
	}
	if r.TotalOutput != 50 {
		t.Errorf("total_output: got %d, want 50", r.TotalOutput)
	}
}

func TestUpsertSessionBranch_Additive(t *testing.T) {
	s, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	err = s.UpsertSession(SessionDelta{
		ID: "sess-1", Project: "test-proj", Model: "claude-sonnet-4-20250514",
		Timestamp: "2026-03-19T10:00:00Z",
		DeltaInput: 300, DeltaOutput: 150,
	})
	if err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}

	// First upsert
	err = s.UpsertSessionBranch(SessionDelta{
		ID: "sess-1", GitBranch: "main",
		Timestamp:  "2026-03-19T10:00:00Z",
		DeltaInput: 100, DeltaOutput: 50,
		DeltaCost: 0.01,
	}, "2026-03-19T10:00:00Z")
	if err != nil {
		t.Fatalf("first UpsertSessionBranch: %v", err)
	}

	// Second upsert with same key — tokens should be additive
	err = s.UpsertSessionBranch(SessionDelta{
		ID: "sess-1", GitBranch: "main",
		Timestamp:  "2026-03-19T11:00:00Z",
		DeltaInput: 200, DeltaOutput: 100,
		DeltaCost: 0.02,
	}, "2026-03-19T10:30:00Z")
	if err != nil {
		t.Fatalf("second UpsertSessionBranch: %v", err)
	}

	rows, _, err := s.ListSessionBranches(10, 0, "cost", "desc")
	if err != nil {
		t.Fatalf("ListSessionBranches: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row (same key), got %d", len(rows))
	}
	r := rows[0]
	if r.TotalInput != 300 {
		t.Errorf("total_input: got %d, want 300 (100+200)", r.TotalInput)
	}
	if r.TotalOutput != 150 {
		t.Errorf("total_output: got %d, want 150 (50+100)", r.TotalOutput)
	}
	// last_seen should be the later timestamp
	if r.LastSeen != "2026-03-19T11:00:00Z" {
		t.Errorf("last_seen: got %q, want %q", r.LastSeen, "2026-03-19T11:00:00Z")
	}
	// first_seen should be the earlier timestamp
	if r.FirstSeen != "2026-03-19T10:00:00Z" {
		t.Errorf("first_seen: got %q, want %q", r.FirstSeen, "2026-03-19T10:00:00Z")
	}
}

func TestUpsertSessionBranch_MultiBranch(t *testing.T) {
	s, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	err = s.UpsertSession(SessionDelta{
		ID: "sess-1", Project: "test-proj", Model: "claude-sonnet-4-20250514",
		Timestamp: "2026-03-19T10:00:00Z",
		DeltaInput: 500, DeltaOutput: 200,
	})
	if err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}

	// Branch A
	err = s.UpsertSessionBranch(SessionDelta{
		ID: "sess-1", GitBranch: "feat/branch-a",
		Timestamp:  "2026-03-19T10:00:00Z",
		DeltaInput: 300, DeltaOutput: 120,
		DeltaCost: 0.03,
	}, "2026-03-19T10:00:00Z")
	if err != nil {
		t.Fatalf("UpsertSessionBranch branch-a: %v", err)
	}

	// Branch B
	err = s.UpsertSessionBranch(SessionDelta{
		ID: "sess-1", GitBranch: "feat/branch-b",
		Timestamp:  "2026-03-19T11:00:00Z",
		DeltaInput: 200, DeltaOutput: 80,
		DeltaCost: 0.02,
	}, "2026-03-19T11:00:00Z")
	if err != nil {
		t.Fatalf("UpsertSessionBranch branch-b: %v", err)
	}

	rows, total, err := s.ListSessionBranches(10, 0, "cost", "desc")
	if err != nil {
		t.Fatalf("ListSessionBranches: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// Verify both branches exist with correct per-branch totals
	branchTotals := make(map[string]int64)
	for _, r := range rows {
		if r.ID != "sess-1" {
			t.Errorf("unexpected session_id: %q", r.ID)
		}
		branchTotals[r.Branch] = r.TotalInput
	}
	if branchTotals["feat/branch-a"] != 300 {
		t.Errorf("branch-a input: got %d, want 300", branchTotals["feat/branch-a"])
	}
	if branchTotals["feat/branch-b"] != 200 {
		t.Errorf("branch-b input: got %d, want 200", branchTotals["feat/branch-b"])
	}
}
