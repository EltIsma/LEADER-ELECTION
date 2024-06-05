package empty

import (
	"context"
	"log/slog"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

func New(logger *slog.Logger) *State {
	logger = logger.With("subsystem", "EmptyState")
	return &State{
		logger: logger,
	}
}

type State struct {
	logger *slog.Logger
}

func (s *State) String() string {
	return "EmptyState"
}

func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Nothing happened")
	return nil, nil
}
