package main

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type schedulePanel struct {
	msg           scheduleMsg
	category      category
	selectedMatch *match
	listPanel
}

func newSchedulePanel() schedulePanel {
	return schedulePanel{
		msg:       newScheduleInitialMsg(),
		listPanel: newListPanel(matchDelegate{}),
	}
}

func (s schedulePanel) Init() tea.Cmd {
	return nil
}

func (s schedulePanel) Update(msg tea.Msg) (schedulePanel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if s.msg.isLoading() {
			s.spinner, cmd = s.spinner.Update(msg)
			return s, cmd
		}
		return s, nil
	case categorySelectionMsg:
		s.category = category(msg)
		s.clear()

		cmd = func() tea.Msg {
			return newScheduleLoadingMsg(s.category)
		}
		cmds = append(cmds, cmd)

		cmd = func() tea.Msg {
			schedule, err := fetchSchedule(s.category.ID)
			if err != nil {
				return newScheduleFailedMsg(s.category, err)
			}
			return newScheduleLoadedMsg(s.category, schedule)
		}
		cmds = append(cmds, cmd)

		cmds = append(cmds, s.spinner.Tick)

		cmd = func() tea.Msg {
			return matchSelectionMsg("")
		}
		cmds = append(cmds, cmd)

		return s, tea.Batch(cmds...)
	case scheduleMsg:
		if !s.category.equal(msg.category) {
			return s, nil
		}

		s.msg = msg
		if msg.isSuccess() {
			var items []list.Item
			for _, m := range msg.matches {
				items = append(items, m)
			}
			s.list.SetItems(items)
		}

		if msg.isSuccess() || msg.isFailed() {
			cmd = tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
				schedule, err := fetchSchedule(msg.category.ID)
				if err != nil {
					return newScheduleFailedMsg(s.category, err)
				}
				return newScheduleLoadedMsg(s.category, schedule)
			})
			cmds = append(cmds, cmd)
		}
	}

	s.list, cmd = s.list.Update(msg)
	cmds = append(cmds, cmd)

	selection, ok := s.list.SelectedItem().(match)
	if !ok {
		s.selectedMatch = nil
		return s, tea.Batch(cmds...)
	}

	if s.selectedMatch == nil || selection.MID != s.selectedMatch.MID {
		cmd = func() tea.Msg {
			return matchSelectionMsg(selection.MID)
		}
		cmds = append(cmds, cmd)
		s.selectedMatch = &selection
	}

	return s, tea.Batch(cmds...)
}

func (s schedulePanel) View(focused bool) string {
	return s.render(focused, s.msg.status, s.msg.err)
}

type matchDelegate struct {
}

func (d matchDelegate) Height() int {
	return 4
}

func (d matchDelegate) Spacing() int {
	return 0
}

func (d matchDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d matchDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(match)
	if !ok {
		return
	}

	timeOnly := ""
	startTime, err := time.Parse("2006-01-02 15:04:05", i.StartTime)
	if err == nil {
		timeOnly = startTime.Format("01-02 15:04")
	}
	title := fmt.Sprintf("%s %s", timeOnly, i.MatchDesc)
	title = ansi.Truncate(title, m.Width(), "...")
	title = lipgloss.NewStyle().Width(m.Width()).
		Align(lipgloss.Center).
		Render(title)

	matchPeriod := "未知"
	switch i.MatchPeriod {
	case matchPeriodComing:
		matchPeriod = "未开始"
	case matchPeriodInProgress:
		matchPeriod = fmt.Sprintf("%s %s", i.Quarter, i.QuarterTime)
	case matchPeriodEnd:
		matchPeriod = "已结束"
	}
	matchPeriod = lipgloss.NewStyle().Width(m.Width()).
		Align(lipgloss.Center).
		Render(matchPeriod)

	desc := ""
	if i.RightName == "" {
		desc = ansi.Truncate(i.LeftName, m.Width()-2, "...")
		desc = lipgloss.NewStyle().Width(m.Width() - 2).
			Align(lipgloss.Center).Render(desc)
	} else {
		score := fmt.Sprintf("%s - %s", i.LeftGoal, i.RightGoal)
		nameWith := (m.Width() - 2 - ansi.StringWidth(score)) / 2

		leftName := ansi.Truncate(i.LeftName, nameWith, "...")
		rightName := ansi.Truncate(i.RightName, nameWith, "...")

		desc = fmt.Sprintf("%s %s %s",
			lipgloss.NewStyle().Width(nameWith).Align(lipgloss.Center).Render(leftName),
			lipgloss.NewStyle().Align(lipgloss.Center).Render(score),
			lipgloss.NewStyle().Width(nameWith).Align(lipgloss.Center).Render(rightName),
		)
	}

	style := lipgloss.NewStyle()
	if index == m.Index() {
		style = listFocusedStyle
	}

	content := fmt.Sprintf("%s\n%s\n%s", title, desc, matchPeriod)
	content = style.Width(m.Width()).Render(content)
	fmt.Fprint(w, content+"\n"+divider(m.Width())) //nolint: errcheck
}
