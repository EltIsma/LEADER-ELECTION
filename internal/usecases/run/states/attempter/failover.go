package attempter

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
)

type FailoverState struct {
	logger           *slog.Logger
	zkServers        []string
	zkConn           *zk.Conn
	leaderTimeout    time.Duration
	attempterTimeout time.Duration
	fileDir          string
	storageCapacity  int
}

func NewFailoverState(
	logger *slog.Logger,
	zkServers []string,
	zkConn *zk.Conn,
	leaderTimeout time.Duration,
	attempterTimeout time.Duration,
	fileDir string,
	storageCapacity int,
) *FailoverState {
	logger = logger.With("subsystem", "FailoverState")
	return &FailoverState{
		logger:           logger,
		zkServers:        zkServers,
		zkConn:           zkConn,
		leaderTimeout:    leaderTimeout,
		attempterTimeout: attempterTimeout,
		fileDir:          fileDir,
		storageCapacity:  storageCapacity,
	}
}

func (s *FailoverState) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Attempts to reconnect")
	maxRetries := 5
	retryInterval := 0 * time.Second
	for i := 0; i < maxRetries; i++ {
		//	time.Sleep(retryInterval)
		retryInterval *= 2
		select {
		case <-ctx.Done():
			return NewStoppingState(s.logger, s.zkConn), nil
		case <-time.After(retryInterval):
			zkConn, _, err := zk.Connect(s.zkServers, 3*time.Second)
			time.Sleep(3 * time.Second)
			if err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "error occured", slog.String("msg", err.Error()))
				return nil, err
			}
			if zkConn.State() != zk.StateHasSession {
				continue
			}
			return NewAttempterState(s.logger, zkConn, s.zkServers, s.leaderTimeout, s.attempterTimeout, s.fileDir, s.storageCapacity), nil
		}

	}
	return &FailoverState{}, errors.New("max failover retries exceeded")
}

func (s *FailoverState) String() string {
	return "Failover"
}
