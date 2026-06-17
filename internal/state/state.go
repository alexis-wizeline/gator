package state

import (
	"github.com/alexis-wizeline/gator/internal/config"
	"github.com/alexis-wizeline/gator/internal/gatordb"
)

type State struct {
	DB     *gatordb.Queries
	Config *config.Config
}

func NewState(queries *gatordb.Queries, cfg *config.Config) *State {
	return &State{
		DB:     queries,
		Config: cfg,
	}
}
