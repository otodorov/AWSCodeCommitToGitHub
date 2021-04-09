package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func githubCreateRepo(gitPass, repo, desc, branch string, private bool) error {
	var err error

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitPass},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	gitRepo := &github.Repository{
		Name:          github.String(repo),
		Private:       github.Bool(private),
		DefaultBranch: github.String(branch),
		MasterBranch:  github.String(branch),
		Description:   aws.String(desc),
	}
	_, _, err = client.Repositories.Create(ctx, "", gitRepo)
	if err != nil {
		fmt.Printf("!!! %q alredy exist in GitHub !!!\n", repo)
		fmt.Println("Skipping...")
	} else {
		fmt.Printf("Creating %q repository in GitHub\n", repo)
	}
	return err
}
