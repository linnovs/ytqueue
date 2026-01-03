package main

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type playingNow struct {
	width         int
	player        *player
	style         lipgloss.Style
	headerStyle   lipgloss.Style
	filenameStyle lipgloss.Style
	progress      progress.Model
	getCtx        contextFn
}

func newPlayingNow(player *player, getCtx contextFn) *playingNow {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("48"))
	headerStyle := lipgloss.NewStyle().
		Bold(true).Padding(0, 1).
		Background(lipgloss.Color("48")).
		Foreground(lipgloss.Color("231"))
	filenameStyle := headerStyle.Background(lipgloss.Color("99"))
	pp := progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage())

	return &playingNow{
		player:        player,
		style:         style,
		headerStyle:   headerStyle,
		filenameStyle: filenameStyle,
		progress:      pp,
		getCtx:        getCtx,
	}
}

func (p *playingNow) Init() tea.Cmd {
	return nil
}

func (p *playingNow) Update(msg tea.Msg) (*playingNow, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width - p.style.GetHorizontalFrameSize()
	case progress.FrameMsg:
		model, cmd := p.progress.Update(msg)
		cmds = append(cmds, cmd)
		p.progress = model.(progress.Model)
	}

	return p, tea.Batch(cmds...)
}

func (p *playingNow) renderHeader() string {
	const header = "Playing Now"

	bar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		p.headerStyle.Render(header),
		p.filenameStyle.Render(p.player.getPlayingFilename()),
	)

	return lipgloss.NewStyle().
		Width(p.width).
		Render(bar)
}

func (p *playingNow) View() string {
	if p.player.getPlayingFilename() == "" {
		return ""
	}

	const maxComponents = 2

	components := make([]string, 0, maxComponents)
	components = append(components, p.renderHeader())
	components = append(components, p.player.renderPlayProgress(p.width))

	return p.style.Width(p.width).Render(lipgloss.JoinVertical(lipgloss.Top, components...))
}
