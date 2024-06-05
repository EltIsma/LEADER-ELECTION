package initt

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/attempter"
	"github.com/go-zookeeper/zk"
)

func NewInitState(
	logger *slog.Logger,
	zkServers []string,
	leaderTimeout time.Duration,
	attempterTimeout time.Duration,
	fileDir string,
	storageCapacity int,
) *InitState {
	logger = logger.With("subsystem", "InitState")
	return &InitState{
		logger:           logger,
		zkServers:        zkServers,
		leaderTimeout:    leaderTimeout,
		attempterTimeout: attempterTimeout,
		fileDir:          fileDir,
		storageCapacity:  storageCapacity,
	}
}

type InitState struct {
	logger           *slog.Logger
	zkServers        []string
	leaderTimeout    time.Duration
	attempterTimeout time.Duration
	fileDir          string
	storageCapacity  int
}

func (s *InitState) String() string {
	return "InitState"
}

func (s *InitState) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Initializing resources")

	conn, _, err := zk.Connect(s.zkServers, 5*time.Second)
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "error occured", slog.String("msg", err.Error()))
		return nil, err
	}
	fmt.Println(conn.State().String())
	if conn.State() != zk.StateHasSession {
		return attempter.NewFailoverState(s.logger, s.zkServers, conn, s.leaderTimeout, s.attempterTimeout, s.fileDir, s.storageCapacity), nil
	}

	return attempter.NewAttempterState(s.logger, conn, s.zkServers, s.leaderTimeout, s.attempterTimeout, s.fileDir, s.storageCapacity), nil
}
