package main

import "syscall"

func (p *player) isPlaying() bool {
	p.playingMu.Lock()
	defer p.playingMu.Unlock()

	return p.playing
}

func (p *player) setPlaying(playing bool) {
	p.playingMu.Lock()
	defer p.playingMu.Unlock()
	p.playing = playing
}

func (p *player) isRunning() bool {
	p.processMu.Lock()
	defer p.processMu.Unlock()

	if p.process == nil {
		return false
	}

	if err := p.process.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	return true
}
