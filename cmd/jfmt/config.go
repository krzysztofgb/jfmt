package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

type config struct {
	Template  string `toml:"template"`
	Indent    string `toml:"indent"`
	Compact   *bool  `toml:"compact"`
	Spec      string `toml:"spec"`
	SortKeys  *bool  `toml:"sort_keys"`
	Verbose   *bool  `toml:"verbose"`
	NoFix     *bool  `toml:"no_fix"`
	Color     *bool  `toml:"color"`
	NoColor   *bool  `toml:"no_color"`
	JSONLines *bool  `toml:"jsonlines"`
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

func applyConfig(cmd *cobra.Command, cfg config, template, indent, spec *string, compact, sortKeys, verbose, noFix, color, noColor, jsonLines *bool) {
	if !cmd.Flags().Changed("template") && cfg.Template != "" {
		*template = cfg.Template
	}

	if !cmd.Flags().Changed("indent") && cfg.Indent != "" {
		*indent = cfg.Indent
	}

	if !cmd.Flags().Changed("spec") && cfg.Spec != "" {
		*spec = cfg.Spec
	}

	if !cmd.Flags().Changed("compact") && cfg.Compact != nil {
		*compact = *cfg.Compact
	}

	if !cmd.Flags().Changed("sort-keys") && cfg.SortKeys != nil {
		*sortKeys = *cfg.SortKeys
	}

	if !cmd.Flags().Changed("verbose") && cfg.Verbose != nil {
		*verbose = *cfg.Verbose
	}

	if !cmd.Flags().Changed("no-fix") && cfg.NoFix != nil {
		*noFix = *cfg.NoFix
	}

	if !cmd.Flags().Changed("color") && cfg.Color != nil {
		*color = *cfg.Color
	}

	if !cmd.Flags().Changed("no-color") && cfg.NoColor != nil {
		*noColor = *cfg.NoColor
	}

	if !cmd.Flags().Changed("jsonlines") && cfg.JSONLines != nil {
		*jsonLines = *cfg.JSONLines
	}
}

func loadConfig() (config, error) {
	path := configPath()
	if path == "" {
		return config{}, nil //nolint:exhaustruct
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
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
