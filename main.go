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

var (
	ExitWith string
	Timeout  time.Duration
)

func main() {
	cfg := linuxServiceConfig()
	if runtime.GOOS == "windows" {
		cfg = windowsServiceConfig()
	}

	d := daemon.NewDaemon(&daemon.DaemonConfig{
		Args:        []string{"run"},
		ExitTimeout: 5 * time.Second,
	})

	rootCmd := cmd.NewRootCmd()
	serviceCmd := cmd.NewServiceCmd(d, cfg)
	daemonCmd := cmd.NewDaemonCmd(d, cfg)

	runCmd := cmd.NewRunCmd(run)
	runCmd.Flags().StringVarP(&ExitWith, "exit-with", "e", "rand", "Exit the program with the specified status: nil, rand, err, panic, fatal")
	runCmd.Flags().DurationVarP(&Timeout, "timeout", "t", 30*time.Second, "Time to run before exiting")

	rootCmd.AddCommand(runCmd, serviceCmd, daemonCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func linuxServiceConfig() *kardianos.Config {
	return &kardianos.Config{
		Name:             "svcapp",
		DisplayName:      "SvcApp",
		Description:      "A simple example of a Go application that can be installed as a service",
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
		Name:             "svcapp",
		DisplayName:      "SvcApp",
		Description:      "A simple example of a Go application that can be installed as a service",
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

	if ExitWith == "rand" {
		ExitWith = []string{"", "err", "panic", "fatal"}[rand.IntN(4)]
	}

	if ExitWith != "" && ExitWith != "nil" {
		slog.Info(fmt.Sprintf("Process will exit with %s", ExitWith))
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(Timeout)
	timeoutChan := time.After(Timeout)

	for {
		select {
		case <-ticker.C:
			slog.Info(fmt.Sprintf("Running... %v left", time.Until(deadline).Truncate(time.Millisecond)))
		case <-timeoutChan:
			slog.Info("Timed out.")
			goto ret
		case <-ctx.Done():
			slog.Info("Context canceled.")
			goto ret
		}
	}

ret:
	slog.Info("Exiting...")
	switch ExitWith {
	case "err", "error":
		return fmt.Errorf("some error occurred")
	case "panic":
		log.Panic("Panic occurred")
	case "fatal":
		log.Fatal("Fatal occurred")
	}

	return nil
}
