package contentgit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/google/uuid"
	"github.com/indieinfra/scribble/config"
	"github.com/indieinfra/scribble/server/util"
	"github.com/indieinfra/scribble/storage/content"
)

type GitContentStore struct {
	cfg      *config.GitContentStrategy
	repo     *git.Repository
	worktree *git.Worktree
	auth     *transport.AuthMethod
}

func NewGitContentStore(cfg *config.GitContentStrategy, repo *git.Repository) (*GitContentStore, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	auth, err := BuildGitAuth(cfg)
	if err != nil {
		return nil, err
	}

	return &GitContentStore{
		cfg:      cfg,
		repo:     repo,
		worktree: worktree,
		auth:     &auth,
	}, nil
}

func (cs *GitContentStore) Create(ctx context.Context, doc util.Mf2Document) (string, bool, error) {
	head, err := cs.repo.Head()
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve content git HEAD: %w", err)
	}

	resetOpts := &git.ResetOptions{Commit: head.Hash(), Mode: git.HardReset}

	contentId, err := uuid.NewRandom()
	if err != nil {
		return "", false, err
	}

	jsonBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", false, err
	}

	filename := fmt.Sprintf("%v.json", contentId.String())
	filePath := filepath.Join(cs.cfg.LocalPath, cs.cfg.Path, filename)
	if filePath == "" {
		return "", false, fmt.Errorf("failed to create file path for new content (localPath: %v, subPath: %v, filename: %v)", cs.cfg.LocalPath, cs.cfg.Path, filename)
	}

	err = os.WriteFile(filePath, jsonBytes, 0644)
	if err != nil {
		msg := "failed to write file"

		if err := cs.worktree.Reset(resetOpts); err != nil {
			return "", false, fmt.Errorf("%v; also, failed to reset content git worktree: %w", msg, err)
		}

		return "", false, fmt.Errorf("%v: %w", msg, err)
	}

	_, err = cs.worktree.Add(".")
	if err != nil {
		msg := "failed to add new file to git"
		if err := cs.worktree.Reset(resetOpts); err != nil {
			return "", false, fmt.Errorf("%v; also, failed to restore worktree: %w", msg, err)
		}

		return "", false, fmt.Errorf("%v: %w", msg, err)
	}

	_, err = cs.worktree.Commit("scribble(add): create content entry", &git.CommitOptions{All: true})
	if err != nil {
		msg := "failed to create new commit"
		if err := cs.worktree.Reset(resetOpts); err != nil {
			return "", false, fmt.Errorf("%v; also, failed to restore worktree: %w", msg, err)
		}

		return "", false, fmt.Errorf("%v: %w", msg, err)
	}

	if err := cs.repo.Push(&git.PushOptions{Auth: *cs.auth}); err != nil {
		msg := "failed to push repo"
		if err := cs.worktree.Reset(resetOpts); err != nil {
			return "", false, fmt.Errorf("%v; also, failed to reset HEAD: %w", msg, err)
		}

		return "", false, fmt.Errorf("%v: %w", msg, err)
	}

	return fmt.Sprintf("%v/%v", cs.cfg.PublicUrl, filename), false, nil
}

func (cs *GitContentStore) Update(ctx context.Context, url string, replacements map[string][]any, additions map[string][]any, deletions any) (string, error) {
	return url, nil
}

func (cs *GitContentStore) Delete(ctx context.Context, url string) error {
	return nil
}

func (cs *GitContentStore) Undelete(ctx context.Context, url string) (string, bool, error) {
	return url, false, nil
}

func (cs *GitContentStore) Get(ctx context.Context, url string) (*content.ContentObject, error) {
	return &content.ContentObject{
		Url:  url,
		Type: []string{"h-entry"},
		Properties: map[string][]any{
			"name":    {"This is a bogus title"},
			"content": {"This is bogus content, sentence one", "sentence two!"},
		},
		Deleted: false,
	}, nil
}
