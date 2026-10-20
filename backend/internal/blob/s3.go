package blob

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// S3 is a file server that speaks the S3 API: AWS S3, MinIO, Ceph, Wasabi, Backblaze B2, Cloudflare R2 and most
// on-premise object stores. Requests are signed with AWS Signature V4 using only the standard library.
//
// Configured as a URL:  s3://ACCESS_KEY:SECRET_KEY@host[:port]/bucket?region=us-east-1&tls=1
// (tls=0 for a plain-HTTP server inside a private network; requests always use path-style addressing).
type S3 struct {
	Endpoint  string // scheme://host[:port]
	Host      string
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	Client    *http.Client
}

func ParseS3(raw string) (*S3, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "s3" || u.Host == "" || u.User == nil {
		return nil, fmt.Errorf("blob: a file server is s3://ACCESS_KEY:SECRET_KEY@host/bucket?region=...")
	}
	secret, _ := u.User.Password()
	bucket := strings.Trim(u.Path, "/")
	if bucket == "" || strings.Contains(bucket, "/") {
		return nil, fmt.Errorf("blob: the file server URL needs one bucket name")
	}
	q := u.Query()
	scheme := "https"
	if q.Get("tls") == "0" {
		scheme = "http"
	}
	region := q.Get("region")
	if region == "" {
		region = "us-east-1"
	}
	return &S3{Endpoint: scheme + "://" + u.Host, Host: u.Host, Bucket: bucket, Region: region, AccessKey: u.User.Username(), SecretKey: secret,
		Client: &http.Client{Timeout: 10 * time.Minute}}, nil
}

func (s *S3) objectPath(key string) string { return "/" + s.Bucket + "/" + key }

func hexSHA(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func hmacSHA(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(data))
	return m.Sum(nil)
}

// sign adds AWS Signature V4 headers. payloadHash is the hex SHA-256 of the body, or UNSIGNED-PAYLOAD.
func (s *S3) sign(req *http.Request, payloadHash string) {
	now := time.Now().UTC()
	amz := now.Format("20060102T150405Z")
	day := now.Format("20060102")
	req.Header.Set("x-amz-date", amz)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Host = s.Host
	names := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	if req.Header.Get("Range") != "" {
		names = append(names, "range")
	}
	sort.Strings(names)
	var canonHeaders strings.Builder
	for _, n := range names {
		v := req.Header.Get(n)
		if n == "host" {
			v = s.Host
		}
		canonHeaders.WriteString(n + ":" + strings.TrimSpace(v) + "\n")
	}
	signed := strings.Join(names, ";")
	q := req.URL.Query()
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, uriEncode(k, true)+"="+uriEncode(q.Get(k), true))
	}
	canon := strings.Join([]string{req.Method, uriEncode(req.URL.Path, false), strings.Join(parts, "&"), canonHeaders.String(), signed, payloadHash}, "\n")
	scope := day + "/" + s.Region + "/s3/aws4_request"
	toSign := "AWS4-HMAC-SHA256\n" + amz + "\n" + scope + "\n" + hexSHA([]byte(canon))
	k := hmacSHA([]byte("AWS4"+s.SecretKey), day)
	k = hmacSHA(k, s.Region)
	k = hmacSHA(k, "s3")
	k = hmacSHA(k, "aws4_request")
	sig := hex.EncodeToString(hmacSHA(k, toSign))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+s.AccessKey+"/"+scope+", SignedHeaders="+signed+", Signature="+sig)
}

func uriEncode(s string, encodeSlash bool) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~' || (c == '/' && !encodeSlash) {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

func (s *S3) do(ctx context.Context, method, path string, body io.Reader, size int64, payloadHash string, hdr map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.Endpoint+path, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.ContentLength = size
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	s.sign(req, payloadHash)
	return s.Client.Do(req)
}

func failed(res *http.Response, what string) error {
	b, _ := io.ReadAll(io.LimitReader(res.Body, 512))
	res.Body.Close()
	return fmt.Errorf("blob: file server %s: %s %s", what, res.Status, strings.TrimSpace(string(b)))
}

// PutFile uploads a spooled file of known size.
func (s *S3) PutFile(ctx context.Context, key, path string, size int64, sum []byte) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	res, err := s.do(ctx, http.MethodPut, s.objectPath(key), f, size, hex.EncodeToString(sum), map[string]string{"Content-Type": "application/octet-stream"})
	if err != nil {
		return err
	}
	if res.StatusCode/100 != 2 {
		return failed(res, "upload")
	}
	res.Body.Close()
	return nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	res, err := s.do(ctx, http.MethodDelete, s.objectPath(key), nil, 0, hexSHA(nil), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 && res.StatusCode != 404 {
		return failed(res, "delete")
	}
	return nil
}

// Ping checks the bucket answers.
func (s *S3) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	res, err := s.do(ctx, http.MethodHead, "/"+s.Bucket, nil, 0, hexSHA(nil), nil)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("blob: file server bucket check: %s", res.Status)
	}
	return nil
}

func (s *S3) Get(ctx context.Context, key string) (io.ReadSeekCloser, int64, error) {
	res, err := s.do(ctx, http.MethodHead, s.objectPath(key), nil, 0, hexSHA(nil), nil)
	if err != nil {
		return nil, 0, err
	}
	res.Body.Close()
	if res.StatusCode == 404 {
		return nil, 0, ErrNotFound
	}
	if res.StatusCode/100 != 2 {
		return nil, 0, fmt.Errorf("blob: file server read: %s", res.Status)
	}
	size, _ := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
	return &s3Reader{s: s, ctx: ctx, key: key, size: size}, size, nil
}

// s3Reader reads an object from an offset, reopening with a Range request after a Seek, so downloads can resume.
type s3Reader struct {
	s    *S3
	ctx  context.Context
	key  string
	size int64
	off  int64
	body io.ReadCloser
}

func (r *s3Reader) Read(p []byte) (int, error) {
	if r.off >= r.size {
		return 0, io.EOF
	}
	if r.body == nil {
		res, err := r.s.do(r.ctx, http.MethodGet, r.s.objectPath(r.key), nil, 0, hexSHA(nil), map[string]string{"Range": fmt.Sprintf("bytes=%d-", r.off)})
		if err != nil {
			return 0, err
		}
		if res.StatusCode/100 != 2 {
			return 0, failed(res, "read")
		}
		r.body = res.Body
	}
	n, err := r.body.Read(p)
	r.off += int64(n)
	return n, err
}

func (r *s3Reader) Seek(offset int64, whence int) (int64, error) {
	var to int64
	switch whence {
	case io.SeekStart:
		to = offset
	case io.SeekCurrent:
		to = r.off + offset
	case io.SeekEnd:
		to = r.size + offset
	}
	if to < 0 {
		return 0, errors.New("blob: negative position")
	}
	if to != r.off && r.body != nil {
		r.body.Close()
		r.body = nil
	}
	r.off = to
	return to, nil
}

func (r *s3Reader) Close() error {
	if r.body != nil {
		return r.body.Close()
	}
	return nil
}
