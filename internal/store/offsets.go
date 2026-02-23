package store

import "time"

func (s *Store) GetFileOffset(path string) (int64, error) {
	var offset int64
	err := s.db.QueryRow("SELECT offset FROM file_offsets WHERE path = ?", path).Scan(&offset)
	if err != nil {
		return 0, nil // not found = start from beginning
	}
	return offset, nil
}

func (s *Store) SetFileOffset(path string, offset int64) error {
	_, err := s.db.Exec(`
		INSERT INTO file_offsets (path, offset, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET offset = excluded.offset, updated_at = excluded.updated_at
	`, path, offset, time.Now().UTC().Format(time.RFC3339))
	return err
}
