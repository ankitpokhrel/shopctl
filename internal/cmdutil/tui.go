package cmdutil

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/mgutz/ansi"
)

const wordWrap = 120

// MDRenderer constructs markdown renderer.
func MDRenderer() (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithEnvironmentConfig(),
		glamour.WithWordWrap(wordWrap),
	)
}

// NoTTYRenderer constructs renderer for non-TTY env.
func NoTTYRenderer() (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithStandardStyle("notty"),
		glamour.WithWordWrap(wordWrap),
	)
}

// Pad pads the msg with spaces to the given limit.
func Pad(msg string, limit int) string {
	var out strings.Builder
	out.WriteString(msg)
	for i := len(msg); i < limit; i++ {
		out.WriteRune(' ')
	}
	return out.String()
}

// ShortenAndPad shortens the msg to the given limit and pads with spaces if necessary.
func ShortenAndPad(msg string, limit int) string {
	if limit > 1 && len(msg) > limit {
		return msg[0:limit-1] + "…"
	}
	return Pad(msg, limit)
}

// Success prints success message in stdout.
func Success(msg string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, fmt.Sprintf("\u001B[0;32m✔\u001B[0m %s\n", msg), args...)
}

// Fail prints failure message in stderr.
func Fail(msg string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\u001B[0;31m✗\u001B[0m %s\n", msg), args...)
}

// IsDumbTerminal checks TERM/WT_SESSION environment variable and returns true if they indicate a dumb terminal.
//
// Dumb terminal indicates terminal with limited capability. It may not provide support
// for special character sequences, e.g., no handling of ANSI escape sequences.
func IsDumbTerminal() bool {
	term := strings.ToLower(os.Getenv("TERM"))
	_, wtSession := os.LookupEnv("WT_SESSION")
	return !wtSession && (term == "" || term == "dumb")
}

// IsNotTTY returns true if the stdout file descriptor is not a TTY.
func IsNotTTY() bool {
	return !isatty.IsTerminal(os.Stdout.Fd())
}

// GetPager returns configured pager.
func GetPager() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	if IsDumbTerminal() {
		return "cat"
	}
	pager := os.Getenv("SHOPIFY_PAGER")
	if pager == "" {
		pgr := os.Getenv("PAGER")
		if pgr == "" {
			pager = "less"
		} else {
			pager = pgr
		}
	}
	return pager
}

// PagerOut try to output to configured pager.
func PagerOut(out string) error {
	pagerCmd := GetPager()
	if pagerCmd == "" {
		_, err := fmt.Print(out)
		return err
	}

	pa := strings.Split(pagerCmd, " ")
	pager, pagerArgs := pa[0], pa[1:]
	if err := cmdExists(pager); err != nil {
		return err
	}

	pagerEnv := os.Environ()
	for i := len(pagerEnv) - 1; i >= 0; i-- {
		if strings.HasPrefix(pagerEnv[i], "PAGER=") {
			pagerEnv = append(pagerEnv[0:i], pagerEnv[i+1:]...)
		}
	}
	if _, ok := os.LookupEnv("LESS"); !ok {
		pagerEnv = append(pagerEnv, "LESS=R")
	}

	cmd := exec.Command(pager, pagerArgs...)
	cmd.Env = pagerEnv
	cmd.Stdin = strings.NewReader(out)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// DiffOut outputs result of a diff operation.
func DiffOut(out string) error {
	var diffTool string

	diffTool = os.Getenv("SHOPIFY_DIFF_TOOL")
	if diffTool == "" {
		diffTool = os.Getenv("GIT_EXTERNAL_DIFF")
	}
	if diffTool == "" {
		if IsDumbTerminal() || IsNotTTY() {
			fmt.Print(out)
			return nil
		}
		return PagerOut(out)
	}

	pa := strings.Split(diffTool, " ")
	tool, args := pa[0], pa[1:]
	if err := cmdExists(tool); err != nil {
		return err
	}

	cmd := exec.Command(tool, args...)
	cmd.Stdin = strings.NewReader(out)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ColoredOut decorates msg with color.
func ColoredOut(msg string, clr color.Attribute, attrs ...color.Attribute) string {
	c := color.New(clr).Add(attrs...)
	return c.Sprint(msg)
}

// GetEditor returns the default editor configured in the system.
func GetEditor() (string, []string) {
	envCheck := func() string {
		editor := os.Getenv("SHOPIFY_EDITOR")
		if editor != "" {
			return editor
		}
		editor = os.Getenv("VISUAL")
		if editor != "" {
			return editor
		}
		editor = os.Getenv("EDITOR")
		if editor != "" {
			return editor
		}

		if runtime.GOOS == "windows" {
			return "notepad"
		}

		for _, e := range []string{"vim", "vi", "nano"} {
			if path, err := exec.LookPath(e); err == nil {
				return path
			}
		}
		return ""
	}
	editor := envCheck()

	parts := strings.Fields(editor)
	if len(parts) > 1 {
		return parts[0], parts[1:]
	}
	return editor, nil
}

// Gray returns gray colored msg.
func Gray(msg string) string {
	if xterm256() {
		return gray256(msg)
	}
	return ansi.ColorFunc("black+h")(msg)
}

func xterm256() bool {
	term := os.Getenv("TERM")
	return strings.Contains(term, "-256color")
}

func gray256(msg string) string {
	return fmt.Sprintf("\x1b[38;5;242m%s\x1b[m", msg)
}

func cmdExists(cmd string) error {
	_, err := exec.LookPath(cmd)
	return err
}
