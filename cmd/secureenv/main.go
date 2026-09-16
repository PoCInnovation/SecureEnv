// Command secureenv manages and synchronises project environment variables
// through the SecureEnv API.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/PoCInnovation/SecureEnv/internal/apiclient"
	"github.com/PoCInnovation/SecureEnv/internal/cli"
	"github.com/PoCInnovation/SecureEnv/internal/gitremote"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "secureenv:", err)
		return cli.ExitError
	}

	return cli.Run(ctx, os.Args[1:], cli.Env{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Getenv:  os.Getenv,
		Dir:     dir,
		Version: version,
		NewAPI: func(cfg cli.APIConfig) (cli.API, error) {
			opts := []apiclient.Option{apiclient.WithUserAgent("secureenv/" + version)}
			if cfg.CACertFile != "" {
				opts = append(opts, apiclient.WithCACertFile(cfg.CACertFile))
			}
			return apiclient.New(cfg.URL, cfg.Token, opts...)
		},
		OriginURL: gitremote.OriginURL,
	})
}
