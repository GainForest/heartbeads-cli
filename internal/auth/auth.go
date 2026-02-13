package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/atclient"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"

	"github.com/adrg/xdg"
)

// ErrNoAuthSession is returned when no auth session file is found
var ErrNoAuthSession = errors.New("no auth session found")

// Session represents a persisted authentication session
type Session struct {
	DID          syntax.DID `json:"did"`
	PDS          string     `json:"pds"`
	Handle       string     `json:"handle"`
	Password     string     `json:"password"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
}

// PersistSession saves the auth session to XDG state directory
func PersistSession(sess *Session) error {
	fPath, err := xdg.StateFile("heartbeads/auth-session.json")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(fPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	authBytes, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(authBytes)
	return err
}

// LoadSessionFile loads the auth session from XDG state directory
func LoadSessionFile() (*Session, error) {
	fPath, err := xdg.SearchStateFile("heartbeads/auth-session.json")
	if err != nil {
		return nil, ErrNoAuthSession
	}

	fBytes, err := os.ReadFile(fPath)
	if err != nil {
		return nil, err
	}

	var sess Session
	err = json.Unmarshal(fBytes, &sess)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// WipeSession deletes the auth session file
func WipeSession() error {
	fPath, err := xdg.SearchStateFile("heartbeads/auth-session.json")
	if err != nil {
		// File doesn't exist, nothing to wipe
		return nil
	}
	return os.Remove(fPath)
}

// AuthRefreshCallback is called when tokens are refreshed
func AuthRefreshCallback(ctx context.Context, data atclient.PasswordSessionData) {
	sess, _ := LoadSessionFile()
	if sess == nil {
		sess = &Session{}
	}

	sess.DID = data.AccountDID
	sess.AccessToken = data.AccessToken
	sess.RefreshToken = data.RefreshToken
	sess.PDS = data.Host

	if err := PersistSession(sess); err != nil {
		slog.Warn("failed to save refreshed auth session data", "err", err)
	}
}

// ConfigDirectory returns an identity directory
func ConfigDirectory() identity.Directory {
	return identity.DefaultDirectory()
}

// LoadClient loads an auth client from the saved session
func LoadClient(ctx context.Context) (*atclient.APIClient, error) {
	sess, err := LoadSessionFile()
	if err != nil {
		return nil, err
	}

	// First try to resume session
	client := atclient.ResumePasswordSession(atclient.PasswordSessionData{
		AccessToken:  sess.AccessToken,
		RefreshToken: sess.RefreshToken,
		AccountDID:   sess.DID,
		Host:         sess.PDS,
	}, AuthRefreshCallback)

	// Check that auth is working
	_, err = comatproto.ServerGetSession(ctx, client)
	if err == nil {
		return client, nil
	}

	// Otherwise try new auth session using saved password
	dir := ConfigDirectory()
	return atclient.LoginWithPassword(ctx, dir, sess.DID.AtIdentifier(), sess.Password, "", AuthRefreshCallback)
}

// GetLoggedInHandle returns the ATProto handle of the logged-in user.
// Returns ErrNoAuthSession if not logged in.
func GetLoggedInHandle() (string, error) {
	sess, err := LoadSessionFile()
	if err != nil {
		return "", err
	}
	return sess.Handle, nil
}

// RequireAuth loads the auth session and returns a descriptive error if not logged in.
func RequireAuth() (*Session, error) {
	sess, err := LoadSessionFile()
	if err != nil {
		if errors.Is(err, ErrNoAuthSession) {
			return nil, fmt.Errorf("Not logged in. Run: hb account login --username <handle> --password <app-password>")
		}
		return nil, fmt.Errorf("failed to load auth session: %w", err)
	}
	return sess, nil
}
