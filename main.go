package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

const timeout = time.Second * 60

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type errMsg error

type model struct {
	textarea  textarea.Model
	textinput textinput.Model
	timer     timer.Model
	wpm       int
	message   string
	err       error
}

func initialModel() model {
	t := timer.NewWithInterval(timeout, time.Second)
	ta := textarea.New()
	ta.Placeholder = "Burst entry here..."
	ta.Focus()
	ti := textinput.New()

	return model{
		timer:     t,
		textarea:  ta,
		textinput: ti,
		wpm:       0,
		err:       nil,
		message:   "",
	}
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd
	case timer.TimeoutMsg:
		m.wpm = len(strings.Fields(m.textarea.Value())) * (int(time.Minute) / int(timeout))
		m.textinput.Focus()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.timer.Timedout() {
				if len(m.textinput.Value()) > 0 {
					os.WriteFile(m.textinput.Value(), []byte(m.textarea.Value()), 0644)
				}
				return m, tea.Quit
			}
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	if m.timer.Timedout() {
		m.textinput, cmd = m.textinput.Update(msg)
	} else {
		m.textarea, cmd = m.textarea.Update(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.timer.Timedout() {
		return fmt.Sprintf(
			"Time is up! Your wpm was %d\n\n%s\n\n%s %s",
			m.wpm,
			m.textarea.View(),
			"Enter filename:",
			m.textinput.View(),
		) + "\n\n"
	}
	return fmt.Sprintf(
		"You have %s to write your burst.\n\n%s\n\n%s",
		m.timer.View(),
		m.textarea.View(),
		"(ctrl+c to quit)",
	) + "\n\n"
}
