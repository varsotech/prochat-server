package filestore

import (
	"context"
	"io"
	"path"
)

// Scope defines a file store that is scoped to a specific bucket and key prefix.
// It is safe for concurrent use by multiple Go routines.
type Scope struct {
	store  FileStore
	prefix string
}

type FileStore interface {
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
	PutObject(ctx context.Context, key string, data io.Reader) (string, error)
}

func NewScope(store FileStore, prefix string) *Scope {
	return &Scope{
		store:  store,
		prefix: prefix,
	}
}

func (s *Scope) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.store.GetObject(ctx, s.buildKey(key))
}

func (s *Scope) PutObject(ctx context.Context, key string, data io.Reader) (string, error) {
	return s.store.PutObject(ctx, s.buildKey(key), data)
}

func (s *Scope) buildKey(key string) string {
	return path.Join(s.prefix, key)
}
