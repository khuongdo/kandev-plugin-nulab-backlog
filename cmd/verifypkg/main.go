// Command verifypkg verifies a built plugin package (BR5.2). All logic lives
// in internal/pkgverify.
package main

import (
	"os"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/pkgverify"
)

func main() {
	os.Exit(pkgverify.Run(os.Args[1:], os.Stdout, os.Stderr))
}
