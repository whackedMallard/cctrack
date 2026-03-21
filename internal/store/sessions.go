package store

import (
	"database/sql"
	"fmt"
)

type Session struct {
	ID             string  `json:"id"`
	Project        string  `json:"project"`
	Slug           string  `json:"slug"`
	Model          string  `json:"model"`
	Branch         string  `json:"branch,omitempty"`
	StartedAt      string  `json:"started_at"`
	LastActivity   string  `json:"last_activity"`
	TotalInput     int64   `json:"total_input"`
	TotalOutput    int64   `json:"total_output"`
	TotalCacheRead int64   `json:"total_cache_read"`
	TotalCacheWrite int64  `json:"total_cache_write"`
	TotalCost      float64 `json:"total_cost"`
}

func (s *Session) TotalTokens() int64 {
	return s.TotalInput + s.TotalOutput + s.TotalCacheRead + s.TotalCacheWrite
}

type SessionDelta struct {
	ID             string
	Project        string
	Slug           string
	Model          string
	GitBranch      string
	Timestamp      string
	DeltaInput     int64
	DeltaOutput    int64
	DeltaCacheRead int64
	DeltaCacheWrite int64
	DeltaCost      float64
}

// UpsertSession adds token deltas to an existing session or creates a new one.
// Token counts are ADDITIVE — new values add to existing totals.
func (s *Store) UpsertSession(d SessionDelta) error {
	_, err := s.db.Exec(`
		INSERT INTO sessions (id, project, slug, model, started_at, last_activity,
			total_input, total_output, total_cache_read, total_cache_write, total_cost)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			slug          = CASE WHEN excluded.slug != '' THEN excluded.slug ELSE sessions.slug END,
			model         = CASE WHEN excluded.model != '' THEN excluded.model ELSE sessions.model END,
			last_activity = CASE WHEN excluded.last_activity > sessions.last_activity THEN excluded.last_activity ELSE sessions.last_activity END,
			total_input   = sessions.total_input   + excluded.total_input,
			total_output  = sessions.total_output  + excluded.total_output,
			total_cache_read  = sessions.total_cache_read  + excluded.total_cache_read,
			total_cache_write = sessions.total_cache_write + excluded.total_cache_write,
			total_cost    = sessions.total_cost    + excluded.total_cost
	`, d.ID, d.Project, d.Slug, d.Model, d.Timestamp, d.Timestamp,
		d.DeltaInput, d.DeltaOutput, d.DeltaCacheRead, d.DeltaCacheWrite, d.DeltaCost)
	return err
}

func (s *Store) GetSession(id string) (*Session, error) {
	row := s.db.QueryRow(`SELECT s.id, s.project, s.slug, s.model, s.started_at, s.last_activity,
		s.total_input, s.total_output, s.total_cache_read, s.total_cache_write, s.total_cost,
		COALESCE((SELECT sb.branch FROM session_branches sb WHERE sb.session_id = s.id ORDER BY sb.last_seen DESC LIMIT 1), '')
		FROM sessions s WHERE s.id = ?`, id)
	sess := &Session{}
	err := row.Scan(&sess.ID, &sess.Project, &sess.Slug, &sess.Model,
		&sess.StartedAt, &sess.LastActivity,
		&sess.TotalInput, &sess.TotalOutput, &sess.TotalCacheRead, &sess.TotalCacheWrite,
		&sess.TotalCost, &sess.Branch)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// --- Request-level tracking ---

type RequestRecord struct {
	RequestID        string  `json:"request_id"`
	SessionID        string  `json:"session_id"`
	Timestamp        string  `json:"timestamp"`
	Model            string  `json:"model"`
	InputTokens      int64   `json:"input_tokens"`
	OutputTokens     int64   `json:"output_tokens"`
	CacheReadTokens  int64   `json:"cache_read_tokens"`
	CacheWriteTokens int64   `json:"cache_write_tokens"`
	Cost             float64 `json:"cost"`
	GitBranch        string  `json:"git_branch"`
}

func (s *Store) UpsertRequest(r RequestRecord) error {
	_, err := s.db.Exec(`
		INSERT INTO requests (request_id, session_id, timestamp, model,
			input_tokens, output_tokens, cache_read_tokens, cache_write_tokens, cost, git_branch)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(request_id) DO UPDATE SET
			timestamp = excluded.timestamp,
			model = excluded.model,
			input_tokens = excluded.input_tokens,
			output_tokens = excluded.output_tokens,
			cache_read_tokens = excluded.cache_read_tokens,
			cache_write_tokens = excluded.cache_write_tokens,
			cost = excluded.cost,
			git_branch = excluded.git_branch
	`, r.RequestID, r.SessionID, r.Timestamp, r.Model,
		r.InputTokens, r.OutputTokens, r.CacheReadTokens, r.CacheWriteTokens, r.Cost, r.GitBranch)
	return err
}

func (s *Store) GetSessionRequests(sessionID string) ([]RequestRecord, error) {
	rows, err := s.db.Query(`
		SELECT request_id, session_id, timestamp, model,
			input_tokens, output_tokens, cache_read_tokens, cache_write_tokens, cost, git_branch
		FROM requests WHERE session_id = ?
		ORDER BY timestamp ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recs []RequestRecord
	for rows.Next() {
		var r RequestRecord
		if err := rows.Scan(&r.RequestID, &r.SessionID, &r.Timestamp, &r.Model,
			&r.InputTokens, &r.OutputTokens, &r.CacheReadTokens, &r.CacheWriteTokens,
			&r.Cost, &r.GitBranch); err != nil {
			return nil, err
		}
		recs = append(recs, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return recs, nil
}

// RebuildSessionTotals recalculates sessions and session_branches totals
// from the correctly-deduped requests table, fixing subagent duplication.
func (s *Store) RebuildSessionTotals(sessionIDs []string) error {
	for _, sid := range sessionIDs {
		// Aggregate all request-level data in a single scan
		var totalInput, totalOutput, totalCacheRead, totalCacheWrite int64
		var totalCost float64
		var minTs, maxTs sql.NullString
		err := s.db.QueryRow(`
			SELECT COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0),
			       COALESCE(SUM(cache_read_tokens), 0), COALESCE(SUM(cache_write_tokens), 0),
			       COALESCE(SUM(cost), 0), MIN(timestamp), MAX(timestamp)
			FROM requests WHERE session_id = ?`, sid).Scan(
			&totalInput, &totalOutput, &totalCacheRead, &totalCacheWrite,
			&totalCost, &minTs, &maxTs)
		if err != nil {
			return fmt.Errorf("rebuild session %s: %w", sid, err)
		}

		// Update session with aggregated values
		startedAt := ""
		if minTs.Valid {
			startedAt = minTs.String
		}
		lastActivity := ""
		if maxTs.Valid {
			lastActivity = maxTs.String
		}

		_, err = s.db.Exec(`
			UPDATE sessions SET
				total_input = ?, total_output = ?,
				total_cache_read = ?, total_cache_write = ?,
				total_cost = ?,
				started_at = CASE WHEN ? != '' THEN ? ELSE started_at END,
				last_activity = CASE WHEN ? != '' THEN ? ELSE last_activity END
			WHERE id = ?`,
			totalInput, totalOutput, totalCacheRead, totalCacheWrite, totalCost,
			startedAt, startedAt, lastActivity, lastActivity, sid)
		if err != nil {
			return fmt.Errorf("rebuild session %s: %w", sid, err)
		}

		// Delete existing session_branches for this session, then rebuild from requests
		_, err = s.db.Exec("DELETE FROM session_branches WHERE session_id = ?", sid)
		if err != nil {
			return fmt.Errorf("clear session_branches %s: %w", sid, err)
		}

		// Rebuild session_branches from requests grouped by git_branch
		_, err = s.db.Exec(`
			INSERT INTO session_branches (session_id, branch, first_seen, last_seen,
				total_input, total_output, total_cache_read, total_cache_write, total_cost)
			SELECT session_id,
				   CASE WHEN git_branch = '' THEN 'No repo' ELSE git_branch END,
				   MIN(timestamp), MAX(timestamp),
				   SUM(input_tokens), SUM(output_tokens),
				   SUM(cache_read_tokens), SUM(cache_write_tokens),
				   SUM(cost)
			FROM requests
			WHERE session_id = ?
			GROUP BY session_id, CASE WHEN git_branch = '' THEN 'No repo' ELSE git_branch END
		`, sid)
		if err != nil {
			return fmt.Errorf("rebuild session_branches %s: %w", sid, err)
		}
	}
	return nil
}

var allowedSortColumns = map[string]string{
	"cost":       "total_cost",
	"date":       "last_activity",
	"started":    "started_at",
	"tokens":     "(total_input + total_output + total_cache_read + total_cache_write)",
	"model":      "model",
	"project":    "project",
}

func (s *Store) ListSessions(limit, offset int, sortBy, sortDir string) ([]Session, int, error) {
	col, ok := allowedSortColumns[sortBy]
	if !ok {
		col = "total_cost"
	}
	dir := "DESC"
	if sortDir == "asc" {
		dir = "ASC"
	}

	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT s.id, s.project, s.slug, s.model, s.started_at, s.last_activity,
		s.total_input, s.total_output, s.total_cache_read, s.total_cache_write, s.total_cost,
		COALESCE((SELECT sb.branch FROM session_branches sb WHERE sb.session_id = s.id ORDER BY sb.last_seen DESC LIMIT 1), '')
		FROM sessions s ORDER BY %s %s LIMIT ? OFFSET ?`, col, dir)

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.Project, &sess.Slug, &sess.Model,
			&sess.StartedAt, &sess.LastActivity,
			&sess.TotalInput, &sess.TotalOutput, &sess.TotalCacheRead, &sess.TotalCacheWrite,
			&sess.TotalCost, &sess.Branch); err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, sess)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

// --- Session-branch tracking ---

type SessionBranch struct {
	SessionID       string  `json:"session_id"`
	Branch          string  `json:"branch"`
	FirstSeen       string  `json:"first_seen"`
	LastSeen        string  `json:"last_seen"`
	TotalInput      int64   `json:"total_input"`
	TotalOutput     int64   `json:"total_output"`
	TotalCacheRead  int64   `json:"total_cache_read"`
	TotalCacheWrite int64   `json:"total_cache_write"`
	TotalCost       float64 `json:"total_cost"`
}

// SessionBranchRow is the joined query result for the Sessions page.
// Each row is a (session, branch) pair with session metadata.
type SessionBranchRow struct {
	ID              string  `json:"id"`
	Project         string  `json:"project"`
	Slug            string  `json:"slug"`
	Model           string  `json:"model"`
	Branch          string  `json:"branch"`
	FirstSeen       string  `json:"first_seen"`
	LastSeen        string  `json:"last_seen"`
	TotalInput      int64   `json:"total_input"`
	TotalOutput     int64   `json:"total_output"`
	TotalCacheRead  int64   `json:"total_cache_read"`
	TotalCacheWrite int64   `json:"total_cache_write"`
	TotalCost       float64 `json:"total_cost"`
}

// UpsertSessionBranch adds token deltas to a (session, branch) row or creates one.
// firstSeen is passed separately because SessionDelta.Timestamp represents last_seen.
func (s *Store) UpsertSessionBranch(d SessionDelta, firstSeen string) error {
	_, err := s.db.Exec(`
		INSERT INTO session_branches (session_id, branch, first_seen, last_seen,
			total_input, total_output, total_cache_read, total_cache_write, total_cost)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(session_id, branch) DO UPDATE SET
			last_seen     = CASE WHEN excluded.last_seen > session_branches.last_seen THEN excluded.last_seen ELSE session_branches.last_seen END,
			first_seen    = CASE WHEN excluded.first_seen < session_branches.first_seen THEN excluded.first_seen ELSE session_branches.first_seen END,
			total_input   = session_branches.total_input   + excluded.total_input,
			total_output  = session_branches.total_output  + excluded.total_output,
			total_cache_read  = session_branches.total_cache_read  + excluded.total_cache_read,
			total_cache_write = session_branches.total_cache_write + excluded.total_cache_write,
			total_cost    = session_branches.total_cost    + excluded.total_cost
	`, d.ID, d.GitBranch, firstSeen, d.Timestamp,
		d.DeltaInput, d.DeltaOutput, d.DeltaCacheRead, d.DeltaCacheWrite, d.DeltaCost)
	return err
}

// allowedBranchSortColumns maps API sort keys to SQL expressions for session_branches queries.
var allowedBranchSortColumns = map[string]string{
	"cost":    "sb.total_cost",
	"date":    "sb.last_seen",
	"started": "sb.first_seen",
	"tokens":  "(sb.total_input + sb.total_output + sb.total_cache_read + sb.total_cache_write)",
	"model":   "s.model",
	"project": "s.project",
	"branch":  "sb.branch",
}

// ListSessionBranches returns paginated (session, branch) rows joined with session metadata.
func (s *Store) ListSessionBranches(limit, offset int, sortBy, sortDir string) ([]SessionBranchRow, int, error) {
	col, ok := allowedBranchSortColumns[sortBy]
	if !ok {
		col = "sb.total_cost"
	}
	dir := "DESC"
	if sortDir == "asc" {
		dir = "ASC"
	}

	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM session_branches").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT s.id, s.project, s.slug, s.model, sb.branch,
		       sb.first_seen, sb.last_seen,
		       sb.total_input, sb.total_output,
		       sb.total_cache_read, sb.total_cache_write,
		       sb.total_cost
		FROM session_branches sb
		JOIN sessions s ON s.id = sb.session_id
		ORDER BY %s %s
		LIMIT ? OFFSET ?`, col, dir)

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []SessionBranchRow
	for rows.Next() {
		var r SessionBranchRow
		if err := rows.Scan(&r.ID, &r.Project, &r.Slug, &r.Model, &r.Branch,
			&r.FirstSeen, &r.LastSeen,
			&r.TotalInput, &r.TotalOutput, &r.TotalCacheRead, &r.TotalCacheWrite,
			&r.TotalCost); err != nil {
			return nil, 0, err
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}
