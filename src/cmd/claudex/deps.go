package main

import (
	"claudex/internal/services/clock"
	"claudex/internal/services/commander"
	"claudex/internal/services/env"
	"claudex/internal/services/uuid"

	"github.com/spf13/afero"
)

// Dependencies holds all external dependencies for the application
type Dependencies struct {
	FS    afero.Fs
	Cmd   commander.Commander
	Clock clock.Clock
	UUID  uuid.UUIDGenerator
	Env   env.Environment
}

// NewDependencies creates a new Dependencies instance with production defaults
func NewDependencies() *Dependencies {
	return &Dependencies{
		FS:    afero.NewOsFs(),
		Cmd:   commander.New(),
		Clock: clock.New(),
		UUID:  uuid.New(),
		Env:   env.New(),
	}
}
