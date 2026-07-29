package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FilesystemImageStore struct {
	root string
}

func NewFilesystemImageStore(root string) (*FilesystemImageStore, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("image storage directory is required")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolving image storage directory: %w", err)
	}
	if err := os.MkdirAll(absoluteRoot, 0o750); err != nil {
		return nil, fmt.Errorf("creating image storage directory: %w", err)
	}
	return &FilesystemImageStore{root: absoluteRoot}, nil
}

func (s *FilesystemImageStore) Save(ctx context.Context, key string, source io.Reader) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(s.root, ".upload-*")
	if err != nil {
		return fmt.Errorf("creating temporary image: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath) //nolint:errcheck -- best-effort cleanup

	if err := temporary.Chmod(0o640); err != nil {
		temporary.Close()
		return fmt.Errorf("setting image permissions: %w", err)
	}
	if _, err := io.Copy(temporary, &contextReader{ctx: ctx, reader: source}); err != nil {
		temporary.Close()
		return fmt.Errorf("writing image: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("syncing image: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("closing image: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("storing image: %w", err)
	}
	return nil
}

func (s *FilesystemImageStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	path, err := s.path(key)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("opening image: %w", err)
	}
	return file, nil
}

func (s *FilesystemImageStore) Delete(_ context.Context, key string) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("deleting image: %w", err)
	}
	return nil
}

func (s *FilesystemImageStore) path(key string) (string, error) {
	if key == "" || filepath.Base(key) != key || strings.ContainsAny(key, `/\`) {
		return "", errors.New("invalid image storage key")
	}
	return filepath.Join(s.root, key), nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(buffer []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(buffer)
	}
}
