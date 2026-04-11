package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/lmittmann/tint"
)

var Output io.Writer = os.Stdout

// Strips ANSI escape codes from a string
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

const (
	Reset  = "\033[0m"
	Faint  = "\033[2m"
	Cyan   = "\033[36m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
)

func Init() {
	logger := slog.New(tint.NewHandler(Output, &tint.Options{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
}

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// SubprocessStdout prints a single stdout line from a subprocess.
// Indented and dimmed to be visually subordinate to slog lines.
//
//	[nextcloud] sending incremental file list
func SubprocessStdout(prefix, line string) {
	fmt.Fprintf(Output, "    %s[%s]%s %s%s%s\n",
		Cyan, prefix, Reset,
		Faint, stripANSI(line), Reset,
	)
}

// SubprocessStderr prints stderr output from a subprocess, one line at a time.
// Yellow to signal it's noteworthy but not an app-level error.
func SubprocessStderr(prefix, block string) {
	for _, line := range strings.Split(strings.TrimRight(block, "\n"), "\n") {
		if line == "" {
			continue
		}
		fmt.Fprintf(Output, "    %s[%s]%s %s%s%s\n",
			Cyan, prefix, Reset,
			Yellow+Faint, stripANSI(line), Reset,
		)
	}
}
