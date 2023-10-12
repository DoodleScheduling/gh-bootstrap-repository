package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/doodlescheduling/gh-bootstrap-repository/internal/repository"
	"github.com/google/go-github/v55/github"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"golang.org/x/oauth2"
)

var (
	host    = "github.com"
	rootCmd = &cobra.Command{
		Use:           "gh-bootstrap-repository [owner/repository-name] [owner/origin-repository]",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flag.Parse()

			if len(args) != 2 {
				return errors.New("expects [owner/repository-name] [owner/origin-repository] as arguments")
			}

			name, fromRepository := args[0], args[1]
			token, _ := auth.TokenForHost(host)

			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			)

			tc := oauth2.NewClient(context.TODO(), ts)
			ghClient := github.NewClient(tc)

			manager := repository.New(ghClient)
			return manager.CreateRepository(context.TODO(), name, fromRepository)
		},
	}
)

func main() {
	flag.StringVar(&host, "host", host, "The github host (github.com)")

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n\n")
		_ = rootCmd.Help()
		os.Exit(1)
	}
}
