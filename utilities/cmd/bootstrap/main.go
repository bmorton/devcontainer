// Command bootstrap steps you through the per-container setup the image can't
// bake in: git identity, GitHub/Claude/Copilot/opencode/pi/Namespace/Azure
// logins and mise tool health. With --check (or when not attached to a TTY)
// it prints a report instead of opening the TUI.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/bmorton/devcontainer/utilities/internal/bootstrap"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	check := flag.Bool("check", false, "print a status report and exit 1 if required steps are incomplete")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: bootstrap [--check] [--version]\n\n")
		fmt.Fprintf(os.Stderr, "Interactive by default; use --check for a non-interactive report.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println("bootstrap", buildVersion())
		return
	}
	os.Exit(bootstrap.Run(bootstrap.Options{Check: *check}))
}

func buildVersion() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				return "dev-" + s.Value[:7]
			}
		}
	}
	return version
}
