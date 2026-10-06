// Command ci runs the plugin's CI and release checks. All logic lives in
// internal/ci.
package main

import (
	"os"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/ci"
)

func main() {
	os.Exit(ci.Run(os.Args[1:], os.Stdout, os.Stderr))
}
