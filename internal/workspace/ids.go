package workspace

import (
	"fmt"
	"time"
)

// NewID names a project, a folder or an environment.
func NewID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
