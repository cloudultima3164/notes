package tasklist

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item interface {
	IsChecked() bool
}

type TaskDelegate interface {
	Render(w io.Writer, m Model, index int, item Item)
	Height(m Model, index int, item Item) int
	Spacing() int
	Update(msg tea.Msg, m *Model) tea.Cmd
}

type Model struct {
	showTitle      bool
	showHelp       bool
	filterText     bool
	filterComplete bool

	Title  string
	Styles Styles

	KeyMap KeyMap

	disableQuitKeybindings  bool
	AdditionalShortHelpKeys func() []key.Binding
	AdditionalFullHelpKeys  func() []key.Binding

	width  int
	height int
	cursor int
	Help   help.Model

	items       []Item
	itemHeights []int

	delegate TaskDelegate
}

func New(items []Item, delegate TaskDelegate, width, height int) Model {
	//styles := DefaultStyles()

	itemHeights := make([]int, len(items))
	m := Model{
		showTitle:      true,
		showHelp:       true,
		filterText:     false,
		filterComplete: false,

		Styles: DefaultStyles(),

		KeyMap: DefaultKeyMap(),
		Title:  "Task List",

		delegate:    delegate,
		items:       items,
		itemHeights: itemHeights,
		height:      height,
		width:       width,
		Help:        help.NewModel(),
	}

	m.updateKeybindings()
	m.regenerateItemHeights()

	return m
}

func (m *Model) SetShowTitle(v bool) {
	m.showTitle = v
}

func (m Model) ShowTitle() bool {
	return m.showTitle
}

func (m *Model) SetShowHelp(v bool) {
	m.showHelp = v
}

func (m Model) ShowHelp() bool {
	return m.showHelp
}

func (m Model) Items() []Item {
	return m.items
}

func (m *Model) InsertItem(index int, item Item) tea.Cmd {
	var cmd tea.Cmd
	m.items = insertItemIntoSlice(m.items, item, index)

	// TODO filtering

	m.updateKeybindings()
	m.regenerateItemHeights()

	return cmd
}

func (m *Model) RemoveItem(index int) {
	m.items = removeItemFromSlice(m.items, index)
	m.itemHeights = removeIntsFromSlice(m.itemHeights, index)

	// TODO filtering

}

func (m *Model) SetItem(index int, item Item) tea.Cmd {
	var cmd tea.Cmd
	m.items[index] = item

	// TODO: filtering?

	return cmd
}

func (m *Model) SetItems(i []Item) tea.Cmd {
	var cmd tea.Cmd
	m.items = i

	// TODO: filtering?
	//if m.filterComplete || m.filterText {
	//	m.filteredItems = nil
	//	cmd = filterItems(*m)
	//}

	m.updateKeybindings()
	m.regenerateItemHeights()
	return cmd
}

func (m Model) Index() int {
	return m.cursor
}

func (m *Model) Select(index int) {
	m.cursor = index
}

func (m *Model) ResetSelected() {
	m.Select(0)
}

func (m *Model) SetDelegate(d TaskDelegate) {
	m.delegate = d
}

func (m Model) Cursor() int {
	return m.cursor
}

func (m Model) SelectedItem() Item {
	i := m.Cursor()
	if i < 0 || len(m.items) == 0 || len(m.items) <= i {
		return nil
	}
	return m.items[i]
}

func (m *Model) CursorUp() {
	m.cursor--
	if m.cursor < 0 {
		m.cursor = 0
		return
	}

	if m.cursor >= 0 {
		return
	}
	// update visible items?
}

func (m *Model) CursorDown() {
	m.cursor++
	if m.cursor < len(m.items) {
		return
	}

	// update visible items?

	m.cursor = len(m.items) - 1
}

func (m Model) Width() int {
	return m.width
}

func (m Model) Height() int {
	return m.height
}

func (m *Model) DisableQuitKeybindings() {
	m.disableQuitKeybindings = true
	m.KeyMap.Quit.SetEnabled(false)
	m.KeyMap.ForceQuit.SetEnabled(false)
}

func (m *Model) SetSize(width, height int) {
	m.setSize(width, height)
}

func (m *Model) SetWidth(v int) {
	m.setSize(v, m.height)
}

func (m *Model) SetHeight(v int) {
	m.setSize(m.width, v)
}

func (m *Model) regenerateItemHeights() {
	for i, item := range m.items {
		h := m.delegate.Height(*m, i, item)
		m.itemHeights[i] = h
	}
}

func (m Model) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.KeyMap.CursorUp,
		m.KeyMap.CursorDown,
	}

	if m.AdditionalShortHelpKeys != nil {
		kb = append(kb, m.AdditionalShortHelpKeys()...)
	}

	kb = append(kb,
		m.KeyMap.Quit,
		m.KeyMap.CloseFullHelp,
	)

	return kb
}

func (m Model) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{{
		m.KeyMap.CursorUp,
		m.KeyMap.CursorDown,
		m.KeyMap.GoToTop,
		m.KeyMap.GoToBottom,
	}}

	listLevelBindings := []key.Binding{}

	if m.AdditionalFullHelpKeys != nil {
		listLevelBindings = append(listLevelBindings, m.AdditionalFullHelpKeys()...)
	}

	return append(kb, listLevelBindings, []key.Binding{
		m.KeyMap.Quit,
		m.KeyMap.CloseFullHelp,
	})
}

func (m *Model) setSize(width, height int) {
	widthUpdated := m.width != width
	m.width = width
	m.height = height
	if widthUpdated {
		m.Help.Width = width
		m.regenerateItemHeights()
	}
}

func (m *Model) updateKeybindings() {
	hasItems := len(m.items) != 0
	m.KeyMap.CursorUp.SetEnabled(hasItems)
	m.KeyMap.CursorDown.SetEnabled(hasItems)

	m.KeyMap.GoToBottom.SetEnabled(hasItems)
	m.KeyMap.GoToTop.SetEnabled(hasItems)

	m.KeyMap.Quit.SetEnabled(!m.disableQuitKeybindings)

	if m.Help.ShowAll {
		m.KeyMap.ShowFullHelp.SetEnabled(true)
		m.KeyMap.CloseFullHelp.SetEnabled(true)
	} else {
		minHelp := countEnabledBindings(m.FullHelp()) > 1
		m.KeyMap.ShowFullHelp.SetEnabled(minHelp)
		m.KeyMap.CloseFullHelp.SetEnabled(minHelp)
	}
}

func countEnabledBindings(groups [][]key.Binding) (agg int) {
	for _, group := range groups {
		for _, kb := range group {
			if kb.Enabled() {
				agg++
			}
		}
	}
	return agg
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.KeyMap.ForceQuit) {
			return m, tea.Quit
		}
	}
	cmds = append(cmds, m.handleBrowsing(msg))

	return m, tea.Batch(cmds...)
}

func (m *Model) handleBrowsing(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	numItems := len(m.items)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Quit):
			return tea.Quit
		case key.Matches(msg, m.KeyMap.CursorUp):
			m.CursorUp()
		case key.Matches(msg, m.KeyMap.CursorDown):
			m.CursorDown()
		case key.Matches(msg, m.KeyMap.GoToTop):
			m.cursor = 0
		case key.Matches(msg, m.KeyMap.GoToBottom):
			m.cursor = numItems - 1
		case key.Matches(msg, m.KeyMap.ShowFullHelp, m.KeyMap.CloseFullHelp):
			m.Help.ShowAll = !m.Help.ShowAll
		}
	}

	cmd := m.delegate.Update(msg, m)
	cmds = append(cmds, cmd)

	if m.cursor > numItems-1 {
		m.cursor = len(m.items) - 1
	}
	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var (
		sections    []string
		availHeight = m.height
	)

	if m.showTitle {
		v := m.titleView()
		sections = append(sections, v)
		availHeight -= lipgloss.Height(v)
	}

	var help string
	if m.showHelp {
		help = m.helpView()
		availHeight -= lipgloss.Height(help)
	}

	content := lipgloss.NewStyle().Height(availHeight).Render(m.populatedView(availHeight))

	sections = append(sections, content)

	if m.showHelp {
		sections = append(sections, help)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) populatedView(availHeight int) string {
	items, min, _ := m.VisibleItems(availHeight)
	var b strings.Builder

	if len(items) == 0 {
		return m.Styles.NoItems.Render("No items found.")
	}

	if len(items) > 0 {
		for i, item := range items {
			m.delegate.Render(&b, m, i+min, item)
			if i != len(items)-1 {
				fmt.Fprint(&b, strings.Repeat("\n", m.delegate.Spacing()+1))
			}
		}
	}

	return b.String()
}

func (m Model) VisibleItems(availHeight int) ([]Item, int, int) {
	if len(m.items) == 0 {
		return []Item{}, 0, 0
	}
	// TODO: filter items

	accHeight := 0

	accHeight += m.itemHeights[m.cursor]
	if accHeight > availHeight {
		return []Item{m.items[m.cursor]}, m.cursor, m.cursor
	}
	minIndex := m.cursor
	maxIndex := m.cursor

	noMoreBelow := false
	noMoreAbove := false
	size := 1
	for accHeight <= availHeight {
		if minIndex-1 >= 0 && accHeight+m.itemHeights[minIndex-1] < availHeight {
			minIndex -= 1
			size++
			accHeight += m.itemHeights[minIndex]
		} else {
			noMoreBelow = true
		}
		if maxIndex+1 < len(m.items) && accHeight+m.itemHeights[maxIndex+1] < availHeight {
			maxIndex += 1
			size++
			accHeight += m.itemHeights[maxIndex]
		} else {
			noMoreAbove = true
		}
		if noMoreAbove && noMoreBelow {
			break
		}
	}

	return m.items[minIndex : maxIndex+1], minIndex, maxIndex
}

func (m Model) titleView() string {
	var (
		view          string
		titleBarStyle = m.Styles.TitleBar.Copy()
	)

	if m.showTitle {
		view += m.Styles.Title.Render(m.Title)
	}

	return titleBarStyle.Render(view)
}

func (m Model) helpView() string {
	return m.Styles.HelpStyle.Render(m.Help.View(m))
}

// func (m *Model) ClearFilter()

func insertItemIntoSlice(items []Item, item Item, index int) []Item {
	if items == nil {
		return []Item{item}
	}
	if index >= len(items) {
		return append(items, item)
	}

	if index < 0 {
		index = 0
	}
	items = append(items, nil)
	copy(items[index+1:], items[index:])
	items[index] = item
	return items
}

func removeItemFromSlice(i []Item, index int) []Item {
	if index >= len(i) {
		return i // no op
	}

	copy(i[index:], i[index+1:])
	i[len(i)-1] = nil
	return i[:len(i)-1]
}

func removeIntsFromSlice(i []int, index int) []int {
	if index >= len(i) {
		return i // no op
	}
	return append(i[:index], i[index+1:]...)
}
