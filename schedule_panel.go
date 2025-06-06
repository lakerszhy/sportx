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
		return s, s.onCategorySelectionMsg(msg)
	case scheduleMsg:
		cmd = s.onScheduleMsg(msg)
		cmds = append(cmds, cmd)
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

func (s *schedulePanel) onCategorySelectionMsg(msg categorySelectionMsg) tea.Cmd {
	s.category = category(msg)
	s.list.SetItems([]list.Item{})

	var cmds []tea.Cmd

	cmd := func() tea.Msg {
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

	return tea.Batch(cmds...)
}

func (s *schedulePanel) onScheduleMsg(msg scheduleMsg) tea.Cmd {
	if !s.category.equal(msg.category) {
		return nil
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
		return tea.Tick(cfg.scheduleRefreshInterval, func(time.Time) tea.Msg {
			schedule, err := fetchSchedule(msg.category.ID)
			if err != nil {
				return newScheduleFailedMsg(s.category, err)
			}
			return newScheduleLoadedMsg(s.category, schedule)
		})
	}

	return nil
}

func (s schedulePanel) View(focused bool) string {
	return s.render(focused, s.msg.status, s.msg.err)
}

type matchDelegate struct {
}

func (d matchDelegate) Height() int {
	return 4 //nolint:mnd // 每场比赛显示4行
}

func (d matchDelegate) Spacing() int {
	return 0
}

func (d matchDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d matchDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(match)
	if !ok {
		return
	}

	width := m.Width() - 2 //nolint:mnd // 左右padding

	timeOnly := ""
	startTime, err := time.Parse("2006-01-02 15:04:05", i.StartTime)
	if err == nil {
		timeOnly = startTime.Format("01-02 15:04")
	}
	title := fmt.Sprintf("%s %s", timeOnly, i.MatchDesc)
	title = ansi.Truncate(title, width, "...")
	title = lipgloss.NewStyle().Width(width).
		Margin(0, 1).
		Align(lipgloss.Center).
		Render(title)

	matchPeriod := "未知"
	switch i.MatchPeriod {
	case periodComing:
		matchPeriod = "未开始"
	case periodInProgress:
		matchPeriod = fmt.Sprintf("%s %s", i.Quarter, i.QuarterTime)
	case periodEnd:
		matchPeriod = "已结束"
	}
	matchPeriod = lipgloss.NewStyle().Width(m.Width()).
		Align(lipgloss.Center).
		Render(matchPeriod)

	desc := ansi.Truncate(i.LeftName, width, "...")
	if i.RightName != "" {
		score := fmt.Sprintf("%s - %s", i.LeftGoal, i.RightGoal)
		nameWith := (width - 2 - ansi.StringWidth(score)) / 2 //nolint:mnd // 两只队伍各占一半

		leftName := ansi.Truncate(i.LeftName, nameWith, "...")
		rightName := ansi.Truncate(i.RightName, nameWith, "...")

		desc = fmt.Sprintf("%s %s %s",
			lipgloss.NewStyle().Width(nameWith).Align(lipgloss.Center).Render(leftName),
			lipgloss.NewStyle().Align(lipgloss.Center).Render(score),
			lipgloss.NewStyle().Width(nameWith).Align(lipgloss.Center).Render(rightName),
		)
	}
	desc = lipgloss.NewStyle().Width(width).Margin(0, 1).
		Align(lipgloss.Center).Render(desc)

	style := lipgloss.NewStyle()
	if index == m.Index() {
		style = listFocusedStyle
	}

	content := fmt.Sprintf("%s\n%s\n%s", title, desc, matchPeriod)
	content = style.Width(m.Width()).Render(content)
	fmt.Fprint(w, content+"\n"+divider(m.Width()))
}
