package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type taskAddOuter struct {
	mod    taskAddModel
	result string
	outCh  chan string
	err    error
}

func NewTaskAdd() taskAddOuter {
	outCh := make(chan string, 1)
	mod := newTaskAdder(outCh)
	return taskAddOuter{
		mod:    mod,
		result: "",
		outCh:  outCh,
		err:    nil,
	}
}

func (w *taskAddOuter) StartGetTaskDetails() error {
	err := tea.NewProgram(w.mod, tea.WithAltScreen()).Start()
	close(w.outCh)
	(*w).result = <-w.outCh
	return err
}

type taskAddModel struct {
	input textinput.Model
	outCh chan string
	err   error
}

func newTaskAdder(ch chan string) taskAddModel {
	in := textinput.New()
	in.Placeholder = "New Task"
	in.Focus()
	in.CharLimit = 256
	in.Width = 20

	return taskAddModel{
		input: in,
		outCh: ch,
		err:   nil,
	}
}

func (m taskAddModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m taskAddModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch message := msg.(type) {
	case tea.KeyMsg:
		switch message.Type {
		case tea.KeyEnter:
			m.outCh <- m.input.Value()
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case error:
		m.err = message
		return m, nil
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m taskAddModel) View() string {
	style := lipgloss.
		NewStyle().
		Align(lipgloss.Center)
	fmtString := fmt.Sprintf(
		"Add task details:\n\n%s\n",
		m.input.View(),
	)
	return style.Render(fmtString)
}
