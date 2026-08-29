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
)

const version = "0.1.0"

// Options captures user-facing execution flags.
type Options struct {
	Verbose bool
}

// NewRootCmd builds opener's root cobra command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "opener <target> | opener <alias> <target>...",
		Short:   "opener is a macOS CLI wrapper around the native `open` mechanism, extended with aliases and configuration.",
		Version: version,
		Args:    cobra.MinimumNArgs(1),
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

// completeArg powers shell completion. The first positional argument is
// usually a file, so a bare completion request (nothing typed yet) is left
// to the shell's file completion. Once the user starts typing, alias and
// template names from ~/.opener.yaml that share that prefix are offered;
// when none match, the shell still falls back to files. Any later argument
// is a plain target -> file completion.
func completeArg(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 || toComplete == "" {
		return nil, cobra.ShellCompDirectiveDefault
	}

	configPath, err := defaultConfigPath()
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	var names []string
	for name := range cfg.Aliases {
		if strings.HasPrefix(name, toComplete) {
			names = append(names, name+"\talias")
		}
	}
	for name := range cfg.Templates {
		if strings.HasPrefix(name, toComplete) {
			names = append(names, name+"\ttemplate")
		}
	}
	sort.Strings(names)

	// ShellCompDirectiveDefault (not NoFileComp): if nothing here matches,
	// the shell still completes file paths.
	return names, cobra.ShellCompDirectiveDefault
}

// run interprets args per REQ.md's argument-count rule: a single argument
// is a target for automatic mode; two or more are an alias followed by its
// target(s).
func run(args []string, opts Options) error {
	var alias string
	targets := args
	if len(args) >= 2 {
		alias, targets = args[0], args[1:]
	}

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

	diag := diagnostic.Context{Logger: logger, ConfigPath: configPath}

	action, err := opener.Resolve(alias, targets, cfg, diag)
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
