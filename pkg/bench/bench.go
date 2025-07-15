package bench

import (
	"context"
	"log/slog"
	"os/exec"
	"strconv"
)

var kubeBenchBinary = "kube-bench"

func appendIfNotEmpty(args []string, flag, value string) []string {
	if value == "" {
		return args
	}

	return append(args, flag, value)
}

func Run(ctx context.Context, targets, benchmark, k8sVersion, logDir, configDir, outputFile string, logLevel int) error {

	args := []string{
		"run",
		"--scored",
		"--nosummary",
		"--noremediations",
		"--json",
	}

	args = append(args, "--targets", targets)

	if benchmark != "" {
		args = append(args, "--benchmark", benchmark)
	} else {
		args = append(args, "--version", k8sVersion)
	}

	args = appendIfNotEmpty(args, "--config-dir", configDir)
	args = appendIfNotEmpty(args, "--log_dir", logDir)
	args = appendIfNotEmpty(args, "--outputfile", outputFile)
	args = appendIfNotEmpty(args, "--v", strconv.Itoa(logLevel))

	cmd := exec.CommandContext(ctx, kubeBenchBinary, args...)
	output, err := cmd.CombinedOutput()

	slog.Debug("run kube-bench", "args", cmd.Args, "error", err, "output", output)

	return err
}
