package main

import (
	"fmt"
	"os"
	"time"

	git "github.com/libgit2/git2go/v31"
)

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
		fmt.Println(err)
		return
	}
}

func gitRemoteAddOriginURL(path, url string) *git.Repository {
	var repo *git.Repository
	var err error

	if repo, err = git.InitRepository(path, false); err != nil {
		fmt.Println(err)
	}
	if _, err = repo.Remotes.Create("origin", url); err != nil {
		fmt.Println(err)
	}
	return repo
}

func gitAdd(repo *git.Repository) {
	var idx *git.Index
	var err error

	if idx, err = repo.Index(); err != nil {
		fmt.Println(err)
	}

	if err = idx.AddAll([]string{}, git.IndexAddDefault, nil); err != nil {
		fmt.Println(err)
	}

	if err = idx.Write(); err != nil {
		fmt.Println(err)
	}
}

func gitCommit(repo *git.Repository, msg, name, email string) {
	var idx *git.Index
	var objectId *git.Oid
	var treeId *git.Tree
	var err error

	if idx, err = repo.Index(); err != nil {
		fmt.Println(err)
	}

	if objectId, err = idx.WriteTreeTo(repo); err != nil {
		fmt.Println(err)
	}

	if treeId, err = repo.LookupTree(objectId); err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
	}
}

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
		fmt.Println(err)
	}
	if err = remote.Push([]string{"refs/heads/" + branch}, pushOptions); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Pushing repo %q to GitHub\n", repoName)
}

func gitRepo(url, user, pass, repoName, description, branch, message, author, email string, private bool) {
	if err := githubCreateRepo(pass, repoName, description, branch, private); err != nil {
		return
	}

	if cd := os.Chdir(repoName); cd != nil {
		fmt.Println(cd)
	}

	os.RemoveAll(".git")
	githubRepo := gitRemoteAddOriginURL("./", url)
	gitAdd(githubRepo)
	gitCommit(githubRepo, message, author, email)
	gitPush(githubRepo, repoName, user, pass, branch, url)
}
