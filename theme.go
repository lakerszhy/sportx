package main

import "github.com/charmbracelet/lipgloss"

var borderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

var borderFocusedStyle = lipgloss.NewStyle().
	Border(lipgloss.ThickBorder()).
	BorderForeground(focusedColor)

var listFocusedStyle = lipgloss.NewStyle().
	Foreground(focusedColor).
	Bold(true)

var focusedColor = lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}
