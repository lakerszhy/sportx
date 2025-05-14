package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type categoryPanel struct {
	msg     categoriesMsg
	list    list.Model
	spinner spinner.Model
	width   int
}

func newCategoryPanel(width int) categoryPanel {
	l := list.New([]list.Item{}, categoryDelegate{}, width, 0)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowPagination(false)
	return categoryPanel{
		msg:   newCategoriesLoadingMsg(),
		list:  l,
		width: width,
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
		),
	}
}

func (c categoryPanel) Init() tea.Cmd {
	return tea.Batch(
		c.spinner.Tick,
		func() tea.Msg {
			categories, err := fetchCategories()
			if err != nil {
				return newCategoriesFailedMsg(err)
			}
			return newCategoriesLoadedMsg(categories)
		},
	)
}

func (c categoryPanel) Update(msg tea.Msg) (categoryPanel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if c.msg.isLoading() {
			c.spinner, cmd = c.spinner.Update(msg)
			return c, cmd
		}
		return c, nil
	case categoriesMsg:
		c.msg = msg
		if msg.isSuccess() {
			if len(msg.categories) == 0 {
				return c, nil
			}

			cmd = func() tea.Msg {
				return categorySelectionMsg(msg.categories[0])
			}

			var items []list.Item
			for _, category := range msg.categories {
				items = append(items, category)
			}
			c.list.SetItems(items)
		}
		return c, cmd
	}

	before, ok1 := c.list.SelectedItem().(category)
	c.list, cmd = c.list.Update(msg)
	after, ok2 := c.list.SelectedItem().(category)
	if ok1 && ok2 && !before.equal(after) {
		cmd = func() tea.Msg {
			return categorySelectionMsg(after)
		}
	}

	return c, cmd
}

func (c categoryPanel) View() string {
	if c.msg.isLoading() {
		return lipgloss.NewStyle().Width(c.width).
			AlignHorizontal(lipgloss.Center).
			Render(c.spinner.View() + "加载中...")
	}

	if c.msg.isFailed() {
		return "加载失败: " + c.msg.err.Error()
	}

	if c.msg.isSuccess() && len(c.msg.categories) == 0 {
		return lipgloss.NewStyle().Width(c.width).
			AlignHorizontal(lipgloss.Center).
			Render("暂无数据")
	}

	return c.list.View()
}

func (c *categoryPanel) SetHeight(v int) {
	c.list.SetHeight(v)
}

type categoryDelegate struct{}

func (d categoryDelegate) Height() int {
	return 1
}

func (d categoryDelegate) Spacing() int {
	return 0
}

func (d categoryDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d categoryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(category)
	if !ok {
		return
	}

	content := fmt.Sprintf("  %s", i.Name)
	if index == m.Index() {
		content = fmt.Sprintf("> %s", i.Name)
		content = listFocusedStyle.Render(content)
	}

	content = ansi.Truncate(content, m.Width(), "...")
	fmt.Fprintf(w, "%s", content)
}
