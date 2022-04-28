package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type taskListKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
}

type taskDelegateKeyMap struct {
	choose key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d taskDelegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d taskDelegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
		},
	}
}

type taskDelegate struct{}

func (d taskDelegate) Height() int                               { return 1 }
func (d taskDelegate) Spacing() int                              { return 0 }
func (d taskDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d taskDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Note)
	if !ok {
		return
	}

	var tags string
	if len(i.Tags) > 0 {
		tags = fmt.Sprintf(": %s", strings.Join(i.Tags, ", "))
	}
	str := fmt.Sprintf("%d. %s %s", index+1, i.Title, tags)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprint(w, fn(str))
}

func newTaskDelegateKeyMap() *taskDelegateKeyMap {
	return &taskDelegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Select"),
		),
	}
}

func newTaskListKeyMap() *taskListKeyMap {
	return &taskListKeyMap{
		toggleSpinner: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle spinner"),
		),
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

//var (
//appStyle = lipgloss.NewStyle().Padding(1, 2)

//itemStyle  = lipgloss.NewStyle().PaddingLeft(4)
//titleStyle = lipgloss.NewStyle().
//		Foreground(lipgloss.Color("#FFFDF5")).
//		Background(lipgloss.Color("#25A065")).
//		Padding(0, 1)
//selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
//)

//type taskListKeyMap struct {
//	toggleSpinner    key.Binding
//	toggleTitleBar   key.Binding
//	toggleStatusBar  key.Binding
//	togglePagination key.Binding
//	toggleHelpMenu   key.Binding
//}
//
//type taskDelegateKeyMap struct {
//	choose key.Binding
//}
//
//// Additional short help entries. This satisfies the help.KeyMap interface and
//// is entirely optional.
//func (d taskDelegateKeyMap) ShortHelp() []key.Binding {
//	return []key.Binding{
//		d.choose,
//	}
//}
//
//// Additional full help entries. This satisfies the help.KeyMap interface and
//// is entirely optional.
//func (d taskDelegateKeyMap) taskFullHelp() [][]key.Binding {
//	return [][]key.Binding{
//		{
//			d.choose,
//		},
//	}
//}
//
//type taskDelegate struct{}
//
//func (d taskDelegate) Height() int                               { return 1 }
//func (d taskDelegate) Spacing() int                              { return 0 }
//func (d taskDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
//func (d taskDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
//	i, ok := listItem.(Note)
//	if !ok {
//		return
//	}
//
//	var tags string
//	if len(i.Tags) > 0 {
//		tags = fmt.Sprintf(": %s", strings.Join(i.Tags, ", "))
//	}
//	str := fmt.Sprintf("%d. %s %s", index+1, i.Title, tags)
//
//	fn := itemStyle.Render
//	if index == m.Index() {
//		fn = func(s string) string {
//			return selectedItemStyle.Render("> " + s)
//		}
//	}
//
//	fmt.Fprint(w, fn(str))
//}
//
//func newDelegateKeyMap() *taskDelegateKeyMap {
//	return &taskDelegateKeyMap{
//		choose: key.NewBinding(
//			key.WithKeys("enter"),
//			key.WithHelp("enter", "Select"),
//		),
//	}
//}
//
//func newListKeyMap() *taskListKeyMap {
//	return &taskListKeyMap{
//		toggleSpinner: key.NewBinding(
//			key.WithKeys("s"),
//			key.WithHelp("s", "toggle spinner"),
//		),
//		toggleTitleBar: key.NewBinding(
//			key.WithKeys("T"),
//			key.WithHelp("T", "toggle title"),
//		),
//		toggleStatusBar: key.NewBinding(
//			key.WithKeys("S"),
//			key.WithHelp("S", "toggle status"),
//		),
//		togglePagination: key.NewBinding(
//			key.WithKeys("P"),
//			key.WithHelp("P", "toggle pagination"),
//		),
//		toggleHelpMenu: key.NewBinding(
//			key.WithKeys("H"),
//			key.WithHelp("H", "toggle help"),
//		),
//	}
//}

type taskModel struct {
	list         list.Model
	keys         *taskListKeyMap
	delegateKeys *taskDelegateKeyMap
	note         Note
}

func NewTaskViewer(n Note) (taskModel, error) {
	var (
		delegateKeys = newTaskDelegateKeyMap()
		listKeys     = newTaskListKeyMap()
	)

	// Make initial list of items

	items := make([]list.Item, 0)
	taskList := list.New(items, taskDelegate{}, 0, 0)
	taskList.Title = fmt.Sprintf("%v: Tasks", n.Title)
	taskList.Styles.Title = titleStyle
	taskList.AdditionalShortHelpKeys = delegateKeys.ShortHelp
	taskList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			delegateKeys.choose,
			listKeys.toggleHelpMenu,
			listKeys.toggleSpinner,
			listKeys.toggleTitleBar,
			listKeys.toggleStatusBar,
			listKeys.togglePagination,
		}
	}

	return taskModel{
		list:         taskList,
		keys:         listKeys,
		delegateKeys: delegateKeys,
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
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.delegateKeys.choose):
			//i, ok := m.list.SelectedItem().(Note)
			//if ok {
			//	m.choice = i
			//}
			return m, tea.Quit
		case key.Matches(msg, m.keys.toggleSpinner):
			cmd := m.list.ToggleSpinner()
			return m, cmd

		case key.Matches(msg, m.keys.toggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.keys.toggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.keys.togglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil
		}
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
