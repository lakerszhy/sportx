package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	categoryPanelWidth = 16
	schedulePanelWidth = 32
	textLivePanelWidth = 48
)

type app struct {
	categoryPanel   categoryPanel
	schedulePanel   schedulePanel
	textLivePanel   textLivePanel
	statisticsPanel statisticsPanel
	focus           focus
	availableHeight int
}

func newApp() app {
	return app{
		categoryPanel:   newCategoryPanel(categoryPanelWidth),
		schedulePanel:   newSchedulePanel(schedulePanelWidth),
		textLivePanel:   newTextLivePanel(textLivePanelWidth),
		statisticsPanel: newStatisticsPanel(),
		focus:           focusCategory,
	}
}

func (a app) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("SportX"),
		a.categoryPanel.Init(),
		a.schedulePanel.Init(),
		a.textLivePanel.Init(),
	)
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

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
	case matchSelectionMsg:
		a.textLivePanel, cmd = a.textLivePanel.Update(msg)
		cmds = append(cmds, cmd)
		a.statisticsPanel, cmd = a.statisticsPanel.Update(msg)
		cmds = append(cmds, cmd)
		return a, tea.Batch(cmds...)
	case textLivesMsg:
		a.textLivePanel, cmd = a.textLivePanel.Update(msg)
		return a, cmd
	case statisticsMsg:
		a.statisticsPanel, cmd = a.statisticsPanel.Update(msg)
		return a, cmd
	case tea.WindowSizeMsg:
		a.availableHeight = msg.Height - borderStyle.GetVerticalBorderSize()
		availableWidth := msg.Width - 4*borderStyle.GetHorizontalBorderSize() - categoryPanelWidth - schedulePanelWidth - textLivePanelWidth
		a.categoryPanel.SetHeight(a.availableHeight)
		a.schedulePanel.SetHeight(a.availableHeight)
		a.textLivePanel.SetHeight(a.availableHeight)
		a.statisticsPanel.SetSize(availableWidth, a.availableHeight)
	case spinner.TickMsg:
		a.categoryPanel, cmd = a.categoryPanel.Update(msg)
		cmds = append(cmds, cmd)
		a.schedulePanel, cmd = a.schedulePanel.Update(msg)
		cmds = append(cmds, cmd)
		a.textLivePanel, cmd = a.textLivePanel.Update(msg)
		cmds = append(cmds, cmd)
		a.statisticsPanel, cmd = a.statisticsPanel.Update(msg)
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
	case focusStatistics:
		a.statisticsPanel, cmd = a.statisticsPanel.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a app) View() string {
	categoryView := a.categoryPanel.View()
	scheduleView := a.schedulePanel.View()
	textLiveView := borderStyle.Width(textLivePanelWidth).
		Height(a.availableHeight).
		Render(a.textLivePanel.View())
	statisticsView := a.statisticsPanel.View()

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
		statisticsView = borderStyle.Render(statisticsView)
	case focusSchedule:
		categoryView = borderStyle.
			Width(categoryPanelWidth).
			Height(a.availableHeight).
			Render(categoryView)
		scheduleView = borderFocusedStyle.
			Width(schedulePanelWidth).
			Height(a.availableHeight).
			Render(scheduleView)
		statisticsView = borderStyle.Render(statisticsView)
	case focusStatistics:
		categoryView = borderStyle.
			Width(categoryPanelWidth).
			Height(a.availableHeight).
			Render(categoryView)
		scheduleView = borderStyle.
			Width(schedulePanelWidth).
			Height(a.availableHeight).
			Render(scheduleView)
		statisticsView = borderFocusedStyle.Render(statisticsView)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left,
		categoryView,
		scheduleView,
		textLiveView,
		statisticsView,
	)
}

type focus int

func (f focus) next() focus {
	return (f + 1) % 3
}

func (f focus) prev() focus {
	return (f - 1 + 3) % 3
}

const (
	focusCategory focus = iota
	focusSchedule
	focusStatistics
)
