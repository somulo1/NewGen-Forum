// Package calltoken is the pass the API gives a device to enter one call room on the calls service. The two services share
// a secret and nothing else: the calls service needs no database access.
package calltoken

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	Room    string `json:"room"`
	Org     string `json:"org"`
	Person  string `json:"person"`
	Name    string `json:"name"`
	Role    string `json:"role"` // host | moderator | participant
	Kind    string `json:"kind"` // voice | video
	MaxPeer int    `json:"max"`
}

func Mint(secret []byte, c Claims, ttl time.Duration) (string, error) {
	now := time.Now()
	c.IssuedAt, c.ExpiresAt, c.Issuer = jwt.NewNumericDate(now), jwt.NewNumericDate(now.Add(ttl)), "hub-api"
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(secret)
}

var ErrInvalid = errors.New("calltoken: invalid or expired")

func Verify(secret []byte, token string) (*Claims, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("hub-api"), jwt.WithExpirationRequired(), jwt.WithLeeway(5*time.Second))
	if err != nil || c.Room == "" || c.Person == "" {
		return nil, ErrInvalid
	}
	return &c, nil
}
