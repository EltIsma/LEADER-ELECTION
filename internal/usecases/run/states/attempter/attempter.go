package attempter

import (
	"context"
	"log/slog"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
)

type AttempterState struct {
	logger           *slog.Logger
	zkConn           *zk.Conn
	zkServers        []string
	leaderTimeout    time.Duration
	attempterTimeout time.Duration
	fileDir          string
	storageCapacity  int
}

func NewAttempterState(
	logger *slog.Logger,
	zkConn *zk.Conn,
	zkServers []string,
	leaderTimeout time.Duration,
	attempterTimeout time.Duration,
	fileDir string,
	storageCapacity int,
) *AttempterState {
	logger = logger.With("subsystem", "AttempterState")
	return &AttempterState{
		logger:           logger,
		zkConn:           zkConn,
		zkServers:        zkServers,
		leaderTimeout:    leaderTimeout,
		attempterTimeout: attempterTimeout,
		fileDir:          fileDir,
		storageCapacity:  storageCapacity,
	}
}

func (a *AttempterState) String() string {
	return "AttempterState"
}

func (a *AttempterState) Run(ctx context.Context) (states.AutomataState, error) {
	for {
		select {
		case <-ctx.Done():
			return NewStoppingState(a.logger, a.zkConn), nil
		case <-time.After(a.attempterTimeout):

			if a.zkConn.State() != zk.StateHasSession {
				return NewFailoverState(a.logger, a.zkServers, a.zkConn, a.leaderTimeout, a.attempterTimeout, a.fileDir, a.storageCapacity), nil
			}
			exists, _, err := a.zkConn.Exists("/leader-election")
			if err != nil {
				a.logger.LogAttrs(ctx, slog.LevelError, "failed to check leader node existence", slog.String("msg", err.Error()))
				continue
			}

			if exists {
				a.logger.Info("leader node already exists, another replica is the leader")
				continue
			}

			// Try to create an ephemeral node in ZooKeeper
			_, err = a.zkConn.Create("/leader-election", []byte("I am the leader"), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
			if err != nil {
				a.logger.LogAttrs(ctx, slog.LevelError, "error occurred", slog.String("msg", err.Error()))
				continue
			}
			//	time.Sleep(60*time.Second)
			leaderState := NewLeaderState(a.logger, a.zkConn, a.leaderTimeout, a.zkServers, a.attempterTimeout, a.fileDir, a.storageCapacity)
			return leaderState, nil
		}
	}
}
