// Command plugin is the Kandev plugin process for Nulab Backlog.
package main

import (
	"github.com/kandev/kandev/pkg/pluginsdk"

	"github.com/khuongdo/kandev-plugin-nulab-backlog/internal/plugin"
)

func main() {
	pluginsdk.Serve(plugin.NewRuntime())
}
