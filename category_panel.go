package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type categoryPanel struct {
	msg categoriesMsg
	listPanel
}

func newCategoryPanel() categoryPanel {
	return categoryPanel{
		msg:       newCategoriesLoadingMsg(),
		listPanel: newListPanel(categoryDelegate{}),
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

func (c categoryPanel) View(focused bool) string {
	return c.render(focused, c.msg.status, c.msg.err)
}

type categoryDelegate struct{}

func (d categoryDelegate) Height() int {
	return 1
}

func (d categoryDelegate) Spacing() int {
	return 0
}

func (d categoryDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
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
