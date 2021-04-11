package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Used to delete list of repositories from GitHub
func githubDeleteRepos(gat string, awsRepoSlice []string) {

	// List of repositories that has to be deleted
	// awsRepoSlice = []string{"AWS_Ansible", "AWS_CodeDeploy"}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gat},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	for _, v := range awsRepoSlice {
		_, err := client.Repositories.Delete(ctx, "otodorov", v)
		if err != nil {
			fmt.Printf("%-10s: %s\n", v, err)
		} else {
			fmt.Printf("%-10s: deleted\n", v)
		}
	}

}
