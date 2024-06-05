package cmdargs

import "time"

type RunArgs struct {
	ZookeeperServers []string
	LeaderTimeout    time.Duration
	AttempterTimeout time.Duration
	FileDir          string
	StorageCapacity  int
}
