// Package cli wires opener's cobra command to the config, resolution, and
// launch pipeline.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/inchestnov/opener/internal/config"
	"github.com/inchestnov/opener/internal/opener"
)

const version = "0.1.0"

// NewRootCmd builds opener's root cobra command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "opener <target> | opener <alias> <target>...",
		Short:   "opener is a macOS CLI wrapper around the native `open` mechanism, extended with aliases and configuration.",
		Version: version,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(args)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Register --version without a shorthand ourselves, since cobra's
	// automatic version flag would otherwise bind -v, colliding with
	// --verbose.
	cmd.Flags().Bool("version", false, "version for opener")

	return cmd
}

// run interprets args per REQ.md's argument-count rule: a single argument
// is a target for automatic mode; two or more are an alias followed by its
// target(s).
func run(args []string) error {
	var alias string
	targets := args
	if len(args) >= 2 {
		alias, targets = args[0], args[1:]
	}

	configPath, err := defaultConfigPath()
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config %s: %w", configPath, err)
	}

	action, err := opener.Resolve(alias, targets, cfg)
	if err != nil {
		return err
	}

	return opener.Launch(action)
}

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}
	return filepath.Join(home, ".opener.yaml"), nil
}
