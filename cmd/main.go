package main

import (
	"context"
	"errors"
	"os"

	"github.com/doodlescheduling/create-repository/internal/repository"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	rootCmd = &cobra.Command{
		Use:           "gh-bootstrap-repository [repository-name] [origin-repository]",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("expects [repository-name] [origin-repository] as arguments")
			}

			name, fromRepository := args[0], args[1]

			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
			)

			tc := oauth2.NewClient(context.TODO(), ts)
			ghClient := github.NewClient(tc)

			manager := repository.New(ghClient)
			return manager.CreateRepository(context.TODO(), name, fromRepository)
		},
	}
)

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
