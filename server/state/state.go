package state

import (
	"github.com/indieinfra/scribble/config"
	"github.com/indieinfra/scribble/storage/content"
	"github.com/indieinfra/scribble/storage/media"
)

type ScribbleState struct {
	Cfg          *config.Config
	ContentStore content.ContentStore
	MediaStore   media.MediaStore
}
