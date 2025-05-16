package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

type listPanel struct {
	list    list.Model
	spinner spinner.Model
}

func newListPanel(d list.ItemDelegate) listPanel {
	l := list.New([]list.Item{}, d, 0, 0)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowPagination(false)
	return listPanel{
		list:    l,
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
	}
}

func (p listPanel) render(focused bool, status status, err error) string {
	style := borderStyle
	if focused {
		style = borderFocusedStyle
	}
	style = style.Width(p.list.Width()).Height(p.list.Height())
	centerStyle := style.AlignHorizontal(lipgloss.Center)

	if status.isInitial() {
		return style.Render("")
	}

	if status.isLoading() {
		return centerStyle.Render(p.spinner.View() + " 加载中...")
	}

	if status.isFailed() {
		return centerStyle.Render("加载失败: " + err.Error())
	}

	if len(p.list.Items()) == 0 {
		return centerStyle.Render("暂无数据")
	}

	indicator := fmt.Sprintf("|%d/%d|", p.list.Index()+1, len(p.list.Items()))
	content := p.list.View()

	border := style.GetBorderStyle()
	bottom := strings.Repeat(border.Bottom, max(p.list.Width()-len(indicator)-1, 0))
	border.Bottom = bottom + indicator + border.Bottom
	return style.Border(border).Render(content)
}

func (p *listPanel) setSize(width int, height int) {
	p.list.SetSize(width, height)
}
