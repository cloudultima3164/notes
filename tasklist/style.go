package tasklist

import (
	"github.com/charmbracelet/lipgloss"
)

//const (
//	bullet   = "•"
//	ellipsis = "…"
//)

type Styles struct {
	TitleBar lipgloss.Style
	Title    lipgloss.Style

	NoItems lipgloss.Style

	HelpStyle lipgloss.Style
}

func DefaultStyles() (s Styles) {
	//verySubduedColor := lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"}
	//subduedColor := lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}

	s.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	s.Title = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)

	s.NoItems = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#909090", Dark: "#626262"})

	s.HelpStyle = lipgloss.NewStyle().Padding(1, 0, 0, 2)

	return s
}
