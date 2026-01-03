package main

import (
	"syscall"
	"time"
)

func (p *player) isPlaying() bool {
	p.playingMu.Lock()
	defer p.playingMu.Unlock()

	return p.playing == playingStatusPlaying
}

func (p *player) getPlaying() playingStatus {
	p.playingMu.RLock()
	defer p.playingMu.RUnlock()

	return p.playing
}

func (p *player) setPlaying(playing playingStatus, id ...string) {
	p.playingMu.Lock()
	defer p.playingMu.Unlock()
	p.playing = playing

	if len(id) > 0 {
		p.currentlyPlayingId = id[0]
	}
}

func (p *player) setPlaytime(playtime time.Duration) {
	p.playtimeMu.Lock()
	defer p.playtimeMu.Unlock()

	p.playtime = playtime
}

func (p *player) getPlaytime() time.Duration {
	p.playtimeMu.RLock()
	defer p.playtimeMu.RUnlock()

	return p.playtime
}

func (p *player) setRemainingTime(remaining time.Duration) {
	p.playtimeMu.Lock()
	defer p.playtimeMu.Unlock()

	p.playtimeRemaining = remaining
}

func (p *player) getRemainingTime() time.Duration {
	p.playtimeMu.RLock()
	defer p.playtimeMu.RUnlock()

	return p.playtimeRemaining
}

func (p *player) setPlayingFilename(filename string) {
	p.playingMu.Lock()
	defer p.playingMu.Unlock()

	p.currentlyPlayingFilename = filename
}

func (p *player) getPlayingFilename() string {
	p.playingMu.RLock()
	defer p.playingMu.RUnlock()

	return p.currentlyPlayingFilename
}

func (p *player) getCurrentlyPlayingId() string {
	p.playingMu.RLock()
	defer p.playingMu.RUnlock()

	if p.playing == playingStatusPlaying {
		return p.currentlyPlayingId
	}

	return ""
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
