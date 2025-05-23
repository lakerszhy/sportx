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

type statsPanel struct {
	spinner  spinner.Model
	matchID  string
	viewport viewport.Model
	msg      statsMsg
}

func newStatsPanel() statsPanel {
	vp := viewport.New(0, 0)
	vp.SetHorizontalStep(3) //nolint:mnd // 水平移动距离
	return statsPanel{
		viewport: vp,
		msg:      newStatsInitialMsg(),
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
		),
	}
}

func (s statsPanel) Init() tea.Cmd {
	return nil
}

func (s statsPanel) Update(msg tea.Msg) (statsPanel, tea.Cmd) {
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
				return newStatsInitialMsg()
			}
			return s, cmd
		}

		cmd = func() tea.Msg {
			return newStatsLoadingMsg(s.matchID)
		}
		cmds = append(cmds, cmd)

		cmd = func() tea.Msg {
			stats, err := fetchStats(s.matchID)
			if err != nil {
				return newStatsFailedMsg(s.matchID, err)
			}
			return newStatsLoadedMsg(s.matchID, stats)
		}
		cmds = append(cmds, cmd)

		cmds = append(cmds, s.spinner.Tick)

		return s, tea.Batch(cmds...)
	case statsMsg:
		s, cmd = s.onStatsMsg(msg)
		return s, cmd
	}

	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s statsPanel) onStatsMsg(msg statsMsg) (statsPanel, tea.Cmd) {
	if s.matchID != msg.matchID {
		return s, nil
	}
	s.msg = msg

	if s.msg.isSuccess() {
		team := s.msg.stats.team
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			s.goalView(s.msg.stats.goal, team),
			s.teamView(s.msg.stats.teamStats, team),
			s.playerView(s.msg.stats.playerStats, team),
		)
		s.viewport.SetContent(content)
	}

	if s.shouldRefresh(msg) {
		cmd := tea.Tick(cfg.statsRefreshInterval, func(time.Time) tea.Msg {
			staticstics, err := fetchStats(msg.matchID)
			if err != nil {
				return newStatsFailedMsg(msg.matchID, err)
			}
			return newStatsLoadedMsg(msg.matchID, staticstics)
		})
		return s, cmd
	}

	return s, nil
}

func (s statsPanel) shouldRefresh(msg statsMsg) bool {
	if msg.isFailed() {
		return true
	}

	if msg.isSuccess() && msg.stats.livePeriod != periodEnd {
		return true
	}

	return false
}

func (s statsPanel) View(focused bool) string {
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

	if s.msg.stats == nil {
		return style.Render("没有数据")
	}

	return style.Render(s.viewport.View())
}

func (s statsPanel) goalView(goal *goalStats, team *team) string {
	if goal == nil || team == nil {
		return ""
	}

	columns := []table.Column{
		{Title: "", Width: team.width()},
	}
	for _, v := range goal.Head {
		columns = append(columns, table.Column{Title: v, Width: 6}) //nolint:mnd // cell宽度
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

func (s *statsPanel) teamView(stats []teamStats, team *team) string {
	if len(stats) == 0 || team == nil {
		return ""
	}

	width := s.viewport.Width - 6 //nolint:mnd // 列之间的pandding
	if width <= 0 {
		return ""
	}

	leftRightWidth := 0
	textWidth := 0

	for _, v := range stats {
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

	progressBarWidth := (width - leftRightWidth*2 - textWidth) / 2 //nolint:mnd // 进度条宽度
	if progressBarWidth <= 0 {
		return ""
	}

	rows := []string{teamRow}
	for _, v := range stats {
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

func (s statsPanel) playerView(stats [][]playerStats, team *team) string {
	if len(stats) == 0 || team == nil {
		return ""
	}

	var tables []table.Model
	columnWidthes := s.columnWidthes(stats)

	for _, v := range stats {
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

	if len(tables) >= 2 { //nolint:mnd // 每只球队一个table
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

func (s statsPanel) columnWidthes(stats [][]playerStats) []int {
	var columnWidth []int
	for _, v := range stats {
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

	maxWidth := 14
	for k, v := range columnWidth {
		if v > maxWidth {
			columnWidth[k] = maxWidth
		}
	}

	return columnWidth
}

func (s *statsPanel) SetSize(width, height int) {
	s.viewport.Width = width
	s.viewport.Height = height
}
