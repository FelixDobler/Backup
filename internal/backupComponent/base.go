package backupComponent

import (
	"errors"
	"fmt"
	"os/exec"
)

type BackupComponent interface {
	PerformBackup(rsyncTargetHost string) error
	GetName() string
	IsEnabled() bool
	GetType() string
	// RunCleanupOnFailure(err error) error
}

type BaseComponentAttributes struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Enabled bool   `yaml:"enabled"`
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

type BackupType string

const (
	OSBackupType        BackupType = "os"
	NextcloudBackupType BackupType = "nextcloud"
	RsyncBackupType     BackupType = "rsync"
)

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
		} else {
			return fallbackErr
		}
	}

	return fmt.Errorf("%s, %w, stderr:\n%s", message, err, outputStderr)
}
