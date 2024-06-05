package attempter

import (
	"context"
	"log/slog"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
)

type StoppingState struct {
	logger *slog.Logger
	zkConn *zk.Conn
}

func NewStoppingState(logger *slog.Logger, zkConn *zk.Conn) *StoppingState {
	logger = logger.With("subsystem", "Stopping")
	return &StoppingState{
		logger: logger,
		zkConn: zkConn,
	}
}

func (s *StoppingState) Run(ctx context.Context) (states.AutomataState, error) {
	<-ctx.Done()
	s.logger.Info("Gracefully shutting down the application")
	s.logger.Debug("Closing ZooKeeper connection")
	s.zkConn.Close()
	s.logger.Info("Application has been shut down")

	return &StoppingState{}, nil
}

func (s *StoppingState) String() string {
	return "StoppingState"
}
