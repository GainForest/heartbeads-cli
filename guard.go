package main

import (
	"errors"
	"fmt"
)

// requireAuth loads the auth session and returns a descriptive error if not logged in.
func requireAuth() (*AuthSession, error) {
	sess, err := loadAuthSessionFile()
	if err != nil {
		if errors.Is(err, ErrNoAuthSession) {
			return nil, fmt.Errorf("Not logged in. Run: hb account login --username <handle> --password <app-password>")
		}
		return nil, fmt.Errorf("failed to load auth session: %w", err)
	}
	return sess, nil
}
