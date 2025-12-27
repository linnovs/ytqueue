package main

import (
	"net/url"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
)

type submitURLMsg struct {
	url string
}

type urlPrompt struct {
	prompt    textinput.Model
	spinner   spinner.Model
	queueList []string
	style     lipgloss.Style
	keymap    promptKeymap
}

func newURLPrompt() *urlPrompt {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(activeBorderColor)
	i := textinput.New()
	i.Placeholder = "Enter URL here..."
	l := make([]string, 0)
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = s.Style.Foreground(lipgloss.Color("99"))

	return &urlPrompt{i, s, l, style, newPromptKeymap()}
}

func (p *urlPrompt) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, p.prompt.Focus(), p.spinner.Tick)
}

func submitURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		return submitURLMsg{url}
	}
}

func (p *urlPrompt) Update(msg tea.Msg) (*urlPrompt, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.prompt.Width = msg.Width - p.style.GetHorizontalFrameSize()
		p.prompt.Width -= lipgloss.Width(p.prompt.Prompt) + 1
	case sectionChangedMsg:
		if msg.section == sectionURLPrompt {
			p.style = p.style.BorderForeground(activeBorderColor)
			cmds = append(cmds, p.prompt.Focus())
		} else {
			p.style = p.style.UnsetBorderForeground()
			p.prompt.Blur()
		}
	case downloadQueuedMsg:
		p.queueList = append(p.queueList, msg.url)
	case downloadCompletedMsg:
		_, p.queueList = p.queueList[0], p.queueList[1:]
	case spinner.TickMsg:
		var cmd tea.Cmd
		p.spinner, cmd = p.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		if !p.prompt.Focused() {
			break
		}

		switch {
		case key.Matches(msg, p.keymap.clear):
			p.prompt.Reset()
		case key.Matches(msg, p.keymap.submit):
			filmUrl, err := url.Parse(p.prompt.Value())
			if err != nil {
				return p, errorCmd(err)
			}

			cmds = append(cmds, submitURLCmd(filmUrl.String()))

			p.prompt.Reset()
		}
	}

	var cmd tea.Cmd
	p.prompt, cmd = p.prompt.Update(msg)
	cmds = append(cmds, cmd)

	return p, tea.Batch(cmds...)
}

func (p *urlPrompt) queueSpinner(l list.Items, i int) string {
	if i == 0 {
		return p.spinner.View()
	}

	return ""
}

func queueListItemStyle(_ list.Items, index int) lipgloss.Style {
	if index == 0 {
		return lipgloss.NewStyle().Bold(true)
	}

	return lipgloss.NewStyle().Italic(true)
}

func (p *urlPrompt) renderQueueList() string {
	if len(p.queueList) == 0 {
		return ""
	}

	l := list.New().
		Enumerator(p.queueSpinner).
		Hide(len(p.queueList) == 0).
		ItemStyleFunc(queueListItemStyle)

	for _, url := range p.queueList {
		l.Item(url)
	}

	return l.String()
}

func (p *urlPrompt) View() string {
	content := p.prompt.View()

	queueList := p.renderQueueList()
	if queueList != "" {
		content = lipgloss.JoinVertical(lipgloss.Top, content, queueList)
	}

	return p.style.Render(content)
}
