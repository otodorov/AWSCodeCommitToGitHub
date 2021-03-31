package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode"

	"gopkg.in/yaml.v2"
)

type Config struct {
	GitHub struct {
		User     string `yaml:"username"`
		Pass     string `yaml:"password"`
		Private  bool   `yaml:"private"`
		Xthreads int    `yaml:"threads"`
		Commit   struct {
			Name  string `yaml:"name"`
			Email string `yaml:"email"`
		} `yaml:"commitMsg"`
	} `yaml:"GitHub"`
	AWSCodeCommit struct {
		User string `yaml:"username"`
		Pass string `yaml:"password"`
	} `yaml:"AWSCodeCommit"`
	AWSIAMAccesskeys struct {
		Access_key_id     string `yaml:"aws_access_key_id"`
		Secret_access_key string `yaml:"aws_secret_access_key"`
	} `yaml:"AWSIAMAccesskeys"`
	AWSRegion string `yaml:"AWSRegion"`
}

func main() {
	const (
		cfgFileName string = "AWSCodeCommitToGitHub.yml"
		AWSURL      string = "https://git-codecommit.%s.amazonaws.com/v1/repos/%s"
		GitHubURL   string = "https://github.com/%s/%s.git"
		branch      string = "master"
		message     string = "Migrated from AWS Codecommit"
	)

	var (
		conf *os.File
		pwd  string
		dir  string
		ch   = make(chan string, 50)
		wg   sync.WaitGroup
		err  error
	)

	// Read config.yml file
	configFile := Config{}
	if conf, err = os.Open(cfgFileName); err != nil {
		fmt.Printf("Cannot %s\n", err)
		return
	}

	defer func() {
		if err := conf.Close(); err != nil {
			fmt.Printf("Can't close configFile: %s", err)
		}
	}()

	// Decode the YAML file
	dec := yaml.NewDecoder(conf)
	if err = dec.Decode(&configFile); err != nil {
		fmt.Printf("Cannot read %q file. Is it empty?\n", cfgFileName)
		fmt.Println(err)
		return
	}

	// Create temp dir for `git clone`
	if pwd, err = os.Getwd(); err != nil {
		fmt.Println(err)
	}

	// Function that return symbol used to split the string
	file := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// Split string
	pwdS := strings.FieldsFunc(pwd, file)
	currentFolder := pwdS[len(pwdS)-1]
	if dir, err = os.MkdirTemp(os.TempDir(), currentFolder+"-"); err != nil {
		fmt.Println(err)
	}
	defer os.RemoveAll(dir)

	if cd := os.Chdir(dir); cd != nil {
		fmt.Println(cd)
	}

	ctx, cfg := awsClient(
		configFile.AWSIAMAccesskeys.Access_key_id,
		configFile.AWSIAMAccesskeys.Secret_access_key,
		configFile.AWSRegion,
	)

	awsRepoSlice := awsListRepositories(ctx, cfg)
	// awsRepoSlice := []string{"emp_terraform_security_group"}

	// Uncomment following two lines to delete repositories defined in githubDeleteRepos.go -> awsRepoSlice
	// githubDeleteRepos(configFile.GitHub.Pass, awsRepoSlice)
	// return

	fmt.Println(strings.Repeat("=", 100))
	wg.Add(configFile.GitHub.Xthreads)
	for i := 0; i < configFile.GitHub.Xthreads; i++ {
		go func() {
		Loop:
			for {
				repoName, ok := <-ch
				if !ok { // if there is nothing to do and the channel has been closed then end the goroutine
					wg.Done()
					return
				}
				codecommitRepoURL := fmt.Sprintf(AWSURL, configFile.AWSRegion, repoName)
				githubRepoURL := fmt.Sprintf(GitHubURL, configFile.GitHub.User, repoName)
				description := awsDescribeRepo(ctx, cfg, repoName)

				if err := githubCreateRepo(
					configFile.GitHub.Pass,
					repoName,
					*description,
					branch,
					configFile.GitHub.Private); err != nil {
					continue Loop
				}

				gitClone(
					configFile.AWSCodeCommit.User,
					configFile.AWSCodeCommit.Pass,
					codecommitRepoURL,
					repoName,
				)

				gitRepo(
					githubRepoURL,
					configFile.GitHub.User,
					configFile.GitHub.Pass,
					repoName,
					branch,
					message,
					configFile.GitHub.Commit.Name,
					configFile.GitHub.Commit.Email,
					configFile.GitHub.Private,
				)

				if err = os.RemoveAll(repoName); err != nil {
					fmt.Printf("Cannot delete folder %s: %s", repoName, err)
				}
			}
		}()
	}

	// Loop over the repository list and send each of them in the channel
	for _, repoName := range awsRepoSlice {
		ch <- repoName
	}
	close(ch) // This tells the goroutines there's nothing else to do
	wg.Wait() // Wait for the threads to finish
}
