package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	categoryPanelWidth = 16
)

type app struct {
	categoryPanel   categoryPanel
	focus           focus
	availableHeight int
}

func newApp() app {
	return app{
		categoryPanel: newCategoryPanel(categoryPanelWidth),
		focus:         focusCategory,
	}
}

func (a app) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("SportX"),
		a.categoryPanel.Init(),
	)
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case categoriesMsg:
		a.categoryPanel, cmd = a.categoryPanel.Update(msg)
		return a, cmd
	case tea.WindowSizeMsg:
		a.availableHeight = msg.Height - borderStyle.GetVerticalBorderSize()
		a.categoryPanel.SetHeight(a.availableHeight)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	}

	a.categoryPanel, cmd = a.categoryPanel.Update(msg)
	return a, cmd
}

func (a app) View() string {
	return borderFocusedStyle.Height(a.availableHeight).Width(categoryPanelWidth).Render(a.categoryPanel.View())
}

type focus int

const (
	focusCategory focus = iota
)
