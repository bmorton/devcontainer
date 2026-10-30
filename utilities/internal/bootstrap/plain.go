package bootstrap

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/bmorton/devcontainer/utilities/internal/ui/theme"
)

// RunCheck runs every check concurrently and prints a non-interactive report.
// It returns 1 when any required step is not ready, otherwise 0.
func RunCheck(ctx context.Context, env Env, steps []*Step, w io.Writer) int {
	results := make([]Result, len(steps))
	var wg sync.WaitGroup
	for i, s := range steps {
		wg.Add(1)
		go func(i int, s *Step) {
			defer wg.Done()
			cctx, cancel := context.WithTimeout(ctx, checkTimeout)
			defer cancel()
			results[i] = s.Check(cctx, env)
		}(i, s)
	}
	wg.Wait()

	fmt.Fprintln(w, theme.Header("bootstrap", "status report"))
	fmt.Fprintln(w)
	exit := 0
	for i, s := range steps {
		r := results[i]
		kind := "unknown"
		switch r.Status {
		case StatusReady:
			kind = "ready"
		case StatusNeeded:
			kind = "needed"
			if s.Optional {
				kind = "optional"
			} else {
				exit = 1
			}
		case StatusError:
			kind = "error"
			if !s.Optional {
				exit = 1
			}
		}
		title := theme.Text.Render(fmt.Sprintf("%-22s", s.Title))
		tag := ""
		if s.Optional {
			tag = theme.Muted.Render(" optional")
		}
		fmt.Fprintf(w, "  %s %s %s%s\n", theme.StatusGlyph(kind), title, theme.Dim.Render(r.Summary), tag)
		if r.Detail != "" && r.Status != StatusReady {
			for _, line := range strings.Split(strings.TrimSpace(r.Detail), "\n") {
				fmt.Fprintf(w, "      %s\n", theme.Muted.Render(line))
			}
		}
	}
	fmt.Fprintln(w)
	if exit == 0 {
		fmt.Fprintln(w, "  "+theme.Gradient("ALL SYSTEMS ONLINE", theme.Acid, theme.Cyan))
	} else {
		fmt.Fprintln(w, "  "+theme.Danger.Render("required steps need attention")+theme.Dim.Render("  run `bootstrap` to step through them"))
	}
	return exit
}
