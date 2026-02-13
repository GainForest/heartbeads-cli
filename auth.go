package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/adrg/xdg"
	"github.com/bluesky-social/indigo/atproto/syntax"
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
