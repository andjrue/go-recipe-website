package storage

import (
	"context"
	"errors"
	"io"
)

var ErrNotFound = errors.New("stored image not found")

type ImageStore interface {
	Save(context.Context, string, io.Reader) error
	Open(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}
