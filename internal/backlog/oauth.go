package backlog

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// OAuthClient is the registered Nulab OAuth application (operator config).
type OAuthClient struct {
	ClientID     string
	ClientSecret string
}

// String hides the client secret so it can never leak through %v.
func (c OAuthClient) String() string {
	return fmt.Sprintf("backlog.OAuthClient{ClientID: %s}", c.ClientID)
}

// GoString hides the client secret for %#v as well.
func (c OAuthClient) GoString() string { return c.String() }

// TokenSet is the result of a code exchange or a refresh (contract C1).
type TokenSet struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// String hides both tokens.
func (t TokenSet) String() string {
	return fmt.Sprintf("backlog.TokenSet{ExpiresAt: %s}", t.ExpiresAt.Format(time.RFC3339))
}

// GoString hides both tokens for %#v as well.
func (t TokenSet) GoString() string { return t.String() }

// parseTokenSet decodes a token response. Anything that is not a Bearer token
// with both tokens and a positive lifetime is Unreachable.
func parseTokenSet(body []byte, now time.Time) (TokenSet, error) {
	var raw struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    *int64 `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &raw); err != nil || raw.AccessToken == "" || raw.RefreshToken == "" ||
		raw.ExpiresIn == nil || *raw.ExpiresIn <= 0 || !strings.EqualFold(raw.TokenType, "Bearer") {
		return TokenSet{}, &Error{Kind: KindUnreachable, Status: 200, Class: "body"}
	}
	return TokenSet{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		ExpiresAt:    now.Add(time.Duration(*raw.ExpiresIn) * time.Second),
	}, nil
}
