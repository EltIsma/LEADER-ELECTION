package depgraph

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run"
	initt "github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/init"
	"github.com/go-zookeeper/zk"
)

type dgEntity[T any] struct {
	sync.Once
	value   T
	initErr error
}

func (e *dgEntity[T]) get(init func() (T, error)) (T, error) {
	e.Do(func() {
		e.value, e.initErr = init()
	})
	if e.initErr != nil {
		return *new(T), e.initErr
	}
	return e.value, nil
}

type DepGraph struct {
	logger      *dgEntity[*slog.Logger]
	stateRunner *dgEntity[*run.LoopRunner]
	initState   *dgEntity[*initt.InitState]
	zkConn      *dgEntity[*zk.Conn]
	config      *dgEntity[*Config]
}

func New() *DepGraph {
	return &DepGraph{
		logger:      &dgEntity[*slog.Logger]{},
		stateRunner: &dgEntity[*run.LoopRunner]{},
		initState:   &dgEntity[*initt.InitState]{},
		zkConn:      &dgEntity[*zk.Conn]{},
		config:      &dgEntity[*Config]{},
	}
}

func (dg *DepGraph) GetLogger() (*slog.Logger, error) {
	return dg.logger.get(func() (*slog.Logger, error) {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})), nil
	})
}

func (dg *DepGraph) GetInitState(
	zookeeperServers []string,
	leaderTimeout time.Duration,
	attempterTimeout time.Duration,
	fileDir string,
	storageCapacity int,
) (*initt.InitState, error) {
	return dg.initState.get(func() (*initt.InitState, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}
		return initt.NewInitState(logger, zookeeperServers, leaderTimeout, attempterTimeout, fileDir, storageCapacity), nil
	})
}

func (dg *DepGraph) GetRunner() (run.Runner, error) {
	return dg.stateRunner.get(func() (*run.LoopRunner, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}
		return run.NewLoopRunner(logger), nil
	})
}

func (dg *DepGraph) GetZkConn() (*zk.Conn, error) {
	return dg.zkConn.get(func() (*zk.Conn, error) {
		conn, _, err := zk.Connect([]string{"localhost:2181"}, 10*time.Second)
		if err != nil {
			return nil, fmt.Errorf("connect to zookeeper: %w", err)
		}
		return conn, nil
	})
}

type Config struct {
	ZKServers        []string
	LeaderTimeout    time.Duration
	AttempterTimeout time.Duration
	FileDir          string
	StorageCapacity  int
}

func (dg *DepGraph) GetConfig() (*Config, error) {
	return dg.config.get(func() (*Config, error) {
		return &Config{
			ZKServers:        []string{"localhost:2181"},
			LeaderTimeout:    10 * time.Second,
			AttempterTimeout: 10 * time.Second,
			FileDir:          "/tmp",
			StorageCapacity:  10,
		}, nil
	})
}
