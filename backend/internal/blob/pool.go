package blob

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Pool keeps file contents on whichever file server is available, in the order they are configured, and on this
// server's own disk when none is configured or none answers. Each stored reference names where its bytes went
// ("fs1:o/<org>/<id>"; a bare key is on the local disk), so reads and deletes always reach the right place even after
// the configuration changes.
type Pool struct {
	Local  Dir
	names  []string           // write order: file servers first, "local" last
	stores map[string]Backend // by name
	mu     sync.Mutex
	down   map[string]time.Time // a server that failed is skipped for a while
}

const LocalName = "local"

// NewPool builds the pool from file server URLs (see ParseS3). With none, everything stays on the local disk.
func NewPool(localRoot string, servers []string) (*Pool, error) {
	p := &Pool{Local: Dir{Root: localRoot}, stores: map[string]Backend{}, down: map[string]time.Time{}}
	for i, raw := range servers {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		s, err := ParseS3(raw)
		if err != nil {
			return nil, err
		}
		name := fmt.Sprintf("fs%d", i+1)
		p.names = append(p.names, name)
		p.stores[name] = s
	}
	p.names = append(p.names, LocalName)
	p.stores[LocalName] = p.Local
	return p, nil
}

// Servers lists where new files can go, in order.
func (p *Pool) Servers() []string { return append([]string(nil), p.names...) }

func split(ref string) (name, key string) {
	if i := strings.IndexByte(ref, ':'); i > 0 {
		return ref[:i], ref[i+1:]
	}
	return LocalName, ref
}

func (p *Pool) backend(ref string) (Backend, string, error) {
	name, key := split(ref)
	b, ok := p.stores[name]
	if !ok {
		return nil, "", fmt.Errorf("blob: file server %q is no longer configured; add it back to reach its files", name)
	}
	return b, key, nil
}

func (p *Pool) isDown(name string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	t, ok := p.down[name]
	return ok && time.Since(t) < 30*time.Second
}

func (p *Pool) mark(name string, failed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if failed {
		p.down[name] = time.Now()
	} else {
		delete(p.down, name)
	}
}

// Put spools the upload to local disk (checking its size and hashing it), then hands it to the first file server that
// takes it, falling back to the local disk.
func (p *Pool) Put(ctx context.Context, key string, r io.Reader, max int64) (string, int64, []byte, error) {
	spool := filepath.Join(p.Local.Root, ".spool")
	if err := os.MkdirAll(spool, 0o750); err != nil {
		return "", 0, nil, err
	}
	tmp, err := os.CreateTemp(spool, "up-*")
	if err != nil {
		return "", 0, nil, err
	}
	defer os.Remove(tmp.Name())
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), io.LimitReader(r, max+1))
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return "", 0, nil, err
	}
	if n > max {
		return "", 0, nil, ErrTooLarge
	}
	sum := h.Sum(nil)
	var last error
	for _, name := range p.names {
		if name != LocalName && p.isDown(name) {
			continue
		}
		if err := p.stores[name].PutFile(ctx, key, tmp.Name(), n, sum); err != nil {
			slog.Warn("file server did not take the file; trying the next", "server", name, "err", err)
			p.mark(name, true)
			last = err
			continue
		}
		p.mark(name, false)
		if name == LocalName {
			return key, n, sum, nil
		}
		return name + ":" + key, n, sum, nil
	}
	return "", 0, nil, fmt.Errorf("blob: nowhere to store the file: %w", last)
}

func (p *Pool) Get(ctx context.Context, ref string) (io.ReadSeekCloser, int64, error) {
	b, key, err := p.backend(ref)
	if err != nil {
		return nil, 0, err
	}
	return b.Get(ctx, key)
}

// Delete removes the bytes for good. A missing object counts as removed.
func (p *Pool) Delete(ctx context.Context, ref string) error {
	b, key, err := p.backend(ref)
	if err != nil {
		return err
	}
	return b.Delete(ctx, key)
}

// Health reports each configured place and whether it answers.
func (p *Pool) Health(ctx context.Context) map[string]string {
	out := map[string]string{}
	for _, name := range p.names {
		if err := p.stores[name].Ping(ctx); err != nil {
			out[name] = err.Error()
		} else {
			out[name] = "ok"
		}
	}
	return out
}
