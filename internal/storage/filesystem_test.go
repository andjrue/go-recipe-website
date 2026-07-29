package storage

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestFilesystemImageStoreLifecycle(t *testing.T) {
	store, err := NewFilesystemImageStore(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	ctx := context.Background()
	if err := store.Save(ctx, "photo.jpg", strings.NewReader("jpeg bytes")); err != nil {
		t.Fatalf("saving image: %v", err)
	}
	image, err := store.Open(ctx, "photo.jpg")
	if err != nil {
		t.Fatalf("opening image: %v", err)
	}
	data, err := io.ReadAll(image)
	image.Close()
	if err != nil || string(data) != "jpeg bytes" {
		t.Fatalf("image data = %q, error = %v", data, err)
	}
	if err := store.Delete(ctx, "photo.jpg"); err != nil {
		t.Fatalf("deleting image: %v", err)
	}
	if _, err := store.Open(ctx, "photo.jpg"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("open after delete error = %v, want not found", err)
	}
}

func TestFilesystemImageStoreRejectsUnsafeKeys(t *testing.T) {
	store, err := NewFilesystemImageStore(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	for _, key := range []string{"", "../photo.jpg", "nested/photo.jpg", `nested\photo.jpg`} {
		if err := store.Save(context.Background(), key, strings.NewReader("data")); err == nil {
			t.Fatalf("Save(%q) succeeded", key)
		}
	}
}
