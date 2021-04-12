package main

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/codecommit"
)

// Client used to communicate with AWS
func awsClient(accessKey, secretKey, region string) (c context.Context, cfg aws.Config) {
	// Context for the Client connection
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     accessKey,
					SecretAccessKey: secretKey,
				},
			},
		),
	)
	if err != nil {
		logHandler("debug", err.Error())
	}

	return ctx, cfg
}

// List all AWS CodeCommit repositories
func awsListRepositories(ctx context.Context, cfg aws.Config) []string {
	var repo *codecommit.ListRepositoriesOutput
	var repoSlice []string
	var err error

	client := codecommit.NewFromConfig(cfg)
	if repo, err = client.ListRepositories(ctx, &codecommit.ListRepositoriesInput{}); err != nil {
		logHandler("debug", err.Error())
	}
	for _, v := range repo.Repositories {
		repoSlice = append(repoSlice, aws.ToString(v.RepositoryName))
	}
	return repoSlice
}

// Gather repository description
func awsDescribeRepo(ctx context.Context, cfg aws.Config, repo string) *string {
	var description *string
	var dr *codecommit.GetRepositoryOutput
	var err error

	client := codecommit.NewFromConfig(cfg)
	if dr, err = client.GetRepository(ctx, &codecommit.GetRepositoryInput{
		RepositoryName: aws.String(repo),
	}); err != nil {
		logHandler("debug", err.Error())
	}

	// If description is empty, add empty string instead
	if description = dr.RepositoryMetadata.RepositoryDescription; description == nil {
		return aws.String("")
	}

	// Remove new line char '\n' -> Dec(10) Hx(A) Oct(12) Char(LF) (NL line feed, new line)
	*description = strings.Replace(*description, "\n", "", -1)
	return description
}
