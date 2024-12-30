//go:build !darwin

package diskutils

import (
	"github.com/ledgerwatch/log/v3"
)

func MountPointForDirPath(dirPath string) string {
	log.Info("[diskutils] Implemented only for darwin")
	return "/"
}
