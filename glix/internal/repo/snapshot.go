package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FileSnapshot captures the bytes of a single file (or its absence) so a
// mutating command can roll back on failure.
type FileSnapshot struct {
	Path   string
	Data   []byte
	Exists bool
}

// TakeFileSnapshot reads path into memory, recording its existence.
func TakeFileSnapshot(path string) (FileSnapshot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return FileSnapshot{Path: path, Exists: false}, nil
		}
		return FileSnapshot{}, err
	}
	return FileSnapshot{Path: path, Data: b, Exists: true}, nil
}

// Restore writes the snapshot back to disk, or removes the file if it did
// not exist at snapshot time.
func (s FileSnapshot) Restore() error {
	if !s.Exists {
		if err := os.Remove(s.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.Path, s.Data, 0o644)
}

// Snapshot captures all files glix may mutate in one operation: a host's
// manifest, the top-level flake.nix, and flake.lock.
type Snapshot struct {
	Manifest FileSnapshot
	Flake    FileSnapshot
	Lock     FileSnapshot
}

// TakeSnapshot captures the current state of mutated files for a host.
func (r *Repo) TakeSnapshot(host string) (Snapshot, error) {
	m, err := TakeFileSnapshot(r.ManifestPath(host))
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot manifest: %w", err)
	}
	f, err := TakeFileSnapshot(r.FlakePath())
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot flake.nix: %w", err)
	}
	l, err := TakeFileSnapshot(filepath.Join(r.Root, "flake.lock"))
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot flake.lock: %w", err)
	}
	return Snapshot{Manifest: m, Flake: f, Lock: l}, nil
}

// Restore writes every captured file back. Best-effort: errors are joined.
func (s Snapshot) Restore() error {
	var errs []error
	if err := s.Manifest.Restore(); err != nil {
		errs = append(errs, fmt.Errorf("manifest: %w", err))
	}
	if err := s.Flake.Restore(); err != nil {
		errs = append(errs, fmt.Errorf("flake.nix: %w", err))
	}
	if err := s.Lock.Restore(); err != nil {
		errs = append(errs, fmt.Errorf("flake.lock: %w", err))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
