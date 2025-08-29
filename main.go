package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"runtime"
	"time"

	"github.com/lucasdecamargo/go-appservice-example/cmd"
	"github.com/lucasdecamargo/go-appservice-example/pkg/daemon"
	"github.com/lucasdecamargo/kardianos"
)

const (
	// Service configuration constants
	serviceName        = "svcapp"
	serviceDisplayName = "SvcApp"
	serviceDescription = "A simple example of a Go application that can be installed as a service"

	// Default timeouts
	defaultExitTimeout = 5 * time.Second
	defaultRunTimeout  = 30 * time.Second

	// Exit modes
	exitModeNil   = "nil"
	exitModeRand  = "rand"
	exitModeErr   = "err"
	exitModePanic = "panic"
	exitModeFatal = "fatal"
)

var (
	ExitWith string
	Timeout  time.Duration
)

func main() {
	cfg := getServiceConfig()

	d := daemon.NewDaemon(&daemon.DaemonConfig{
		Args:        []string{"run"},
		ExitTimeout: defaultExitTimeout,
	})

	rootCmd := cmd.NewRootCmd()
	serviceCmd := cmd.NewServiceCmd(d, cfg)
	daemonCmd := cmd.NewDaemonCmd(d, cfg)

	runCmd := cmd.NewRunCmd(run)
	runCmd.Flags().StringVarP(&ExitWith, "exit-with", "e", exitModeRand,
		fmt.Sprintf("Exit the program with the specified status: %s, %s, %s, %s, %s",
			exitModeNil, exitModeRand, exitModeErr, exitModePanic, exitModeFatal))
	runCmd.Flags().DurationVarP(&Timeout, "timeout", "t", defaultRunTimeout, "Time to run before exiting")

	rootCmd.AddCommand(runCmd, serviceCmd, daemonCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Failed to execute command:", err)
	}
}

func getServiceConfig() *kardianos.Config {
	if runtime.GOOS == "windows" {
		return windowsServiceConfig()
	}
	return linuxServiceConfig()
}

func linuxServiceConfig() *kardianos.Config {
	return &kardianos.Config{
		Name:             serviceName,
		DisplayName:      serviceDisplayName,
		Description:      serviceDescription,
		WorkingDirectory: "~/.",
		Arguments:        []string{"daemon"},

		Option: kardianos.KeyValue{
			"LogOutput":         false,
			"PIDFile":           "/var/run/svcapp.pid",
			"Restart":           "on-success",
			"SuccessExitStatus": "0 2 SIGKILL",
			"LimitNOFILE":       -1,
		},

		Dependencies: []string{
			"After=network-online.target",
			"Wants=network-online.target",
		},
	}
}

func windowsServiceConfig() *kardianos.Config {
	return &kardianos.Config{
		Name:             serviceName,
		DisplayName:      serviceDisplayName,
		Description:      serviceDescription,
		WorkingDirectory: "~/.",
		Arguments:        []string{"daemon"},

		Option: kardianos.KeyValue{
			"StartType":              "automatic",
			"OnFailure":              "restart",
			"OnFailureDelayDuration": "10s",
		},
	}
}

func run(ctx context.Context, args []string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Determine exit mode
	exitMode := determineExitMode(ExitWith)
	if exitMode != exitModeNil {
		slog.Info("Process will exit with", "mode", exitMode)
	}

	// Run the main loop
	return runMainLoop(ctx, exitMode)
}

func determineExitMode(mode string) string {
	if mode == exitModeRand {
		modes := []string{exitModeNil, exitModeErr, exitModePanic, exitModeFatal}
		return modes[rand.IntN(len(modes))]
	}
	return mode
}

func runMainLoop(ctx context.Context, exitMode string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(Timeout)
	timeoutChan := time.After(Timeout)

	for {
		select {
		case <-ticker.C:
			remaining := time.Until(deadline).Truncate(time.Millisecond)
			slog.Info("Running...", "timeLeft", remaining)
		case <-timeoutChan:
			slog.Info("Timed out.")
			return exitWithMode(exitMode)
		case <-ctx.Done():
			slog.Info("Context canceled.")
			return exitWithMode(exitMode)
		}
	}
}

func exitWithMode(mode string) error {
	slog.Info("Exiting...", "mode", mode)

	switch mode {
	case exitModeErr:
		return fmt.Errorf("some error occurred")
	case exitModePanic:
		log.Panic("Panic occurred")
		// This line is unreachable, but needed for compilation
		return fmt.Errorf("panic occurred")
	case exitModeFatal:
		log.Fatal("Fatal occurred")
		// This line is unreachable, but needed for compilation
		return fmt.Errorf("fatal occurred")
	default:
		return nil
	}
}
