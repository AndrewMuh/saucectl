package docker

import (
	"context"

	"github.com/saucelabs/saucectl/internal/framework"
	"github.com/saucelabs/saucectl/internal/job"
	"github.com/saucelabs/saucectl/internal/report"
	"github.com/saucelabs/saucectl/internal/testcafe"
)

// TestcafeRunner represents the docker implementation of a test runner.
type TestcafeRunner struct {
	ContainerRunner
	Project testcafe.Project
}

// NewTestcafe creates a new TestcafeRunner instance.
func NewTestcafe(c testcafe.Project, ms framework.MetadataService, wr job.Writer, jr job.Reader, dl job.ArtifactDownloader, reps []report.Reporter) (*TestcafeRunner, error) {
	r := TestcafeRunner{
		Project: c,
		ContainerRunner: ContainerRunner{
			Ctx:             context.Background(),
			containerConfig: &containerConfig{},
			Framework: framework.Framework{
				Name:    c.Kind,
				Version: c.Testcafe.Version,
			},
			FrameworkMeta:          ms,
			ShowConsoleLog:         c.ShowConsoleLog,
			JobWriter:              wr,
			JobReader:              jr,
			ArtfactDownloader:      dl,
			Reporters:              reps,
			MetadataSearchStrategy: framework.NewSearchStrategy(c.Testcafe.Version, c.RootDir),
		},
	}

	var err error
	r.docker, err = Create()
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// RunProject runs the tests defined in config.Project.
func (r *TestcafeRunner) RunProject() (int, error) {
	verifyFileTransferCompatibility(r.Project.Sauce.Concurrency, &r.Project.Docker)

	if err := r.fetchImage(&r.Project.Docker); err != nil {
		return 1, err
	}

	sigChan := r.registerSkipSuitesOnSignal()
	defer unregisterSignalCapture(sigChan)

	containerOpts, results := r.createWorkerPool(r.Project.Sauce.Concurrency)
	defer close(results)

	go func() {
		for _, suite := range r.Project.Suites {
			containerOpts <- containerStartOptions{
				Docker:         r.Project.Docker,
				BeforeExec:     r.Project.BeforeExec,
				Project:        r.Project,
				Browser:        suite.BrowserName,
				DisplayName:    suite.Name,
				SuiteName:      suite.Name,
				Environment:    suite.Env,
				RootDir:        r.Project.RootDir,
				Sauceignore:    r.Project.Sauce.Sauceignore,
				ConfigFilePath: r.Project.ConfigFilePath,
				CLIFlags:       r.Project.CLIFlags,
				Timeout:        suite.Timeout,
			}
		}
		close(containerOpts)
	}()

	hasPassed := r.collectResults(r.Project.Artifacts.Download, results, len(r.Project.Suites))
	if !hasPassed {
		return 1, nil
	}

	return 0, nil
}
