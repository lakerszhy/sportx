package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type textLivePanel struct {
	spinner spinner.Model
	matchID string
	msg     textLivesMsg
	width   int
	height  int
}

func newTextLivePanel(width int) textLivePanel {
	return textLivePanel{
		width: width,
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
		),
		msg: newTextLivesInitialMsg(),
	}
}

func (t textLivePanel) Init() tea.Cmd {
	return nil
}

func (t textLivePanel) Update(msg tea.Msg) (textLivePanel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if t.msg.isLoading() {
			t.spinner, cmd = t.spinner.Update(msg)
			return t, cmd
		}
		return t, nil
	case matchSelectionMsg:
		t.matchID = string(msg)
		if t.matchID == "" {
			cmd = func() tea.Msg {
				return newTextLivesInitialMsg()
			}
			return t, cmd
		}

		cmd = func() tea.Msg {
			return newTextLivesLoadingMsg(t.matchID)
		}
		cmds = append(cmds, cmd)

		cmd = func() tea.Msg {
			hasData, err := fetchMatchHasTextLives(t.matchID)
			if err != nil {
				return newTextLivesFailedMsg(t.matchID, err)
			}
			if !hasData {
				return newTextLivesNoDataMsg(t.matchID)
			}
			textLives, err := fetchTextLives(t.matchID)
			if err != nil {
				return newTextLivesFailedMsg(t.matchID, err)
			}
			return newTextLivesLoadedMsg(t.matchID, textLives)
		}
		cmds = append(cmds, cmd)

		cmds = append(cmds, t.spinner.Tick)

		return t, tea.Batch(cmds...)
	case textLivesMsg:
		if t.matchID != msg.matchID {
			return t, nil
		}

		t.msg = msg

		if t.msg.hasData && (msg.isSuccess() || msg.isFailed()) {
			cmd = tea.Tick(cfg.textLiveRefreshInterval, func(time.Time) tea.Msg {
				textLives, err := fetchTextLives(msg.matchID)
				if err != nil {
					return newTextLivesFailedMsg(msg.matchID, err)
				}
				return newTextLivesLoadedMsg(msg.matchID, textLives)
			})
		}
		return t, cmd
	}

	return t, nil
}

func (t textLivePanel) View() string {
	style := lipgloss.NewStyle().
		Height(t.height).
		MaxHeight(t.height).
		Width(t.width).
		Padding(1, 1)

	if t.msg.isInitial() {
		return style.AlignHorizontal(lipgloss.Center).Render("")
	}

	if t.msg.isLoading() {
		return style.AlignHorizontal(lipgloss.Center).
			Render(t.spinner.View() + "加载中...")
	}

	if t.msg.isFailed() {
		return style.AlignHorizontal(lipgloss.Center).
			Render("加载失败: " + t.msg.err.Error())
	}

	if t.msg.isSuccess() && len(t.msg.textLives) == 0 {
		return style.AlignHorizontal(lipgloss.Center).
			Render("暂无数据")
	}

	var b strings.Builder

	goal := fmt.Sprintf("%s - %s",
		t.msg.textLives[0].LeftGoal,
		t.msg.textLives[0].RightGoal,
	)
	if t.msg.textLives[0].Quarter != "" {
		goal = fmt.Sprintf("第%s节 %s", t.msg.textLives[0].Quarter, goal)
	}

	goalView := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Padding(0, 1).
		Render(goal)
	b.WriteString(goalView + "\n\n")

	for _, v := range t.msg.textLives {
		content := v.Content
		if v.Time != "" {
			content = fmt.Sprintf("%s %s", v.Time, v.Content)
		}
		if v.Plus != "" {
			plus := fmt.Sprintf("%s(%s-%s)", v.Plus, v.LeftGoal, v.RightGoal)
			plus = listFocusedStyle.Render(plus)
			content = fmt.Sprintf("%s %s", content, plus)
		}
		if strings.TrimSpace(content) != "" {
			b.WriteString(content + "\n")
		}
	}

	return style.Render(b.String())
}

func (t *textLivePanel) SetHeight(v int) {
	t.height = v
}
