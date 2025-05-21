package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
		s, cmd = s.onStatisticsMsg(msg)
		return s, cmd
	}

	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s statisticsPanel) onStatisticsMsg(msg statisticsMsg) (statisticsPanel, tea.Cmd) {
	if s.matchID != msg.matchID {
		return s, nil
	}
	s.msg = msg

	if !s.msg.isSuccess() {
		return s, nil
	}

	team := s.msg.statistics.team
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		s.goalView(s.msg.statistics.goal, team),
		s.teamView(s.msg.statistics.teamStatistics, team),
		s.playerView(s.msg.statistics.playerStatistics, team),
	)
	s.viewport.SetContent(content)

	if s.shouldRefresh(msg) {
		cmd := tea.Tick(cfg.statisticsRefreshInterval, func(time.Time) tea.Msg {
			staticstics, err := fetchStatistics(msg.matchID)
			if err != nil {
				return newStatisticsFailedMsg(msg.matchID, err)
			}
			return newStatisticsLoadedMsg(msg.matchID, staticstics)
		})
		return s, cmd
	}

	return s, nil
}

func (s statisticsPanel) shouldRefresh(msg statisticsMsg) bool {
	if msg.isFailed() {
		return true
	}

	if msg.isSuccess() && msg.statistics.livePeriod != periodEnd {
		return true
	}

	return false
}

func (s statisticsPanel) View(focused bool) string {
	style := borderStyle
	if focused {
		style = borderFocusedStyle
	}
	style = style.Width(s.viewport.Width).Height(s.viewport.Height).AlignHorizontal(lipgloss.Center)

	if s.msg.isInitial() {
		return style.Render("")
	}

	if s.msg.isLoading() {
		return style.Render(s.spinner.View() + "加载中...")
	}

	if s.msg.isFailed() {
		return style.Render("加载失败: " + s.msg.err.Error())
	}

	if s.msg.statistics == nil {
		return style.Render("没有数据")
	}

	return style.Render(s.viewport.View())
}

func (s statisticsPanel) goalView(goal *goalStatistics, team *team) string {
	if goal == nil || team == nil {
		return ""
	}

	columns := []table.Column{
		{Title: "", Width: team.width()},
	}
	for _, v := range goal.Head {
		columns = append(columns, table.Column{Title: v, Width: 6})
	}

	rows := []table.Row{
		{team.LeftName},
		{team.RightName},
	}
	for i := range rows {
		rows[i] = append(rows[i], goal.Rows[i]...)
	}

	t := table.New(
		table.WithFocused(false),
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(1+len(rows)),
		table.WithStyles(table.Styles{
			Header: lipgloss.NewStyle().Padding(0, 1),
			Cell:   lipgloss.NewStyle().Padding(0, 1),
		}),
	)
	return lipgloss.NewStyle().
		Width(s.viewport.Width).
		AlignHorizontal(lipgloss.Center).
		Render(t.View()) + "\n"
}

func (s *statisticsPanel) teamView(statistics []teamStatistics, team *team) string {
	if len(statistics) == 0 || team == nil {
		return ""
	}

	width := s.viewport.Width - 6
	if width <= 0 {
		return ""
	}

	leftRightWidth := 0
	textWidth := 0

	for _, v := range statistics {
		t := ansi.StringWidth(v.Text)
		if t > textWidth {
			textWidth = t
		}
		l := ansi.StringWidth(v.LeftVal)
		if l > leftRightWidth {
			leftRightWidth = l
		}
		r := ansi.StringWidth(v.RightVal)
		if r > leftRightWidth {
			leftRightWidth = r
		}
	}

	teamRow := fmt.Sprintf("%s%s%s",
		team.LeftName,
		lipgloss.NewStyle().Width(textWidth).Align(lipgloss.Center).Render("vs"),
		team.RightName,
	)
	teamRow = lipgloss.NewStyle().
		Width(width).
		AlignHorizontal(lipgloss.Center).
		Render(teamRow)

	progressBarWidth := (width - leftRightWidth*2 - textWidth) / 2
	if progressBarWidth <= 0 {
		return ""
	}

	rows := []string{teamRow}
	for _, v := range statistics {
		leftVal := strings.TrimSuffix(v.LeftVal, "%")
		rightVal := strings.TrimSuffix(v.RightVal, "%")
		leftPercent := 0.0
		rightPencent := 0.0
		left, err1 := strconv.ParseFloat(leftVal, 64)
		right, err2 := strconv.ParseFloat(rightVal, 64)
		if err1 == nil && err2 == nil {
			if left+right != 0 {
				leftPercent = left / (left + right)
				rightPencent = right / (left + right)
			}
		}
		leftWidth := int(leftPercent * float64(progressBarWidth))
		if leftWidth == 0 {
			leftWidth = 1
		}
		rightWidth := int(rightPencent * float64(progressBarWidth))
		if rightWidth == 0 {
			rightWidth = 1
		}
		leftStyle := lipgloss.NewStyle().Width(progressBarWidth).Align(lipgloss.Right)
		rightStyle := lipgloss.NewStyle().Width(progressBarWidth).Align(lipgloss.Left)
		if leftPercent > rightPencent {
			leftStyle = leftStyle.Foreground(focusedColor)
		} else if rightPencent > leftPercent {
			rightStyle = rightStyle.Foreground(focusedColor)
		}
		row := fmt.Sprintf(" %s %s %s %s %s ",
			leftStyle.Width(leftRightWidth).Align(lipgloss.Left).Render(v.LeftVal),
			leftStyle.Render(strings.Repeat("━", leftWidth)),
			lipgloss.NewStyle().Width(textWidth).AlignHorizontal(lipgloss.Center).Render(v.Text),
			rightStyle.Render(strings.Repeat("━", rightWidth)),
			rightStyle.Width(leftRightWidth).Align(lipgloss.Right).Render(v.RightVal),
		)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n") + "\n"
}

func (s statisticsPanel) playerView(statistics [][]playerStatistics, team *team) string {
	if len(statistics) == 0 || team == nil {
		return ""
	}

	var tables []table.Model
	columnWidthes := s.columnWidthes(statistics)

	for _, v := range statistics {
		columns := []table.Column{}
		rows := []table.Row{}
		for _, vv := range v {
			for k, h := range vv.Head {
				columns = append(columns, table.Column{Title: h, Width: columnWidthes[k]})
			}
			if len(vv.Row) > 0 {
				if len(vv.Row) > 2 && vv.Row[2] == "0'0\"" {
					continue
				}
				rows = append(rows, table.Row(vv.Row))
			}
		}
		t := table.New(
			table.WithFocused(false),
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithHeight(1+len(rows)),
			table.WithStyles(table.Styles{
				Header: lipgloss.NewStyle().Padding(0, 1),
				Cell:   lipgloss.NewStyle().Padding(0, 1),
			}),
		)
		tables = append(tables, t)
	}

	if len(tables) >= 2 {
		tables[0].Columns()[0].Title = team.LeftName
		tables[1].Columns()[0].Title = team.RightName
	}

	content := make([]string, len(tables))
	for k, v := range tables {
		content[k] = v.View() + "\n"
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content...,
	)
}

func (s statisticsPanel) columnWidthes(statistics [][]playerStatistics) []int {
	var columnWidth []int
	for _, v := range statistics {
		for _, vv := range v {
			for _, h := range vv.Head {
				columnWidth = append(columnWidth, ansi.StringWidth(h))
			}
			for k, r := range vv.Row {
				w := ansi.StringWidth(r)
				if w > columnWidth[k] {
					columnWidth[k] = w
				}
			}
		}
	}

	for k, v := range columnWidth {
		if v > 14 {
			columnWidth[k] = 14
		}
	}

	return columnWidth
}

func (s *statisticsPanel) SetSize(width, height int) {
	s.viewport.Width = width
	s.viewport.Height = height
}
