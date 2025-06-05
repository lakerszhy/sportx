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

	s.updateContent()

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

func (s *statsPanel) updateContent() {
	if !s.msg.isSuccess() || s.msg.stats.team == nil {
		return
	}

	content := []string{}
	goalView := s.goalView()
	if goalView != "" {
		content = append(content, goalView)
	}
	teamView := s.teamView()
	if teamView != "" {
		content = append(content, teamView)
	}
	playerView := s.playerView()
	if playerView != "" {
		content = append(content, playerView)
	}

	if len(content) == 0 {
		s.viewport.SetContent("")
	} else {
		s.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, content...))
	}
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

	content := s.viewport.View()
	if strings.TrimSpace(content) == "" {
		return style.Render("暂无数据")
	}

	return style.Render(content)
}

func (s statsPanel) goalView() string {
	goalStats := s.msg.stats.goal
	if goalStats == nil {
		return ""
	}

	columns := []table.Column{
		{Title: "", Width: s.msg.stats.team.width()},
	}
	for _, v := range goalStats.Head {
		columns = append(columns, table.Column{Title: v, Width: 6}) //nolint:mnd // cell宽度
	}

	rows := []table.Row{
		{s.msg.stats.team.LeftName},
		{s.msg.stats.team.RightName},
	}
	for i := range rows {
		rows[i] = append(rows[i], goalStats.Rows[i]...)
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

func (s *statsPanel) teamView() string {
	stats := s.msg.stats.teamStats
	if len(stats) == 0 {
		return ""
	}

	totalWidth, valueWidth, itemWidth, progressBarWidth := s.calcWidths(stats)

	teamRow := fmt.Sprintf("%s%s%s",
		s.msg.stats.team.LeftName,
		lipgloss.NewStyle().Width(itemWidth).Align(lipgloss.Center).Render("vs"),
		s.msg.stats.team.RightName,
	)
	teamRow = lipgloss.NewStyle().
		Width(totalWidth).
		AlignHorizontal(lipgloss.Center).
		Render(teamRow)

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
			leftStyle.Width(valueWidth).Align(lipgloss.Left).Render(v.LeftVal),
			leftStyle.Render(strings.Repeat("━", leftWidth)),
			lipgloss.NewStyle().Width(itemWidth).AlignHorizontal(lipgloss.Center).Render(v.Text),
			rightStyle.Render(strings.Repeat("━", rightWidth)),
			rightStyle.Width(valueWidth).Align(lipgloss.Right).Render(v.RightVal),
		)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n") + "\n"
}

func (s statsPanel) calcWidths(stats []teamStats) (int, int, int, int) {
	totalWidth := s.viewport.Width - 6 //nolint:mnd // 列之间的pandding

	valueWidth := 0
	itemWidth := 0

	for _, v := range stats {
		t := ansi.StringWidth(v.Text)
		if t > itemWidth {
			itemWidth = t
		}
		l := ansi.StringWidth(v.LeftVal)
		if l > valueWidth {
			valueWidth = l
		}
		r := ansi.StringWidth(v.RightVal)
		if r > valueWidth {
			valueWidth = r
		}
	}

	progressBarWidth := (totalWidth - valueWidth*2 - itemWidth) / 2 //nolint:mnd // 进度条宽度

	return totalWidth, valueWidth, itemWidth, progressBarWidth
}

func (s statsPanel) playerView() string {
	stats := s.msg.stats.playerStats
	if len(stats) == 0 {
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

	if len(tables) == 0 {
		return ""
	}

	if len(tables) >= 2 { //nolint:mnd // 每只球队一个table
		tables[0].Columns()[0].Title = s.msg.stats.team.LeftName
		tables[1].Columns()[0].Title = s.msg.stats.team.RightName
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
