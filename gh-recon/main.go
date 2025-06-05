package ghrecon

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/google/go-github/v72/github"
)

type Recon struct {
	Client      *github.Client
	Logger      *log.Logger
	Ctx         context.Context
	Silent      bool
	JsonFile    string
	MaxRepoSize int
	MaxDistance int
	ShowSource  bool
}
