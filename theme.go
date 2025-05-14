package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var borderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

var borderFocusedStyle = lipgloss.NewStyle().
	Border(lipgloss.ThickBorder()).
	BorderForeground(focusedColor)

var listFocusedStyle = lipgloss.NewStyle().
	Foreground(focusedColor).
	Bold(true)

var focusedColor = lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}

func divider(width int) string {
	s := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})
	return s.Render(strings.Repeat("─", width))
}
