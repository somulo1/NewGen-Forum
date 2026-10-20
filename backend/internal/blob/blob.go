// Package blob stores encrypted file contents. What it holds is ciphertext only: the key that opens a file never reaches it.
package blob

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var ErrTooLarge = errors.New("blob: larger than allowed")
var ErrNotFound = errors.New("blob: not found")

// Store is where the app keeps file contents. Put returns the stored reference (it names which file server holds the
// bytes); Get and Delete take that reference.
type Store interface {
	// Put streams r into key, refusing more than max bytes, and returns the reference, size and SHA-256 of what was stored.
	Put(ctx context.Context, key string, r io.Reader, max int64) (ref string, size int64, sum []byte, err error)
	Get(ctx context.Context, ref string) (io.ReadSeekCloser, int64, error)
	Delete(ctx context.Context, ref string) error
}

// Backend is one place bytes can live: this server's disk or a file server.
type Backend interface {
	PutFile(ctx context.Context, key, path string, size int64, sum []byte) error
	Get(ctx context.Context, key string) (io.ReadSeekCloser, int64, error)
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}

var keyRe = regexp.MustCompile(`^[a-z0-9][a-z0-9/_.-]{3,200}$`)

// Dir is a filesystem-backed store for development and single-node installs. Production uses an object store (S3, GCS)
// behind the same interface, with server-side encryption as a second layer.
type Dir struct{ Root string }

func (d Dir) path(key string) (string, error) {
	if !keyRe.MatchString(key) || filepath.Clean(key) != key {
		return "", fmt.Errorf("blob: bad key")
	}
	return filepath.Join(d.Root, key), nil
}

func (d Dir) Put(ctx context.Context, key string, r io.Reader, max int64) (int64, []byte, error) {
	p, err := d.path(key)
	if err != nil {
		return 0, nil, err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return 0, nil, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), ".up-*")
	if err != nil {
		return 0, nil, err
	}
	defer os.Remove(tmp.Name())
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), io.LimitReader(r, max+1))
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return 0, nil, err
	}
	if n > max {
		return 0, nil, ErrTooLarge
	}
	if err := os.Rename(tmp.Name(), p); err != nil {
		return 0, nil, err
	}
	return n, h.Sum(nil), nil
}

func (d Dir) Get(ctx context.Context, key string) (io.ReadSeekCloser, int64, error) {
	p, err := d.path(key)
	if err != nil {
		return nil, 0, err
	}
	f, err := os.Open(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, ErrNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	return f, st.Size(), nil
}

func (d Dir) Delete(ctx context.Context, key string) error {
	p, err := d.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// PutFile moves a spooled file into place (the spool lives under the same root, so this is a rename).
func (d Dir) PutFile(ctx context.Context, key, path string, size int64, sum []byte) error {
	p, err := d.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return err
	}
	if err := os.Link(path, p); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return copyFile(path, p)
		}
	}
	return nil
}

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp, err := os.CreateTemp(filepath.Dir(to), ".cp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), to)
}

// Ping checks the directory can be written.
func (d Dir) Ping(ctx context.Context) error {
	if err := os.MkdirAll(d.Root, 0o750); err != nil {
		return err
	}
	f, err := os.CreateTemp(d.Root, ".ping-*")
	if err != nil {
		return err
	}
	f.Close()
	return os.Remove(f.Name())
}
