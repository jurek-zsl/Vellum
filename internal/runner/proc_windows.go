//go:build windows

package runner

import (
	"os/exec"
)

func setSysProcAttr(cmd *exec.Cmd) {
	// Windows does not support Unix pgid attributes
}
