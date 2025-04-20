//nolint:mnd
package table

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InteractiveTable wraps bubble table and viewport.
type InteractiveTable struct {
	table       Model
	viewport    viewport.Model
	isAltScreen bool
	footerTexts []string
	helpTexts   []string
	enterFunc   func(string) error
	copyFunc    func(string, string) error
	showHelp    bool
}

// InteractiveTableOption is a functional opt for InteractiveTable.
type InteractiveTableOption func(*InteractiveTable)

// NewInteractiveTable builds a new table.
func NewInteractiveTable(cols []Column, rows []Row, opts ...InteractiveTableOption) *InteractiveTable {
	height := min(len(rows)+1, 25)
	t := New(
		WithColumns(cols),
		WithRows(rows),
		WithFocused(true),
		WithHeight(height),
	)

	vp := viewport.New(min(200, len(cols)*20), height+1)
	vp.SetContent(t.View())

	it := InteractiveTable{
		table:    t,
		viewport: vp,
	}
	for _, o := range opts {
		o(&it)
	}
	return &it
}

// WithHelpTexts sets help text.
func WithHelpTexts(txt []string) InteractiveTableOption {
	return func(t *InteractiveTable) {
		t.helpTexts = txt
	}
}

// WithFooterTexts sets footer text.
func WithFooterTexts(txt []string) InteractiveTableOption {
	return func(t *InteractiveTable) {
		t.footerTexts = txt
	}
}

// WithEnterFunc registers a method to call when user presses an 'enter' key.
func WithEnterFunc(fn func(id string) error) InteractiveTableOption {
	return func(t *InteractiveTable) {
		t.enterFunc = fn
	}
}

// WithCopyFunc registers a method to call when user presses 'c'.
func WithCopyFunc(fn func(id string, key string) error) InteractiveTableOption {
	return func(t *InteractiveTable) {
		t.copyFunc = fn
	}
}

// Init is required for initialization.
func (t *InteractiveTable) Init() tea.Cmd { return nil }

// Update is the Bubble Tea update loop.
func (t *InteractiveTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) { //nolint:gocritic
	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			t.showHelp = !t.showHelp
		case "m":
			if !t.isAltScreen {
				cmds = append(cmds, tea.EnterAltScreen)
			} else {
				cmds = append(cmds, tea.ExitAltScreen)
			}
			t.isAltScreen = !t.isAltScreen
		case "c", "C":
			cmd := func() tea.Msg {
				id := t.table.SelectedRow()[0]
				return t.copyFunc(id, msg.String())
			}
			return t, tea.Batch(cmd)
		case "esc":
			if t.isAltScreen {
				cmds = append(cmds, tea.ExitAltScreen)
				t.isAltScreen = false
			} else {
				return t, tea.Quit
			}
		case "q", "ctrl+c":
			return t, tea.Quit
		case "h", "left":
			t.table.GetViewport().ScrollLeft(5)
			t.viewport.ScrollLeft(5)
		case "l", "right":
			t.table.GetViewport().ScrollRight(5)
			t.viewport.ScrollRight(5)
		case "enter":
			cmd := func() tea.Msg {
				id := t.table.SelectedRow()[0]
				return t.enterFunc(id)
			}
			return t, tea.Batch(cmd)
		}
	}

	updatedTable, cmd := t.table.Update(msg)
	t.table = updatedTable
	cmds = append(cmds, cmd)

	t.viewport.SetContent(t.table.View())
	updatedViewport, vCmd := t.viewport.Update(msg)
	t.viewport = updatedViewport
	cmds = append(cmds, vCmd)

	return t, tea.Batch(cmds...)
}

// View renders the viewport into a string.
func (t *InteractiveTable) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(t.viewport.Width)
	separator := "  »  "

	footer := ""
	if t.showHelp {
		footer = footerStyle.Render(
			strings.Join(t.helpTexts, separator),
		)
	} else {
		footer = footerStyle.Render(
			strings.Join(t.footerTexts, separator),
		)
	}
	return baseStyle.Render(t.viewport.View()) + "\n" + footer
}

// Render renders the final table.
func (t *InteractiveTable) Render() error {
	s := DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.MarkdownBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.table.SetStyles(s)

	if _, err := tea.NewProgram(t).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return nil
}
