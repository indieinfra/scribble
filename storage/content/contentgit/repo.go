package contentgit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/indieinfra/scribble/config"
)

type repoState int

const (
	repoMissing repoState = iota
	repoNoPermission
	repoUnknownError
	repoNotDir
	repoNotGit
	repoWrongRemote
	repoValid
)

func inspectRepo(path, expectedRemote string) (repoState, *git.Repository, error) {
	info, err := os.Stat(path)
	if err != nil {
		switch true {
		case os.IsNotExist(err):
			log.Println("...repo missing")
			return repoMissing, nil, nil
		case os.IsPermission(err):
			log.Println("...no permission")
			return repoNoPermission, nil, fmt.Errorf("permission denied: %w", err)
		default:
			log.Println("...unknown error (cannot stat)")
			return repoUnknownError, nil, err
		}
	}

	if !info.IsDir() {
		log.Println("...is not a directory")
		return repoNotDir, nil, nil
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		log.Println("...is not a git repository")
		return repoNotGit, nil, nil
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		log.Println("...missing remote 'origin'")
		return repoWrongRemote, nil, nil
	}

	if !slices.Contains(remote.Config().URLs, expectedRemote) {
		log.Println("...origin URL does not match repository URL")
		return repoWrongRemote, nil, nil
	}

	log.Println("...ok!")
	return repoValid, repo, nil
}

func ensureCleanPath(path string, state repoState) error {
	switch state {
	case repoNotDir, repoNotGit, repoWrongRemote:
		log.Println("...removing existing data")
		return os.RemoveAll(path)
	case repoMissing, repoValid:
		log.Println("...ok!")
		return nil
	}

	log.Println("...unknown repo state (error)!")
	return fmt.Errorf("unknown state %v", state)
}

func BuildGitAuth(cfg *config.GitContentStrategy) (transport.AuthMethod, error) {
	switch cfg.Auth.Method {
	case "plain":
		log.Println("...using basic auth")
		return &http.BasicAuth{
			Username: cfg.Auth.Plain.Username,
			Password: cfg.Auth.Plain.Password,
		}, nil
	case "ssh":
		log.Println("...using ssh key auth; loading public key...")
		pubkeys, err := ssh.NewPublicKeysFromFile(cfg.Auth.Ssh.Username, cfg.Auth.Ssh.PrivateKeyFilePath, cfg.Auth.Ssh.Passphrase)

		if err != nil {
			log.Println("...error!")
			return nil, fmt.Errorf("failed to prepare content git ssh authentication: %w", err)
		}

		log.Println("...ok!")
		return pubkeys, nil
	default:
		log.Println("...unknown auth method!")
		return nil, fmt.Errorf("invalid git authentication method %v", cfg.Auth.Method)
	}
}

func OpenOrClone(cfg *config.GitContentStrategy) (*git.Repository, error) {
	log.Println("inspecting content git repo directory...")
	state, repo, err := inspectRepo(cfg.LocalPath, cfg.Repository)
	if err != nil {
		log.Println("...error!")
		return nil, err
	}

	log.Println("ensuring clean path...")
	if err := ensureCleanPath(cfg.LocalPath, state); err != nil {
		log.Println("...error!")
		return nil, err
	}

	log.Println("checking if clone required...")

	if state == repoValid {
		log.Println("...using existing repo!")

		log.Println("pulling latest repo state...")
		w, err := repo.Worktree()
		if err != nil {
			log.Println("...failed to get worktree")
			return nil, err
		}

		err = w.Pull(&git.PullOptions{})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			log.Println("...failed to pull latest changes")
			return nil, err
		}

		log.Println("...ok!")
		return repo, nil
	}

	log.Println("...need to clone; building git authentication...")
	auth, err := BuildGitAuth(cfg)
	if err != nil {
		log.Println("...error!")
		return nil, err
	}

	log.Println("cloning repository...")
	return git.PlainClone(cfg.LocalPath, &git.CloneOptions{
		URL:  cfg.Repository,
		Auth: auth,
	})
}
