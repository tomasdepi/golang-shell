package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

type Prompt struct{}

func (p Prompt) isWithinGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func (p Prompt) getGitBranch() (string, bool) {
	out, err := exec.Command(
		"git",
		"symbolic-ref",
		"--short",
		"HEAD",
	).Output()

	if err != nil {
		return "", false
	}

	return strings.TrimSpace(string(out)), true
}

func (p Prompt) Render() {

	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	cwd, _ := os.Getwd()
	now := time.Now().Format("15:04:05")

	width, _, _ := term.GetSize(os.Stdout.Fd())

	hostStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#fde68a")). // amber
		Bold(true)

	userStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a7f3d0")) // mint

	cwdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#bfdbfe")) // blue

	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#fbcfe8")) // pinkish

	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff"))

	host := hostStyle.Render(hostname)
	user = userStyle.Render(user)
	cwd = cwdStyle.Render(cwd)
	time := timeStyle.Render(now)
	sep := sepStyle.Render(" • ")

	elements := []string{host,
		sep,
		user,
		sep,
		cwd,
		sep,
		time,
	}

	isGit := p.isWithinGitRepo()
	if isGit {
		branch, _ := p.getGitBranch()
		elements = append(elements, sep, userStyle.Render(branch))
	}

	if branch, ok := p.getGitBranch(); ok {
		icon := "\ue725"

		git := lipgloss.JoinHorizontal(
			lipgloss.Left,
			userStyle.Render(icon),
			userStyle.Render(branch),
		)

		elements = append(elements, sep, git)
	}

	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		elements...,
	)

	headerStyle := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("#ff69b4")).
		// Background(lipgloss.Color("#2e9aaaff")).
		Foreground(lipgloss.Color("#ffffff"))

	fmt.Println(headerStyle.Render(content))
}
