package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/security-scan/pkg/bench"
	"github.com/rancher/security-scan/pkg/helper"
	"github.com/rancher/security-scan/pkg/kb-summarizer/summarizer"
	"github.com/urfave/cli/v3"
)

const (
	K8SVersionFlag          = "k8s-version"
	K8SVersionEnvVar        = "RANCHER_K8S_VERSION"
	BenchmarkVersionFlag    = "benchmark-version"
	BenchMarkOverrideEnvVar = "OVERRIDE_BENCHMARK_VERSION"

	ControlsDirFlag   = "controls-dir"
	ControlsDirEnvVar = "CONFIG_DIR"

	OutputDirFlag         = "output-dir"
	OutputDirEnvVar       = "OUTPUT_DIR"
	outputFilenameDefault = "output.js"
	OutputFileNameFlag    = "output-filename"
)

var (
	VERSION = "v0.0.0-dev"

	benchmarkVersion string
	k8sVersion       string
	controlsDir      string
	outputDir        string
	outputFilename   string

	debugMode string
)

func main() {
	app := &cli.Command{
		Name:    "security-scan",
		Version: VERSION,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        BenchmarkVersionFlag,
				Destination: &benchmarkVersion,
				Sources:     cli.EnvVars(BenchMarkOverrideEnvVar),
			},
			&cli.StringFlag{
				Name:        K8SVersionFlag,
				Destination: &k8sVersion,
				Sources:     cli.EnvVars(K8SVersionEnvVar),
			},
			&cli.StringFlag{
				Name:        ControlsDirFlag,
				Destination: &controlsDir,
				Value:       summarizer.DefaultControlsDirectory,
				Sources:     cli.EnvVars(ControlsDirEnvVar),
			},

			&cli.StringFlag{
				Name:        OutputFileNameFlag,
				Destination: &outputFilename,
				Value:       outputFilenameDefault,
			},
			&cli.StringFlag{
				Name:        OutputDirFlag,
				Destination: &outputDir,
				Sources:     cli.EnvVars(OutputDirEnvVar),
			},

			&cli.StringFlag{
				Name:        "debug",
				Destination: &debugMode,
			},
		},
		Action: run,
	}

	err := app.Run(context.TODO(), os.Args)
	if err != nil {
		slog.Error("fatal error running application",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
}

func run(ctx context.Context, c *cli.Command) error {
	benchLogLevel := 0
	if strings.EqualFold(debugMode, "all") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		benchLogLevel = 7
	}

	var s *summarizer.Summarizer
	var err error

	if benchmarkVersion != "" {
		slog.Info("Running in benchmark specific version", "benchmark", benchmarkVersion)
		k8sVersion = ""
	} else {
		benchmarkVersion = ""
		controlsDir = ""

		if k8sVersion != "" {
			slog.Info("Provided Rancher Kubernetes Version", "version", k8sVersion)
		} else {
			v, err := helper.ServerVersion()
			if err != nil {
				return err
			}
			k8sVersion = v

			slog.Info("Calculated Rancher Kubernetes Version", "version", k8sVersion)
		}
	}

	err = bench.Run(context.TODO(), "master", benchmarkVersion, k8sVersion,
		"", controlsDir, filepath.Join(outputDir, "master.json"), benchLogLevel)
	if err != nil {
		return fmt.Errorf("master run: %w", err)
	}

	s, err = summarizer.NewSummarizer(
		k8sVersion,
		benchmarkVersion,
		controlsDir,
		outputDir,
		outputDir,
		outputFilename,
		"",
		"",
		"",
		false,
	)

	if err != nil {
		return fmt.Errorf("error creating summarizer: %w", err)
	}

	if err := s.Summarize(); err != nil {
		return fmt.Errorf("error summarizing: %w", err)
	}
	return nil
}
