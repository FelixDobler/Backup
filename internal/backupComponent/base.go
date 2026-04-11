package backupComponent

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"

	"fedob/backup/internal/logging"
)

type BackupComponent interface {
	PerformBackup(rsyncTargetHost string) error
	GetName() string
	IsEnabled() bool
	GetType() string
	SetDefaults()
	// RunCleanupOnFailure(err error) error
}

type BaseComponentAttributes struct {
	Name    string       `yaml:"name" validate:"required"`
	Type    string       `yaml:"type" validate:"required"`
	Enabled bool         `yaml:"enabled"`
	Output  OutputConfig `yaml:"output"`
}

type OutputConfig struct {
	ShowOutput bool   `yaml:"showOutput"`
	ShowStderr bool   `yaml:"showStderr"`
	Prefix     string `yaml:"prefix"`
}

// DefaultOutputConfig returns an OutputConfig with sensible defaults.
// Call this when building a BaseComponentAttributes before YAML unmarshalling.
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		ShowOutput: true,
		ShowStderr: true,
	}
}

func (bca BaseComponentAttributes) GetName() string {
	return bca.Name
}

func (bca BaseComponentAttributes) IsEnabled() bool {
	return bca.Enabled
}

func (bca BaseComponentAttributes) GetType() string {
	return bca.Type
}

func (bca *BaseComponentAttributes) SetDefaults() {
	bca.Output = DefaultOutputConfig()
}

func (bca BaseComponentAttributes) printSubprocessLine(line string) {
	logging.SubprocessStdout(bca.outputPrefix(), line)
}
func (bca BaseComponentAttributes) printSubprocessStderr(block string) {
	logging.SubprocessStderr(bca.outputPrefix(), block)
}

func (bca BaseComponentAttributes) outputPrefix() string {
	if bca.Output.Prefix != "" {
		return bca.Output.Prefix
	}
	return bca.Name
}

type BackupType string

const (
	OSBackupType        BackupType = "os"
	NextcloudBackupType BackupType = "nextcloud"
	RsyncBackupType     BackupType = "rsync"
)

func dirExists(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("Directory path is empty")
	}

	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("Failed to check directory: %w", err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", dirPath)
	}
	return nil
}

// Appends `err` to `message`. If `err` is of type exec.ExitError and contains stderr output,
// the stderr output is also appended to the returned error.
// If stderr is specified, it is used instead of extracting it from err.
func AppendExecErrStderr(message string, err error, stderr string) error {
	fallbackErr := fmt.Errorf("%s: %w", message, err)

	// check if error is of type exitError so we can access the stderr
	exitErr := &exec.ExitError{}
	if !errors.As(err, &exitErr) {
		return fallbackErr
	}

	var outputStderr string
	if stderr != "" {
		outputStderr = stderr
	} else {
		if exitErr.Stderr != nil {
			outputStderr = string(exitErr.Stderr)
			if len(outputStderr) > 100 {
				outputStderr = outputStderr[:100] + "[...]"
			}
		} else {
			return fallbackErr
		}
	}

	return fmt.Errorf("%s, %w, stderr:\n%s", message, err, outputStderr)
}

// ExecuteCommand runs cmd, streaming stdout live if ShowOutput is true.
// Stderr is always buffered and flushed after completion.
// Returns captured stdout bytes (useful for commands whose output is parsed by callers).
func (bca BaseComponentAttributes) ExecuteCommand(cmd *exec.Cmd) ([]byte, error) {
	slog.Info("Output Configuration", "component", bca, "output", bca.Output)
	if !bca.Output.ShowOutput {
		// Silent: buffer everything, stderr ends up in the error via AppendExecErrStderr
		out, err := cmd.Output()
		if err != nil {
			return nil, AppendExecErrStderr("command failed", err, "")
		}
		return out, nil
	}

	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
		mu        sync.Mutex // guards writes to Output from stdout goroutine vs deferred stderr
	)

	// Stream stdout line-by-line with prefix, also capture
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	// Stderr: buffer only
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start command: %w", err)
	}

	// Stream stdout in a goroutine so cmd.Wait() doesn't deadlock
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutReader)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutBuf.WriteString(line + "\n")
			mu.Lock()
			bca.printSubprocessLine(line)
			mu.Unlock()
		}
	}()

	wg.Wait()
	err = cmd.Wait()

	// Flush stderr after stdout is done — keeps output ordered and readable
	if bca.Output.ShowStderr && stderrBuf.Len() > 0 {
		mu.Lock()
		bca.printSubprocessStderr(stderrBuf.String())
		mu.Unlock()
	}

	if err != nil {
		return stderrBuf.Bytes(), AppendExecErrStderr("command failed", err, "")
	}

	return stdoutBuf.Bytes(), nil
}
