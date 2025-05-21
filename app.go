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
		categoryPanel:   newCategoryPanel(),
		schedulePanel:   newSchedulePanel(),
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
		a.onWindowSizeMsg(msg)
		return a, nil
	case spinner.TickMsg:
		return a.onSpinnerTickMsg(msg)
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
	return lipgloss.JoinHorizontal(lipgloss.Left,
		a.categoryPanel.View(a.focus == focusCategory),
		a.schedulePanel.View(a.focus == focusSchedule),
		a.statisticsPanel.View(a.focus == focusStatistics),
		a.textLivePanel.View(),
	)
}

func (a app) onSpinnerTickMsg(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	a.categoryPanel, cmd = a.categoryPanel.Update(msg)
	cmds = append(cmds, cmd)
	a.schedulePanel, cmd = a.schedulePanel.Update(msg)
	cmds = append(cmds, cmd)
	a.textLivePanel, cmd = a.textLivePanel.Update(msg)
	cmds = append(cmds, cmd)
	a.statisticsPanel, cmd = a.statisticsPanel.Update(msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a *app) onWindowSizeMsg(msg tea.WindowSizeMsg) {
	statisticsWidth := msg.Width - 4*borderStyle.GetHorizontalBorderSize() - categoryPanelWidth - schedulePanelWidth - textLivePanelWidth
	a.availableHeight = msg.Height - borderStyle.GetVerticalBorderSize()

	a.categoryPanel.setSize(categoryPanelWidth, a.availableHeight)
	a.schedulePanel.setSize(schedulePanelWidth, a.availableHeight)
	a.textLivePanel.SetHeight(a.availableHeight)
	a.statisticsPanel.SetSize(statisticsWidth, a.availableHeight)
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
