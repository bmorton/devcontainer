package bootstrap

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// plain strips SGR color sequences so assertions can match visible text.
func plain(s string) string { return ansiRe.ReplaceAllString(s, "") }

// TestModelRendersWithoutTerminal drives the model headlessly through a
// resize, a batch of results and a cursor move, and checks the view holds
// together. It guards against panics in View and keeps the layout contract
// (title, progress label, footer) honest.
func TestModelRendersWithoutTerminal(t *testing.T) {
	steps := Steps()
	var m tea.Model = NewModel(testEnv(nil, nil, nil), steps)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	for i := range steps {
		res := Ready("ok")
		if steps[i].Optional {
			res = Needed("not logged in", "")
		}
		m, _ = m.Update(checkDoneMsg{idx: i, res: res})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})

	v := plain(m.View().Content)
	for _, want := range []string{"B O O T S T R A P", "ALL SYSTEMS ONLINE", "optional", "re-check"} {
		if !strings.Contains(v, want) {
			t.Errorf("view missing %q", want)
		}
	}
	if strings.Contains(v, "0/") {
		t.Errorf("progress label did not update: %s", v)
	}
}

func TestModelQuits(t *testing.T) {
	var m tea.Model = NewModel(testEnv(nil, nil, nil), Steps())
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", cmd())
	}
}
