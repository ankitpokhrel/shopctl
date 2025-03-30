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

// Table wraps bubble table and viewport.
type Table struct {
	table       Model
	viewport    viewport.Model
	isAltScreen bool
}

// NewTable builds a new table.
func NewTable(cols []Column, rows []Row) *Table {
	t := New(
		WithColumns(cols),
		WithRows(rows),
		WithFocused(true),
		WithHeight(25),
	)

	vp := viewport.New(200, 26)
	vp.SetContent(t.View())

	return &Table{
		table:    t,
		viewport: vp,
	}
}

// Init is required for initialization.
func (t *Table) Init() tea.Cmd { return nil }

// Update is the Bubble Tea update loop.
func (t *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) { //nolint:gocritic
	case tea.KeyMsg:
		switch msg.String() {
		case "m":
			if !t.isAltScreen {
				cmds = append(cmds, tea.EnterAltScreen)
			} else {
				cmds = append(cmds, tea.ExitAltScreen)
			}
			t.isAltScreen = !t.isAltScreen
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
func (t *Table) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	helpText := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		strings.Join([]string{
			"↑ k/j ↓: Navigate top & down",
			"← h/l →: Navigate left & right",
			"m: Toggle distraction free mode",
			"q/CTRL+c/ESC: Quit",
		}, " • "),
	)

	return baseStyle.Render(t.viewport.View()) + "\n" + helpText
}

// Render renders the final table.
func (t *Table) Render() error {
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
