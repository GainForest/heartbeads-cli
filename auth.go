package main

import (
	"context"
	"encoding/json"
	"errors"
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

// AuthSession represents a persisted authentication session
type AuthSession struct {
	DID          syntax.DID `json:"did"`
	PDS          string     `json:"pds"`
	Handle       string     `json:"handle"`
	Password     string     `json:"password"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
}

// persistAuthSession saves the auth session to XDG state directory
func persistAuthSession(sess *AuthSession) error {
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

// loadAuthSessionFile loads the auth session from XDG state directory
func loadAuthSessionFile() (*AuthSession, error) {
	fPath, err := xdg.SearchStateFile("heartbeads/auth-session.json")
	if err != nil {
		return nil, ErrNoAuthSession
	}

	fBytes, err := os.ReadFile(fPath)
	if err != nil {
		return nil, err
	}

	var sess AuthSession
	err = json.Unmarshal(fBytes, &sess)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// wipeAuthSession deletes the auth session file
func wipeAuthSession() error {
	fPath, err := xdg.SearchStateFile("heartbeads/auth-session.json")
	if err != nil {
		// File doesn't exist, nothing to wipe
		return nil
	}
	return os.Remove(fPath)
}

// authRefreshCallback is called when tokens are refreshed
func authRefreshCallback(ctx context.Context, data atclient.PasswordSessionData) {
	sess, _ := loadAuthSessionFile()
	if sess == nil {
		sess = &AuthSession{}
	}

	sess.DID = data.AccountDID
	sess.AccessToken = data.AccessToken
	sess.RefreshToken = data.RefreshToken
	sess.PDS = data.Host

	if err := persistAuthSession(sess); err != nil {
		slog.Warn("failed to save refreshed auth session data", "err", err)
	}
}

// configDirectory returns an identity directory
func configDirectory() identity.Directory {
	return identity.DefaultDirectory()
}

// loadAuthClient loads an auth client from the saved session
func loadAuthClient(ctx context.Context) (*atclient.APIClient, error) {
	sess, err := loadAuthSessionFile()
	if err != nil {
		return nil, err
	}

	// First try to resume session
	client := atclient.ResumePasswordSession(atclient.PasswordSessionData{
		AccessToken:  sess.AccessToken,
		RefreshToken: sess.RefreshToken,
		AccountDID:   sess.DID,
		Host:         sess.PDS,
	}, authRefreshCallback)

	// Check that auth is working
	_, err = comatproto.ServerGetSession(ctx, client)
	if err == nil {
		return client, nil
	}

	// Otherwise try new auth session using saved password
	plcHost := os.Getenv("ATP_PLC_HOST")
	if plcHost == "" {
		plcHost = "https://plc.directory"
	}
	dir := configDirectory()
	return atclient.LoginWithPassword(ctx, dir, sess.DID.AtIdentifier(), sess.Password, "", authRefreshCallback)
}

// getLoggedInHandle returns the ATProto handle of the logged-in user.
// Returns ErrNoAuthSession if not logged in.
func getLoggedInHandle() (string, error) {
	sess, err := loadAuthSessionFile()
	if err != nil {
		return "", err
	}
	return sess.Handle, nil
}
