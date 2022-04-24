package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	itemStyle  = lipgloss.NewStyle().PaddingLeft(4)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type listKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
	//insertItem       key.Binding
}

//func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
//	d := list.NewDefaultDelegate()
//
//	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
//		var title string
//
//		if i, ok := m.SelectedItem().(item); ok {
//			title = i.Title()
//		} else {
//			return nil
//		}
//
//		switch msg := msg.(type) {
//		case tea.KeyMsg:
//			switch {
//			case key.Matches(msg, keys.choose):
//				return m.NewStatusMessage(statusMessageStyle("You chose " + title))
//
//			case key.Matches(msg, keys.remove):
//				index := m.Index()
//				m.RemoveItem(index)
//				if len(m.Items()) == 0 {
//					keys.remove.SetEnabled(false)
//				}
//				return m.NewStatusMessage(statusMessageStyle("Deleted " + title))
//			}
//		}
//
//		return nil
//	}
//
//	help := []key.Binding{keys.choose, keys.remove}
//
//	d.ShortHelpFunc = func() []key.Binding {
//		return help
//	}
//
//	d.FullHelpFunc = func() [][]key.Binding {
//		return [][]key.Binding{help}
//	}
//
//	return d
//}

type delegateKeyMap struct {
	choose key.Binding
	remove key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
		d.remove,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
			d.remove,
		},
	}
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
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

	fmt.Fprintf(w, fn(str))
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
		remove: key.NewBinding(
			key.WithKeys("x", "backspace"),
			key.WithHelp("x", "delete"),
		),
	}
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		//		insertItem: key.NewBinding(
		//			key.WithKeys("a"),
		//			key.WithHelp("a", "add item"),
		//		),
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

type model struct {
	list list.Model
	//itemGenerator *randomItemGenerator
	keys         *listKeyMap
	delegateKeys *delegateKeyMap
	choice       Note
}

func newModel(title string, headerOnly bool) model {
	var (
		//		itemGenerator randomItemGenerator
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
	)

	// Make initial list of items
	notes, err := collectFiles(headerOnly)
	if err != nil {
		notes = []Note{}
		// TODO: uh explode or something
	}
	items := make([]list.Item, len(notes))
	for i, n := range notes {
		items[i] = n
	}
	//	for i := 0; i < numItems; i++ {
	//		items[i] = itemGenerator.next()
	//	}

	// Setup list
	//delegate := newItemDelegate(delegateKeys)
	fileList := list.New(items, itemDelegate{}, 0, 0)
	fileList.Title = title
	fileList.Styles.Title = titleStyle
	fileList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleSpinner,
			//listKeys.insertItem,
			listKeys.toggleTitleBar,
			listKeys.toggleStatusBar,
			listKeys.togglePagination,
			listKeys.toggleHelpMenu,
		}
	}

	return model{
		list:         fileList,
		keys:         listKeys,
		delegateKeys: delegateKeys,
		//		itemGenerator: &itemGenerator,
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			i, ok := m.list.SelectedItem().(Note)
			if ok {
				m.choice = i
			}
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

			//	case key.Matches(msg, m.keys.insertItem):
			//		m.delegateKeys.remove.SetEnabled(true)
			//		//newItem := m.itemGenerator.next()
			//		insCmd := m.list.InsertItem(0, nil)
			//		statusCmd := m.list.NewStatusMessage(statusMessageStyle("Added " /*newItem.Title()*/))
			//		return m, tea.Batch(insCmd, statusCmd)
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(m.list.View())
}

/*
if err := tea.NewProgram(newModel()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
*/
