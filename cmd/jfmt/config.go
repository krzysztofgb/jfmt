package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type config struct {
	Template  string `toml:"template"`
	Indent    string `toml:"indent,omitempty"`
	Compact   bool   `toml:"compact"`
	Spec      string `toml:"spec"`
	SortKeys  bool   `toml:"sort_keys"`
	Verbose   bool   `toml:"verbose"`
	NoFix     bool   `toml:"no_fix"`
	Color     bool   `toml:"color"`
	NoColor   bool   `toml:"no_color"`
	JSONLines bool   `toml:"jsonlines"`
}

// writeTo encodes the config as TOML and writes it to w.
func (c config) writeTo(w io.Writer) error {
	return toml.NewEncoder(w).Encode(c)
}

func configPath() string {
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, "jfmt", "config.toml")
	}

	if runtime.GOOS == "windows" {
		if base := os.Getenv("APPDATA"); base != "" {
			return filepath.Join(base, "jfmt", "config.toml")
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".config", "jfmt", "config.toml")
}

// applyConfig merges non-zero file config values into cfg, skipping any field
// whose corresponding CLI flag was explicitly set.
func applyConfig(cmd *cobra.Command, fileCfg config, cfg *config) {
	if !cmd.Flags().Changed("template") && fileCfg.Template != "" {
		cfg.Template = fileCfg.Template
	}

	if !cmd.Flags().Changed("indent") && fileCfg.Indent != "" {
		cfg.Indent = fileCfg.Indent
	}

	if !cmd.Flags().Changed("spec") && fileCfg.Spec != "" {
		cfg.Spec = fileCfg.Spec
	}

	if !cmd.Flags().Changed("compact") && fileCfg.Compact {
		cfg.Compact = true
	}

	if !cmd.Flags().Changed("sort-keys") && fileCfg.SortKeys {
		cfg.SortKeys = true
	}

	if !cmd.Flags().Changed("verbose") && fileCfg.Verbose {
		cfg.Verbose = true
	}

	if !cmd.Flags().Changed("no-fix") && fileCfg.NoFix {
		cfg.NoFix = true
	}

	if !cmd.Flags().Changed("color") && fileCfg.Color {
		cfg.Color = true
	}

	if !cmd.Flags().Changed("no-color") && fileCfg.NoColor {
		cfg.NoColor = true
	}

	if !cmd.Flags().Changed("jsonlines") && fileCfg.JSONLines {
		cfg.JSONLines = true
	}
}

func loadConfig(explicitPath string) (config, error) {
	path := explicitPath
	if path == "" {
		path = configPath()
	}

	if path == "" {
		return config{}, nil //nolint:exhaustruct
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if explicitPath != "" {
			return config{}, fmt.Errorf("config file not found: %s", explicitPath) //nolint:exhaustruct
		}

		return config{}, nil //nolint:exhaustruct
	}

	if err != nil {
		return config{}, err //nolint:exhaustruct
	}

	var cfg config

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return config{}, err //nolint:exhaustruct
	}

	return cfg, nil
}
