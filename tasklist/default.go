package tasklist

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type DefaultItem interface {
	Item
	Content() string
}

type DefaultItemStyles struct {
	NormalContent    lipgloss.Style
	CompletedContent lipgloss.Style
	SelectedContent  lipgloss.Style
	DimmedContent    lipgloss.Style
}

func NewDefaultItemStyles() DefaultItemStyles {
	s := DefaultItemStyles{
		NormalContent: lipgloss.
			NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
			Padding(0, 0, 0, 2),
		CompletedContent: lipgloss.
			NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
			Padding(0, 0, 0, 2),
		SelectedContent: lipgloss.
			NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
			Padding(0, 0, 0, 2),
		DimmedContent: lipgloss.
			NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
			Padding(0, 0, 0, 2),
	}
	return s
}

type DefaultTaskDelegate struct {
	Styles        DefaultItemStyles
	UpdateFunc    func(tea.Msg, *Model) tea.Cmd
	ShortHelpFunc func() []key.Binding
	FullHelpFunc  func() [][]key.Binding
	spacing       int
}

func NewDefaultDelegate() DefaultTaskDelegate {
	return DefaultTaskDelegate{
		Styles:  NewDefaultItemStyles(),
		spacing: 1,
	}
}

func (d DefaultTaskDelegate) Height(m Model, index int, item Item) int {
	var content string
	if i, ok := item.(DefaultItem); ok {
		content = i.Content()
	} else {
		return 1
	}
	style := d.getStyleForItem(m, index, item)
	content = wordwrap.String(content, m.Width())
	h := lipgloss.Height(style.MaxWidth(m.Width()).Render(content))
	h += d.spacing + 1
	return h
}

func (d *DefaultTaskDelegate) SetSpacing(i int) {
	d.spacing = i
}

func (d DefaultTaskDelegate) Spacing() int {
	return d.spacing
}

func (d DefaultTaskDelegate) Update(msg tea.Msg, m *Model) tea.Cmd {
	if d.UpdateFunc == nil {
		return nil
	}
	return d.UpdateFunc(msg, m)
}

func (d DefaultTaskDelegate) getStyleForItem(m Model, index int, item Item) lipgloss.Style {
	var (
		s          = &d.Styles
		isSelected = index == m.Index()
	)

	var style lipgloss.Style
	if isSelected {
		style = s.SelectedContent
	} else if !item.IsChecked() {
		style = s.NormalContent
	} else {
		style = s.DimmedContent
	}
	return style
}

func (d DefaultTaskDelegate) Render(w io.Writer, m Model, index int, item Item) {
	var (
		content string
	)

	if i, ok := item.(DefaultItem); ok {
		content = i.Content()
	} else {
		return
	}

	style := d.getStyleForItem(m, index, item)
	content = wordwrap.String(content, m.Width())
	s := style.Width(m.width).Render(content)
	fmt.Fprintf(w, "%s", s)
}

func (d DefaultTaskDelegate) ShortHelp() []key.Binding {
	if d.ShortHelpFunc != nil {
		return d.ShortHelpFunc()
	}

	return nil
}

func (d DefaultTaskDelegate) FullHelp() [][]key.Binding {
	if d.FullHelpFunc != nil {
		return d.FullHelpFunc()
	}
	return nil
}
