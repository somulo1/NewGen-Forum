package push

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// apns sends to iPhones through Apple's push service (HTTP/2, token authentication with a .p8 key).
type apns struct {
	keyID, team, topic, host string
	key                      *ecdsa.PrivateKey
	http                     *http.Client

	mu     sync.Mutex
	bearer string
	made   time.Time
}

func newAPNs(pem []byte, keyID, team, topic string, sandbox bool, client *http.Client) (*apns, error) {
	key, err := jwt.ParseECPrivateKeyFromPEM(pem)
	if err != nil {
		return nil, fmt.Errorf("HUB_APNS_KEY: %w", err)
	}
	if keyID == "" || team == "" || topic == "" {
		return nil, fmt.Errorf("HUB_APNS_KEY_ID, HUB_APNS_TEAM_ID and HUB_APNS_TOPIC are needed with HUB_APNS_KEY")
	}
	host := "https://api.push.apple.com"
	if sandbox {
		host = "https://api.sandbox.push.apple.com"
	}
	return &apns{keyID: keyID, team: team, topic: topic, host: host, key: key, http: client}, nil
}

// Apple wants a fresh token at most every hour and no more than every 20 minutes.
func (a *apns) token(fresh bool) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !fresh && a.bearer != "" && time.Since(a.made) < 45*time.Minute {
		return a.bearer, nil
	}
	t := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{"iss": a.team, "iat": time.Now().Unix()})
	t.Header["kid"] = a.keyID
	s, err := t.SignedString(a.key)
	if err != nil {
		return "", err
	}
	a.bearer, a.made = s, time.Now()
	return s, nil
}

func (a *apns) send(ctx context.Context, token string, n Note) error {
	payload := map[string]any{"aps": map[string]any{
		"alert": map[string]string{"title": n.Title, "body": n.Body}, "sound": "default", "thread-id": n.Tag,
	}}
	for k, v := range n.Data {
		payload[k] = v
	}
	body, _ := json.Marshal(payload)
	expires := time.Now().Add(24 * time.Hour)
	if n.Urgent {
		expires = time.Now().Add(time.Minute)
	}
	for attempt := 0; attempt < 2; attempt++ {
		bearer, err := a.token(attempt > 0)
		if err != nil {
			return err
		}
		req, _ := http.NewRequestWithContext(ctx, "POST", a.host+"/3/device/"+token, bytes.NewReader(body))
		req.Header.Set("authorization", "bearer "+bearer)
		req.Header.Set("apns-topic", a.topic)
		req.Header.Set("apns-push-type", "alert")
		req.Header.Set("apns-priority", "10")
		req.Header.Set("apns-expiration", fmt.Sprint(expires.Unix()))
		if len(n.Tag) <= 64 {
			req.Header.Set("apns-collapse-id", n.Tag)
		}
		res, err := a.http.Do(req)
		if err != nil {
			return err
		}
		reply, _ := io.ReadAll(io.LimitReader(res.Body, 4<<10))
		res.Body.Close()
		switch {
		case res.StatusCode == 200:
			return nil
		case res.StatusCode == 403 && bytes.Contains(reply, []byte("ExpiredProviderToken")) && attempt == 0:
			continue
		case res.StatusCode == 410 || bytes.Contains(reply, []byte("BadDeviceToken")) || bytes.Contains(reply, []byte("Unregistered")):
			return errGone
		default:
			return fmt.Errorf("apns %d: %s", res.StatusCode, bytes.TrimSpace(reply))
		}
	}
	return fmt.Errorf("apns refused the provider token")
}
