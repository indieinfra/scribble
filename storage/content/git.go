package content

import (
	"context"

	"github.com/indieinfra/scribble/server/util"
)

type GitContentStore struct{}

func (cs *GitContentStore) Create(ctx context.Context, doc util.Mf2Document) (string, bool, error) {
	return "https://noop.example.org/noop", false, nil
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

func (cs *GitContentStore) Get(ctx context.Context, url string) (*ContentObject, error) {
	return &ContentObject{
		Url:  url,
		Type: []string{"h-entry"},
		Properties: map[string][]any{
			"name":    {"This is a bogus title"},
			"content": {"This is bogus content, sentence one", "sentence two!"},
		},
		Deleted: false,
	}, nil
}
