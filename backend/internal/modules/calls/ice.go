package calls

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"

	"hub/internal/tenant"
)

// ICEServer is one STUN or TURN server in the shape a browser's RTCPeerConnection takes.
type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// iceServers is what a browser needs to reach the other person across networks. STUN lets two browsers behind ordinary routers
// find each other; a TURN relay carries the call when they cannot (mobile networks, strict office firewalls). TURN is set up
// with HUB_TURN_URLS and either HUB_TURN_USER/HUB_TURN_PASS or, for a server of your own, HUB_TURN_SECRET (short-lived passwords).
func (m *Module) iceServers(p *tenant.Principal) []ICEServer {
	out := []ICEServer{{URLs: []string{"stun:stun.l.google.com:19302", "stun:stun1.l.google.com:19302", "stun:stun.cloudflare.com:3478"}}}
	cfg := m.A.Cfg
	if len(cfg.TurnURLs) == 0 {
		return out
	}
	t := ICEServer{URLs: cfg.TurnURLs, Username: cfg.TurnUser, Credential: cfg.TurnPass}
	if cfg.TurnSecret != "" {
		// the coturn "REST API" scheme: the user name carries the expiry, the password is its HMAC under the shared secret
		t.Username = fmt.Sprintf("%d:%s", time.Now().Add(2*time.Hour).Unix(), p.Person)
		mac := hmac.New(sha1.New, []byte(cfg.TurnSecret))
		mac.Write([]byte(t.Username))
		t.Credential = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	}
	if t.Username != "" && t.Credential != "" {
		out = append(out, t)
	}
	return out
}
