package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type config struct {
	DownloadPath string `koanf:"download.path"`
	TempName     string `koanf:"download.temp_name"`
	tempDir      string
}

func loadConfig() (*config, error) {
	configfile, err := xdg.ConfigFile("ytqueue/config.toml")
	if err != nil {
		return nil, err
	}

	cfg := &config{}
	k := koanf.New(".")

	if err := k.Load(file.Provider(configfile), toml.Parser()); err != nil {
		return nil, err
	}

	if err := k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{FlatPaths: true}); err != nil {
		return nil, err
	}

	if cfg.DownloadPath == "" {
		cfg.DownloadPath = "~/Downloads"
	}

	var isHome bool

	if cfg.DownloadPath, isHome = strings.CutPrefix(cfg.DownloadPath, "~"); isHome {
		cfg.DownloadPath = filepath.Join(xdg.Home, cfg.DownloadPath)
	}

	if cfg.TempName == "" {
		cfg.TempName = "ytqueue_temp"
	}

	const filePerm = 0o744

	if err := os.MkdirAll(cfg.DownloadPath, os.ModeDir|filePerm); err != nil {
		return nil, err
	}

	cfg.tempDir, err = os.MkdirTemp(os.TempDir(), cfg.TempName)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func cleanupTempDir(tempDir string) {
	if err := os.RemoveAll(tempDir); err != nil {
		slog.Error("unable to remove temp dir", slog.String("error", err.Error()))
	}
}
