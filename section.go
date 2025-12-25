package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type sectionType int

type sectionChangedMsg struct {
	section sectionType
}

const (
	sectionURLPrompt sectionType = iota
	sectionDatatable
)

var sections = []sectionType{sectionURLPrompt, sectionDatatable} // nolint: gochecknoglobals

func (s sectionType) prev() sectionType {
	prevIdx := (int(s) - 1 + len(sections)) % len(sections)
	return sections[prevIdx]
}

func (s sectionType) next() sectionType {
	nextIdx := (int(s) + 1) % len(sections)
	return sections[nextIdx]
}

func sectionChangedCmd(s sectionType) tea.Cmd {
	return func() tea.Msg {
		return sectionChangedMsg{section: s}
	}
}
