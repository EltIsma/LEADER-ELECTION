package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands"
)

func main() {
	rootCmd, err := commands.InitRunCommand()
	if err != nil {
		fmt.Println("init run command: %w", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		sig := <-c
		fmt.Printf("Received signal: %v\n", sig)
		cancel()
	}()
	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		fmt.Println("run command: %w", err)
		os.Exit(1)
	}
}
