// Command loadtest drives a running Hub API with many signed-in people at once and reports latency and errors per route.
// Each virtual person signs in, opens a realtime socket and then loops over the reads a person does all day. Run it from a
// machine other than the server, against a staging deployment seeded with demo people (the same password for all).
//
//	go run ./cmd/loadtest -url https://staging.example.com -people 500 -duration 5m
//
// Release gate (see docs/GO_LIVE.md): error rate under 0.5%, p95 under 500 ms for reads, no socket drops, database pool never full.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

type result struct {
	d      time.Duration
	status int
	err    bool
}

var (
	mu    sync.Mutex
	stats = map[string][]result{}
)

func record(route string, d time.Duration, status int, err error) {
	mu.Lock()
	stats[route] = append(stats[route], result{d, status, err != nil})
	mu.Unlock()
}

func main() {
	base := flag.String("url", "http://127.0.0.1:8080", "API address")
	people := flag.Int("people", 100, "virtual people (signed in at once)")
	dur := flag.Duration("duration", time.Minute, "how long to run")
	file := flag.String("people-file", "seed/data/dir.json", "demo people (JSON list with an email field)")
	pass := flag.String("password", "Hub!Demo2026", "their shared password")
	org := flag.String("org", "demo", "organization slug")
	think := flag.Duration("think", 2*time.Second, "average pause between a person's actions")
	sockets := flag.Bool("sockets", true, "also hold a realtime socket per person")
	spread := flag.Bool("spread-ip", false, "send a distinct X-Forwarded-For per person (only honoured from trusted proxies; use against a local server)")
	flag.Parse()

	raw, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var dir []struct{ Email string }
	_ = json.Unmarshal(raw, &dir)
	if len(dir) < *people {
		*people = len(dir)
	}
	ctx, cancel := context.WithTimeout(context.Background(), *dur)
	defer cancel()
	var wg sync.WaitGroup
	var socketsUp, socketDrops, loginFail atomic.Int64
	start := time.Now()
	for i := 0; i < *people; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond) // ramp up
			ip := ""
			if *spread {
				ip = fmt.Sprintf("10.%d.%d.%d", i/65536%256, i/256%256, i%256+1)
			}
			c := &client{base: *base, ip: ip, h: &http.Client{Timeout: 20 * time.Second}}
			if err := c.login(dir[i].Email, *pass, *org); err != nil {
				loginFail.Add(1)
				return
			}
			if *sockets {
				go c.socket(ctx, &socketsUp, &socketDrops)
			}
			reads := []string{"/me", "/conversations?limit=30", "/people?limit=20", "/tickets?mine=1", "/notifications", "/approvals?limit=20", "/org/structure", "/calls/history?limit=20", "/comms", "/projects"}
			for ctx.Err() == nil {
				p := reads[rand.Intn(len(reads))]
				c.get(p)
				select {
				case <-ctx.Done():
				case <-time.After(time.Duration(float64(*think) * (0.5 + rand.Float64()))):
				}
			}
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("\n%d people for %s against %s. Sign-in failures: %d. Sockets opened: %d, dropped early: %d.\n\n", *people, elapsed.Round(time.Second), *base, loginFail.Load(), socketsUp.Load(), socketDrops.Load())
	fmt.Printf("%-28s %8s %8s %8s %8s %8s %8s\n", "route", "requests", "errors", "429", "p50 ms", "p95 ms", "p99 ms")
	var all []result
	routes := make([]string, 0, len(stats))
	for r := range stats {
		routes = append(routes, r)
	}
	sort.Strings(routes)
	bad, limited := 0, 0
	for _, r := range routes {
		rs := stats[r]
		all = append(all, rs...)
		e, l := 0, 0
		ds := make([]time.Duration, 0, len(rs))
		for _, x := range rs {
			ds = append(ds, x.d)
			if x.err || x.status >= 500 {
				e++
			}
			if x.status == 429 {
				l++
			}
		}
		bad += e
		limited += l
		sort.Slice(ds, func(i, j int) bool { return ds[i] < ds[j] })
		pc := func(p float64) float64 { return float64(ds[int(float64(len(ds)-1)*p)].Microseconds()) / 1000 }
		fmt.Printf("%-28s %8d %8d %8d %8.0f %8.0f %8.0f\n", r, len(rs), e, l, pc(.5), pc(.95), pc(.99))
	}
	rate := 0.0
	if len(all) > 0 {
		rate = float64(bad) / float64(len(all)) * 100
	}
	fmt.Printf("\n%d requests, %.1f per second, %.2f%% server errors, %d rate-limited.\n", len(all), float64(len(all))/elapsed.Seconds(), rate, limited)
	if rate > 0.5 || loginFail.Load() > int64(*people)/50 || socketDrops.Load() > 0 {
		fmt.Println("RESULT: FAIL (gate: errors under 0.5%, sign-in failures under 2%, no dropped sockets)")
		os.Exit(1)
	}
	fmt.Println("RESULT: PASS")
}

type client struct {
	base, ip, token string
	h               *http.Client
}

func (c *client) do(method, path string, body any) (*http.Response, time.Duration, error) {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.base+"/v1"+path, rd)
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.ip != "" {
		req.Header.Set("X-Forwarded-For", c.ip)
	}
	t := time.Now()
	res, err := c.h.Do(req)
	return res, time.Since(t), err
}

func (c *client) login(email, pass, org string) error {
	res, d, err := c.do("POST", "/auth/login", map[string]string{"email": email, "password": pass, "org": org})
	if err != nil {
		record("POST /auth/login", d, 0, err)
		return err
	}
	defer res.Body.Close()
	record("POST /auth/login", d, res.StatusCode, nil)
	var t struct {
		Access string `json:"access_token"`
	}
	_ = json.NewDecoder(res.Body).Decode(&t)
	if res.StatusCode != 200 || t.Access == "" {
		return fmt.Errorf("sign-in %d", res.StatusCode)
	}
	c.token = t.Access
	return nil
}

func (c *client) get(path string) {
	res, d, err := c.do("GET", path, nil)
	route := "GET " + strings.SplitN(path, "?", 2)[0]
	if err != nil {
		record(route, d, 0, err)
		return
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	record(route, d, res.StatusCode, nil)
}

func (c *client) socket(ctx context.Context, up, drops *atomic.Int64) {
	res, _, err := c.do("POST", "/realtime/ticket", nil)
	if err != nil {
		return
	}
	var t struct{ Ticket string }
	_ = json.NewDecoder(res.Body).Decode(&t)
	res.Body.Close()
	u := "ws" + strings.TrimPrefix(c.base, "http") + "/v1/realtime?ticket=" + t.Ticket
	conn, _, err := websocket.Dial(ctx, u, nil)
	if err != nil {
		drops.Add(1)
		return
	}
	up.Add(1)
	defer conn.CloseNow()
	go func() {
		tk := time.NewTicker(30 * time.Second)
		defer tk.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				_ = conn.Write(ctx, websocket.MessageText, []byte(`{"op":"ping"}`))
			}
		}
	}()
	for {
		if _, _, err := conn.Read(ctx); err != nil {
			if ctx.Err() == nil {
				drops.Add(1)
			}
			return
		}
	}
}
