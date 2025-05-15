package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type statisticsPanel struct {
	spinner  spinner.Model
	matchID  string
	viewport viewport.Model
	msg      statisticsMsg
}

func newStatisticsPanel() statisticsPanel {
	vp := viewport.New(0, 0)
	vp.SetHorizontalStep(3)
	return statisticsPanel{
		viewport: vp,
		msg:      newStatisticsInitialMsg(),
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
		),
	}
}

func (s statisticsPanel) Init() tea.Cmd {
	return nil
}

func (s statisticsPanel) Update(msg tea.Msg) (statisticsPanel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if s.msg.isLoading() {
			s.spinner, cmd = s.spinner.Update(msg)
			return s, cmd
		}
		return s, nil
	case matchSelectionMsg:
		s.matchID = string(msg)

		if s.matchID == "" {
			cmd = func() tea.Msg {
				return newStatisticsInitialMsg()
			}
			return s, cmd
		}

		cmd = func() tea.Msg {
			return newStatisticsLoadingMsg(s.matchID)
		}
		cmds = append(cmds, cmd)

		cmd = func() tea.Msg {
			statistics, err := fetchStatistics(s.matchID)
			if err != nil {
				return newStatisticsFailedMsg(s.matchID, err)
			}
			return newStatisticsLoadedMsg(s.matchID, statistics)
		}
		cmds = append(cmds, cmd)

		cmds = append(cmds, s.spinner.Tick)

		return s, tea.Batch(cmds...)
	case statisticsMsg:
		if s.matchID != msg.matchID {
			return s, nil
		}
		s.msg = msg
	}

	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s statisticsPanel) View(focused bool) string {
	style := borderStyle
	if focused {
		style = borderFocusedStyle
	}
	style = style.Width(s.viewport.Width).Height(s.viewport.Height)

	if s.msg.isInitial() {
		return style.AlignHorizontal(lipgloss.Center).Render("")
	}

	if s.msg.isLoading() {
		return style.AlignHorizontal(lipgloss.Center).
			Render(s.spinner.View() + "加载中...")
	}

	if s.msg.isFailed() {
		return style.AlignHorizontal(lipgloss.Center).
			Render("加载失败: " + s.msg.err.Error())
	}

	if s.msg.statistics == nil {
		return style.AlignHorizontal(lipgloss.Center).
			Render("没有数据")
	}

	content := ""
	goal := s.msg.statistics.goal
	columns := []table.Column{}
	rows := []table.Row{}
	for _, v := range goal.Head {
		columns = append(columns, table.Column{Title: v, Width: 6})
	}
	for i := range goal.Rows {
		rows = append(rows, goal.Rows[i])
	}

	t := table.New(
		table.WithFocused(false),
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(1+len(rows)),
	)
	content = t.View()

	s.viewport.SetContent(content)

	return style.Render(s.viewport.View())
}

func (s *statisticsPanel) SetSize(width, height int) {
	s.viewport.Width = width
	s.viewport.Height = height
}
