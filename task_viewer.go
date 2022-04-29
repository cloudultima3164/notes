package main

import (
	"fmt"

	"github.com/JamieCrisman/notes/tasklist"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type taskModel struct {
	list tasklist.Model
	// note         Note
}

func NewTaskViewer(n Note) (taskModel, error) {
	// Make initial list of items

	segments := Segmentize(n.Content)

	items := make([]tasklist.Item, len(segments))
	for in, it := range segments {
		items[in] = it
	}

	tl := tasklist.New(items, tasklist.NewDefaultDelegate(), 0, 0)
	tl.Title = fmt.Sprintf("%v: Tasks", n.Title)

	return taskModel{
		list: tl,
	}, nil
}

func (m taskModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m taskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		switch {

		case key.Matches(msg, m.list.KeyMap.SetInProgress):
			i := m.list.Cursor()
			item := m.list.Items()[i]
			if t, ok := item.(NoteSegment); ok {
				t.SetCheck("•")
				m.list.SetItem(i, t)
			}
		case key.Matches(msg, m.list.KeyMap.ClearStatus):
			i := m.list.Cursor()
			item := m.list.Items()[i]
			if t, ok := item.(NoteSegment); ok {
				t.SetCheck(" ")
				m.list.SetItem(i, t)
			}
		case key.Matches(msg, m.list.KeyMap.SetComplete):
			i := m.list.Cursor()
			item := m.list.Items()[i]
			if t, ok := item.(NoteSegment); ok {
				t.SetCheck("✓")
				m.list.SetItem(i, t)
			}
		}
		//	case tea.KeyMsg:
		//		// Don't match any of the keys below if we're actively filtering.
		//		if m.list.FilterState() == list.Filtering {
		//			break
		//		}
		//
		//		switch {
		//		case key.Matches(msg, m.delegateKeys.choose):
		//			//i, ok := m.list.SelectedItem().(Note)
		//			//if ok {
		//			//	m.choice = i
		//			//}
		//			return m, tea.Quit
		//		case key.Matches(msg, m.keys.toggleSpinner):
		//			cmd := m.list.ToggleSpinner()
		//			return m, cmd
		//
		//		case key.Matches(msg, m.keys.toggleTitleBar):
		//			v := !m.list.ShowTitle()
		//			m.list.SetShowTitle(v)
		//			m.list.SetShowFilter(v)
		//			m.list.SetFilteringEnabled(v)
		//			return m, nil
		//
		//		case key.Matches(msg, m.keys.toggleStatusBar):
		//			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
		//			return m, nil
		//
		//		case key.Matches(msg, m.keys.togglePagination):
		//			m.list.SetShowPagination(!m.list.ShowPagination())
		//			return m, nil
		//
		//		case key.Matches(msg, m.keys.toggleHelpMenu):
		//			m.list.SetShowHelp(!m.list.ShowHelp())
		//			return m, nil
		//		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m taskModel) View() string {
	return appStyle.Render(m.list.View())
}
