package push

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const fcmScope = "https://www.googleapis.com/auth/firebase.messaging"

// errGone means the platform no longer knows this address: it is removed and never tried again.
var errGone = errors.New("push address no longer valid")

// fcm sends to Android phones through Firebase Cloud Messaging (HTTP v1), signed in with a Google service account.
type fcm struct {
	project, email, tokenURI, api string
	key                           *rsa.PrivateKey
	http                          *http.Client

	mu     sync.Mutex
	access string
	until  time.Time
}

func newFCM(credentials []byte, client *http.Client) (*fcm, error) {
	var sa struct {
		ProjectID   string `json:"project_id"`
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
		TokenURI    string `json:"token_uri"`
	}
	if err := json.Unmarshal(credentials, &sa); err != nil || sa.ProjectID == "" || sa.ClientEmail == "" {
		return nil, fmt.Errorf("HUB_FCM_CREDENTIALS is not a Firebase service account key")
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(sa.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("HUB_FCM_CREDENTIALS private key: %w", err)
	}
	if sa.TokenURI == "" {
		sa.TokenURI = "https://oauth2.googleapis.com/token"
	}
	return &fcm{project: sa.ProjectID, email: sa.ClientEmail, tokenURI: sa.TokenURI, api: "https://fcm.googleapis.com", key: key, http: client}, nil
}

func (f *fcm) token(ctx context.Context, fresh bool) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !fresh && f.access != "" && time.Now().Before(f.until) {
		return f.access, nil
	}
	now := time.Now()
	assertion, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": f.email, "scope": fcmScope, "aud": f.tokenURI, "iat": now.Unix(), "exp": now.Add(time.Hour).Unix(),
	}).SignedString(f.key)
	if err != nil {
		return "", err
	}
	form := url.Values{"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"}, "assertion": {assertion}}
	req, _ := http.NewRequestWithContext(ctx, "POST", f.tokenURI, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := f.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if res.StatusCode != 200 || json.NewDecoder(res.Body).Decode(&out) != nil || out.AccessToken == "" {
		return "", fmt.Errorf("google sign-in for FCM failed (%d)", res.StatusCode)
	}
	f.access, f.until = out.AccessToken, now.Add(time.Duration(out.ExpiresIn)*time.Second-5*time.Minute)
	return f.access, nil
}

func (f *fcm) send(ctx context.Context, token string, n Note) error {
	ttl := "86400s"
	if n.Urgent {
		ttl = "60s"
	}
	body, _ := json.Marshal(map[string]any{"message": map[string]any{
		"token":        token,
		"notification": map[string]string{"title": n.Title, "body": n.Body},
		"data":         n.Data,
		"android": map[string]any{
			"priority": "HIGH", "collapse_key": n.Tag, "ttl": ttl,
			"notification": map[string]any{
				"channel_id": channelID, "tag": n.Tag, "color": accent, "sound": "default", "default_vibrate_timings": true,
				// shown on the lock screen; its text is hidden there when the person chose to hide sensitive content
				"notification_priority": "PRIORITY_HIGH", "visibility": "PRIVATE",
			},
		},
	}})
	for attempt := 0; attempt < 2; attempt++ {
		access, err := f.token(ctx, attempt > 0)
		if err != nil {
			return err
		}
		req, _ := http.NewRequestWithContext(ctx, "POST", f.api+"/v1/projects/"+f.project+"/messages:send", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+access)
		req.Header.Set("Content-Type", "application/json")
		res, err := f.http.Do(req)
		if err != nil {
			return err
		}
		reply, _ := io.ReadAll(io.LimitReader(res.Body, 8<<10))
		res.Body.Close()
		switch {
		case res.StatusCode == 200:
			return nil
		case res.StatusCode == 401 && attempt == 0:
			continue
		case res.StatusCode == 404 || (res.StatusCode == 400 && bytes.Contains(reply, []byte("registration token"))) || bytes.Contains(reply, []byte("UNREGISTERED")):
			return errGone
		default:
			return fmt.Errorf("fcm %d: %s", res.StatusCode, bytes.TrimSpace(reply))
		}
	}
	return fmt.Errorf("fcm refused the sign-in")
}
