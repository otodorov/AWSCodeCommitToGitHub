package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	git "github.com/libgit2/git2go/v31"
)

// Represents the `git clone` command.
func gitClone(user, password, url, path string) {
	fmt.Println("Cloning", url)
	credentialsCallback := func(url, username string, allowedTypes git.CredentialType) (*git.Credential, error) {
		credential, err := git.NewCredentialUserpassPlaintext(user, password)
		return credential, err
	}

	certificateCheckCallback := func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
		return 0
	}

	cloneOptions := &git.CloneOptions{
		CheckoutOpts: &git.CheckoutOptions{
			Strategy:       git.CheckoutForce,
			DisableFilters: true,
			DirMode:        0,
			FileMode:       0,
			FileOpenFlags:  0,
			NotifyFlags:    git.CheckoutNotifyAll,
		},
		FetchOptions: &git.FetchOptions{
			RemoteCallbacks: git.RemoteCallbacks{
				CredentialsCallback:      credentialsCallback,
				CertificateCheckCallback: certificateCheckCallback,
			},
			Prune:           0,
			UpdateFetchhead: false,
			DownloadTags:    0,
			ProxyOptions:    git.ProxyOptions{},
		},
		Bare: false,
	}

	if _, err := git.Clone(url, path, cloneOptions); err != nil {
		logHandler("debug", err.Error())
		return
	}
}

// Represents the `git add origin remote` command
func gitRemoteAddOriginURL(path, url string) *git.Repository {
	var repo *git.Repository
	var err error

	if repo, err = git.InitRepository(path, false); err != nil {
		logHandler("debug", err.Error())
	}
	if _, err = repo.Remotes.Create("origin", url); err != nil {
		logHandler("debug", err.Error())
	}
	return repo
}

// Represents the `git add` command.
func gitAdd(repo *git.Repository) {
	var idx *git.Index
	var err error

	if idx, err = repo.Index(); err != nil {
		logHandler("debug", err.Error())
	}

	if err = idx.AddAll([]string{}, git.IndexAddDefault, nil); err != nil {
		logHandler("debug", err.Error())
	}

	if err = idx.Write(); err != nil {
		logHandler("debug", err.Error())
	}
}

// Represents the `git commit` command.
func gitCommit(repo *git.Repository, msg, name, email string) {
	var idx *git.Index
	var objectId *git.Oid
	var treeId *git.Tree
	var err error

	if idx, err = repo.Index(); err != nil {
		logHandler("debug", err.Error())
	}

	if objectId, err = idx.WriteTreeTo(repo); err != nil {
		logHandler("debug", err.Error())
	}

	if treeId, err = repo.LookupTree(objectId); err != nil {
		logHandler("debug", err.Error())
	}

	signature := &git.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}

	if _, err = repo.CreateCommit(
		"HEAD",
		signature,
		signature,
		msg+" on "+time.Now().Format("2 Jan 2006"),
		treeId,
	); err != nil {
		logHandler("debug", err.Error())
	}
}

//Represents the `git push` command.
func gitPush(repo *git.Repository, repoName, user, password, branch, url string) {
	var remote *git.Remote
	var err error

	credentialsCallback := func(url, username string, allowedTypes git.CredentialType) (*git.Credential, error) {
		credential, err := git.NewCredentialUserpassPlaintext(user, password)
		return credential, err
	}

	certificateCheckCallback := func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
		return 0
	}

	pushOptions := &git.PushOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      credentialsCallback,
			CertificateCheckCallback: certificateCheckCallback,
		},
		PbParallelism: 0,
		Headers:       []string{},
	}

	if remote, err = repo.Remotes.Create(branch, url); err != nil {
		logHandler("debug", err.Error())
	}
	if err = remote.Push([]string{"refs/heads/" + branch}, pushOptions); err != nil {
		logHandler("debug", err.Error())
	}
	fmt.Printf("Pushing repo %q to GitHub\n", repoName)
}

// Execute `git add remote origin; git add; git commit; git push`
func gitRepo(url, user, pass, repoName, branch, message, author, email string, private bool) {
	var repoDir string
	var err error

	if repoDir, err = filepath.Abs(repoName); err != nil {
		logHandler("debug", err.Error())
	}

	if err := os.RemoveAll(repoDir + "/.git"); err != nil {
		logHandler("debug", err.Error())
	}

	githubRepo := gitRemoteAddOriginURL(repoDir, url)
	gitAdd(githubRepo)
	gitCommit(githubRepo, message, author, email)
	gitPush(githubRepo, repoName, user, pass, branch, url)
}
