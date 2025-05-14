package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	categoryPanelWidth = 16
	schedulePanelWidth = 32
)

type app struct {
	categoryPanel   categoryPanel
	schedulePanel   schedulePanel
	focus           focus
	availableHeight int
}

func newApp() app {
	return app{
		categoryPanel: newCategoryPanel(categoryPanelWidth),
		schedulePanel: newSchedulePanel(schedulePanelWidth),
		focus:         focusCategory,
	}
}

func (a app) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("SportX"),
		a.categoryPanel.Init(),
		a.schedulePanel.Init(),
	)
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case categoriesMsg:
		a.categoryPanel, cmd = a.categoryPanel.Update(msg)
		return a, cmd
	case categorySelectionMsg:
		a.schedulePanel, cmd = a.schedulePanel.Update(msg)
		return a, cmd
	case scheduleMsg:
		a.schedulePanel, cmd = a.schedulePanel.Update(msg)
		return a, cmd
	case tea.WindowSizeMsg:
		a.availableHeight = msg.Height - borderStyle.GetVerticalBorderSize()
		a.categoryPanel.SetHeight(a.availableHeight)
		a.schedulePanel.SetHeight(a.availableHeight)
	case spinner.TickMsg:
		var cmds []tea.Cmd
		a.categoryPanel, cmd = a.categoryPanel.Update(msg)
		cmds = append(cmds, cmd)
		a.schedulePanel, cmd = a.schedulePanel.Update(msg)
		cmds = append(cmds, cmd)
		return a, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyTab.String():
			a.focus = a.focus.next()
			return a, nil
		case tea.KeyShiftTab.String():
			a.focus = a.focus.prev()
			return a, nil
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	}

	switch a.focus {
	case focusCategory:
		a.categoryPanel, cmd = a.categoryPanel.Update(msg)
		return a, cmd
	case focusSchedule:
		a.schedulePanel, cmd = a.schedulePanel.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a app) View() string {
	categoryView := a.categoryPanel.View()
	scheduleView := a.schedulePanel.View()

	switch a.focus {
	case focusCategory:
		categoryView = borderFocusedStyle.
			Width(categoryPanelWidth).
			Height(a.availableHeight).
			Render(categoryView)
		scheduleView = borderStyle.
			Width(schedulePanelWidth).
			Height(a.availableHeight).
			Render(scheduleView)
	case focusSchedule:
		categoryView = borderStyle.
			Width(categoryPanelWidth).
			Height(a.availableHeight).
			Render(categoryView)
		scheduleView = borderFocusedStyle.
			Width(schedulePanelWidth).
			Height(a.availableHeight).
			Render(scheduleView)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, categoryView, scheduleView)
}

type focus int

func (f focus) next() focus {
	return (f + 1) % 2
}

func (f focus) prev() focus {
	return (f - 1 + 2) % 2
}

const (
	focusCategory focus = iota
	focusSchedule
)
