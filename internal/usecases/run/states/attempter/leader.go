package attempter

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/go-zookeeper/zk"
)

type LeaderState struct {
	logger           *slog.Logger
	zkConn           *zk.Conn
	leaderTimeout    time.Duration
	zkServers        []string
	fileDir          string
	storageCapacity  int
	attempterTimeout time.Duration
}

func NewLeaderState(
	logger *slog.Logger,
	zkConn *zk.Conn,
	leaderTimeout time.Duration,
	zkServers []string,
	attempterTimeout time.Duration,
	fileDir string,
	storageCapacity int,
) *LeaderState {
	logger = logger.With("subsystem", "LeaderState")
	return &LeaderState{
		logger:           logger,
		zkConn:           zkConn,
		leaderTimeout:    leaderTimeout,
		zkServers:        zkServers,
		attempterTimeout: attempterTimeout,
		fileDir:          fileDir,
		storageCapacity:  storageCapacity,
	}
}

func (s *LeaderState) Run(ctx context.Context) (states.AutomataState, error) {
	// Write a file to disk every leader timeout
	ticker := time.NewTicker(s.leaderTimeout)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			return NewStoppingState(s.logger, s.zkConn), nil
		case <-ticker.C:
			// Generate a file name with the current time and the leader's ID
			fileName := fmt.Sprintf("%d-%s", time.Now().Unix(), s.zkConn.Server())
			filePath := filepath.Join(s.fileDir, fileName)

			if s.zkConn.State() != zk.StateHasSession {
				return NewFailoverState(s.logger, s.zkServers, s.zkConn, s.leaderTimeout, s.attempterTimeout, s.fileDir, s.storageCapacity), nil
			}

			err := os.WriteFile(filePath, []byte("Leader"), 0o644)
			if err != nil {
				return s, fmt.Errorf("write file: %w", err)
			}

			files, err := os.ReadDir("./" + s.fileDir)
			if err != nil {
				return s, fmt.Errorf("read directory: %w", err)
			}

			// Delete old files if there are too many
			if len(files) > s.storageCapacity {
				for i := 0; i < len(files)-s.storageCapacity; i++ {
					os.Remove(filepath.Join(s.fileDir, files[i].Name()))
				}
			}
		}
	}
}

func (s *LeaderState) String() string {
	return "LeaderState"
}
