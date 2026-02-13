package main

import (
	"context"
	"errors"
	"fmt"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/atproto/atclient"
	"github.com/bluesky-social/indigo/atproto/syntax"

	"github.com/urfave/cli/v3"
)

// cmdAccount is the account management subcommand group
var cmdAccount = &cli.Command{
	Name:  "account",
	Usage: "Auth session and account management",
	Commands: []*cli.Command{
		{
			Name:  "login",
			Usage: "Login with ATProto credentials",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"u"},
					Usage:    "Handle or DID",
					Required: true,
					Sources:  cli.EnvVars("ATP_USERNAME"),
				},
				&cli.StringFlag{
					Name:     "password",
					Aliases:  []string{"p"},
					Usage:    "App password",
					Required: true,
					Sources:  cli.EnvVars("ATP_PASSWORD"),
				},
				&cli.StringFlag{
					Name:    "pds-host",
					Usage:   "Override PDS URL",
					Sources: cli.EnvVars("ATP_PDS_HOST"),
				},
			},
			Action: runAccountLogin,
		},
		{
			Name:   "logout",
			Usage:  "Delete current session",
			Action: runAccountLogout,
		},
		{
			Name:   "status",
			Usage:  "Check login status",
			Action: runAccountStatus,
		},
	},
}

func runAccountLogin(ctx context.Context, cmd *cli.Command) error {
	var client *atclient.APIClient
	var err error

	pdsHost := cmd.String("pds-host")
	username := cmd.String("username")
	password := cmd.String("password")

	if pdsHost != "" {
		client, err = atclient.LoginWithPasswordHost(ctx, pdsHost, username, password, "", authRefreshCallback)
	} else {
		atid, parseErr := syntax.ParseAtIdentifier(username)
		if parseErr != nil {
			return fmt.Errorf("invalid username: %w", parseErr)
		}
		dir := configDirectory()
		client, err = atclient.LoginWithPassword(ctx, dir, atid, password, "", authRefreshCallback)
	}
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	passAuth, ok := client.Auth.(*atclient.PasswordAuth)
	if !ok {
		return fmt.Errorf("unexpected auth type")
	}

	// Get handle for display
	sessResp, err := comatproto.ServerGetSession(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get session info: %w", err)
	}

	sess := AuthSession{
		DID:          passAuth.Session.AccountDID,
		PDS:          passAuth.Session.Host,
		Handle:       sessResp.Handle,
		Password:     password,
		AccessToken:  passAuth.Session.AccessToken,
		RefreshToken: passAuth.Session.RefreshToken,
	}
	if err := persistAuthSession(&sess); err != nil {
		return fmt.Errorf("failed to persist session: %w", err)
	}

	fmt.Fprintf(cmd.Root().Writer, "Logged in as %s (%s)\n", sessResp.Handle, sessResp.Did)
	return nil
}

func runAccountLogout(ctx context.Context, cmd *cli.Command) error {
	err := wipeAuthSession()
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.Root().Writer, "Logged out")
	return nil
}

func runAccountStatus(ctx context.Context, cmd *cli.Command) error {
	client, err := loadAuthClient(ctx)
	if errors.Is(err, ErrNoAuthSession) {
		return fmt.Errorf("not logged in (run: hb account login)")
	}
	if err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	sessResp, err := comatproto.ServerGetSession(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	status, err := comatproto.ServerCheckAccountStatus(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to check account status: %w", err)
	}

	w := cmd.Root().Writer
	fmt.Fprintf(w, "DID:    %s\n", sessResp.Did)
	fmt.Fprintf(w, "Handle: %s\n", sessResp.Handle)
	fmt.Fprintf(w, "PDS:    %s\n", client.Host)

	if status.Activated {
		fmt.Fprintln(w, "Status: active")
	} else {
		fmt.Fprintln(w, "Status: deactivated")
	}

	return nil
}
