package commands

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph"
	"github.com/spf13/cobra"
)

var (
	defaultZKServers        = []string{"zoo1:2181", "zoo2:2181", "zoo3:2181"}
	defaultLeaderTimeout    = 10 * time.Second
	defaultAttempterTimeout = 10 * time.Second
	defaultFileDir          = "tmp"
	defaultStorageCapacity  = 10
)

func InitRunCommand() (cobra.Command, error) {
	cmdArgs := cmdargs.RunArgs{}
	cmd := cobra.Command{
		Use:   "run",
		Short: "Starts a leader election node",
		Long: `This command starts the leader election node that connects to zookeeper
		and starts to try to acquire leadership by creation of ephemeral node`,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				fmt.Println(arg)
			}
			dg := depgraph.New()
			// zkConn, err := dg.GetZkConn()
			logger, err := dg.GetLogger()
			if err != nil {
				return fmt.Errorf("get logger: %w", err)
			}
			logger.Info("args received", slog.String("servers", strings.Join(cmdArgs.ZookeeperServers, ", ")))

			runner, err := dg.GetRunner()
			if err != nil {
				return fmt.Errorf("get runner: %w", err)
			}
			firstState, err := dg.GetInitState(cmdArgs.ZookeeperServers, cmdArgs.LeaderTimeout, cmdArgs.AttempterTimeout, cmdArgs.FileDir, cmdArgs.StorageCapacity)
			if err != nil {
				return fmt.Errorf("get first state: %w", err)
			}
			err = runner.Run(cmd.Context(), firstState)
			if err != nil {
				return fmt.Errorf("run states: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&(cmdArgs.ZookeeperServers), "zk-servers", "s", defaultZKServers, "Set the zookeeper servers.")
	cmd.Flags().DurationVarP(&(cmdArgs.LeaderTimeout), "leader-timeout", "l", defaultLeaderTimeout, "Leader timeout duration.")
	cmd.Flags().DurationVarP(&(cmdArgs.AttempterTimeout), "attempter-timeout", "a", defaultAttempterTimeout, "Attempter timeout duration.")
	cmd.Flags().StringVarP(&(cmdArgs.FileDir), "file-dir", "f", defaultFileDir, "File directory for leader to write files.")
	cmd.Flags().IntVarP(&(cmdArgs.StorageCapacity), "storage-capacity", "c", defaultStorageCapacity, "Maximum number of files to keep in file directory.")

	return cmd, nil
}
