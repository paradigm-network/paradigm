// +build !linux

package file

import (
	"os"
)

func Chown(_ string, _ os.FileInfo) error {
	return nil
}
