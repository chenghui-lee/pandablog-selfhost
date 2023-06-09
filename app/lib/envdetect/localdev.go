package envdetect

import (
	"os"
)

// RunningLocalDev returns true if the PBB_LOCAL environment variable is set.
func RunningLocalDev() bool {
	s := os.Getenv("PBB_LOCAL")
	return len(s) > 0
}
