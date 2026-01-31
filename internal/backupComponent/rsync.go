package backupComponent

import (
	"log/slog"
	"os/exec"
)

type RsyncBackup struct {
	BaseComponentAttributes `yaml:",inline"`
	Source                  string   `yaml:"source"`
	DstRelativePath         string   `yaml:"dstRelativePath"`
	UseDefaultArgs          bool     `yaml:"useDefaultArgs"`
	RsyncArgs               []string `yaml:"rsyncArgs"` // Further rsync options, e.g. ["-a", "--delete"]
}

var rsyncDefaultArgs = []string{"-avx", "--mkpath", "--delete"}

func (rsyncBackup RsyncBackup) PerformBackup(rsyncTargetHost string) error {
	slog.Debug("---Performing Rsync backup---")

	destination := rsyncTargetHost + rsyncBackup.DstRelativePath
	command := "rsync"

	var args []string
	if rsyncBackup.UseDefaultArgs {
		if len(rsyncBackup.RsyncArgs) > 0 {
			args = append(rsyncDefaultArgs, rsyncBackup.RsyncArgs...)
			slog.Warn("Default Rsync Args AND rsyncArgs are both used", "defaultArgs", rsyncDefaultArgs, "specifiedArgs", rsyncBackup.RsyncArgs)
		} else {
			args = rsyncDefaultArgs
		}
	} else {
		args = rsyncBackup.RsyncArgs
	}
	
	completeArgs := append(args, rsyncBackup.Source, destination)

	_, err := exec.Command(command, completeArgs...).Output()
	if err != nil {
		return AppendExecErrStderr("Error executing rsync command", err, "")
	}

	slog.Debug("Rsync command executed successfully", "command", command, "args", completeArgs)
	return nil
}
