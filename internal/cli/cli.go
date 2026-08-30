// Package cli wires opener's cobra command to the config, resolution, and
// launch pipeline.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inchestnov/opener/internal/config"
	"github.com/inchestnov/opener/internal/diagnostic"
	"github.com/inchestnov/opener/internal/opener"
	"github.com/inchestnov/opener/internal/source"
)

const version = "0.2.0"

// Options captures user-facing execution flags.
type Options struct {
	Verbose bool
}

// NewRootCmd builds opener's root cobra command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "opener <alias> <target>...",
		Short:   "opener is a macOS CLI wrapper around the native `open` mechanism, driven by aliases and configuration.",
		Version: version,
		Args:    requireAliasAndTarget,
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			return run(args, Options{Verbose: verbose})
		},
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeArg,
	}

	// Register --version without a shorthand ourselves, since cobra's
	// automatic version flag would otherwise bind -v, colliding with
	// --verbose.
	cmd.Flags().Bool("version", false, "version for opener")
	cmd.Flags().BoolP("verbose", "v", false, "print the resolution decision trail to stderr")

	return cmd
}

// requireAliasAndTarget enforces the `opener <alias> <target>...` form and
// points a user with old muscle memory at the replacement.
func requireAliasAndTarget(_ *cobra.Command, args []string) error {
	switch {
	case len(args) == 0:
		return fmt.Errorf("usage: opener <alias> <target>...")
	case len(args) == 1:
		return fmt.Errorf("opener needs an alias and a target: opener <alias> <target>...\n" +
			"  a bare `opener <target>` is no longer supported - define an alias (e.g. `open`) in ~/.opener.yaml")
	}
	return nil
}

// completeArg powers shell completion.
//
// The first positional argument completes to alias names. Once an alias is
// in place ("opener <alias> <target>..."), its targets are completed from
// that alias's `source`, if it has one; an alias with no source falls back
// to plain file completion.
func completeArg(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg := completionConfig()

	if len(args) == 0 {
		var names []string
		for name, a := range cfg.Aliases {
			if strings.HasPrefix(name, toComplete) {
				names = append(names, name+"\t"+aliasDescription(a))
			}
		}
		sort.Strings(names)
		return names, cobra.ShellCompDirectiveNoFileComp
	}

	a, ok := cfg.Aliases[args[0]]
	if !ok {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	if a.Source.IsZero() {
		return nil, cobra.ShellCompDirectiveDefault
	}

	src, err := source.New(a.Source, cfg.Sources)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	cands, err := src.Candidates(toComplete)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return cands, cobra.ShellCompDirectiveNoFileComp
}

// aliasDescription is the short completion hint shown next to an alias name.
func aliasDescription(a config.Alias) string {
	if a.Cmd != "" {
		return "cmd: " + a.Cmd
	}
	return "app: " + a.App
}

// completionConfig loads the config for completion, treating any problem as
// an empty config so completion degrades to plain file paths.
func completionConfig() *config.Config {
	path, err := defaultConfigPath()
	if err != nil {
		return &config.Config{}
	}
	cfg, err := config.LoadConfig(path)
	if err != nil {
		return &config.Config{}
	}
	return cfg
}

// run interprets args as an alias followed by its target(s).
func run(args []string, opts Options) error {
	alias, targets := args[0], args[1:]

	logger := diagnostic.Noop
	if opts.Verbose {
		logger = diagnostic.NewWriterLogger(os.Stderr)
	}

	configPath, err := defaultConfigPath()
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config %s: %w", configPath, err)
	}

	action, err := opener.Resolve(alias, targets, cfg, logger)
	if err != nil {
		return err
	}

	return opener.Launch(action, logger)
}

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}
	return filepath.Join(home, ".opener.yaml"), nil
}
