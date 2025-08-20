package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"run-pod-spec/pkg/kube"
	"run-pod-spec/pkg/logging"
)

func fail(message string, v ...any) {
	slog.Error(message, v...)
	os.Exit(1)
}

func main() {
	var filepath string
	flag.StringVar(&filepath, "f", "", "The path to a Pod manifest file (YAML)")

	var debug bool
	flag.BoolVar(&debug, "debug", false, "Whether to enable debug mode")

	var noRm bool
	flag.BoolVar(&noRm, "no-rm", false, "Whether to avoid deletion (removal) of the pod at the end of the run")

	var createTimeout int
	flag.IntVar(&createTimeout, "create-timeout", 60, "Amount of seconds we wait for create pod command to finish")

	var runTimeout int
	flag.IntVar(&runTimeout, "run-timeout", 600, "Amount of seconds we wait maximally for the pod to finish its work")

	var noReplace bool
	flag.BoolVar(&noReplace, "no-replace", false, "Whether to not replace old pods. If set to true execution will fail rather than cleaning up and spawning new pod.")

	flag.Parse()

	if debug {
		logging.InitializeLogging(slog.LevelDebug, os.Stderr)
	} else {
		logging.InitializeLogging(slog.LevelInfo, os.Stderr)
	}
	slog.Debug("CLI argument", "filepath", filepath)
	slog.Debug("CLI argument", "debug", debug)
	slog.Debug("CLI argument", "create-timeout", createTimeout)
	slog.Debug("CLI argument", "run-timeout", runTimeout)
	slog.Debug("CLI argument", "no-rm", noRm)
	slog.Debug("CLI argument", "no-replace", noReplace)

	cleanupPod := !noRm
	slog.Debug("Action variable", "cleanupPod", cleanupPod)

	if filepath == "" {
		fail("A path to a pod manifest must be provided. Use '-f' to specify a manifest file.")
	}

	mainCtx := context.Background()

	opts := kube.RunPodOptions{
		CreateTimeout: time.Duration(createTimeout) * time.Second,
		RunTimeout:    time.Duration(runTimeout) * time.Second,
		CleanupPod:    cleanupPod,
		ReplaceOldPod: !noReplace,
	}

	slog.Debug("Going to run pod", "options", opts)
	err := kube.RunPod(mainCtx, filepath, &opts)

	if err == nil {
		os.Exit(0)
	}

	fail("Encountered error", "error", err)

}
