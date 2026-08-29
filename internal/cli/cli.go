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

// completeArg powers shell completion.
//
// The first positional argument is usually a file, so a bare request
// (nothing typed) is left to the shell's file completion; once a prefix is
// typed, alias and template names from ~/.opener.yaml sharing it are
// offered, still falling back to files when none match.
//
// Once an alias is in place ("opener <alias> <target>..."), its targets are
// completed according to that alias's `complete:` setting.
func completeArg(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg := completionConfig()

	if len(args) >= 1 {
		if rule, ok := cfg.Aliases[args[0]]; ok {
			return completeAliasTarget(rule, toComplete)
		}
		return nil, cobra.ShellCompDirectiveDefault
	}

	if toComplete == "" {
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

// completeAliasTarget completes one target for an alias per its `complete:`
// setting.
func completeAliasTarget(rule config.Rule, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch rule.Complete {
	case "dirs":
		return nil, cobra.ShellCompDirectiveFilterDirs

	case "git-dirs":
		return completeGitDirs(toComplete)

	case "files":
		if exts := normalizeExtensions(rule.Extensions); len(exts) > 0 {
			return exts, cobra.ShellCompDirectiveFilterFileExt
		}
		return nil, cobra.ShellCompDirectiveDefault

	case "urls":
		var urls []string
		for _, u := range rule.URLs {
			if strings.HasPrefix(u, toComplete) {
				urls = append(urls, u)
			}
		}
		sort.Strings(urls)
		// When the prefix matches no URL, the shell falls back to files.
		return urls, cobra.ShellCompDirectiveDefault

	default: // "", "any"
		return nil, cobra.ShellCompDirectiveDefault
	}
}

// normalizeExtensions strips a leading dot from each extension, as cobra's
// file-extension filter expects bare extensions ("go", not ".go").
func normalizeExtensions(exts []string) []string {
	if len(exts) == 0 {
		return nil
	}
	out := make([]string, 0, len(exts))
	for _, e := range exts {
		if e = strings.TrimPrefix(strings.TrimSpace(e), "."); e != "" {
			out = append(out, e)
		}
	}
	return out
}

// completeGitDirs offers directories one path segment deep from what's been
// typed that directly contain a .git entry.
func completeGitDirs(toComplete string) ([]string, cobra.ShellCompDirective) {
	const stop = cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp

	typedDir, prefix := splitPath(toComplete)
	scanDir := expandUser(typedDir)
	if scanDir == "" {
		scanDir = "."
	}

	entries, err := os.ReadDir(scanDir)
	if err != nil {
		return nil, stop
	}

	var out []string
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		if _, err := os.Stat(filepath.Join(scanDir, e.Name(), ".git")); err != nil {
			continue
		}
		out = append(out, typedDir+e.Name()+"/")
	}
	sort.Strings(out)
	return out, stop
}

// splitPath splits s at its last slash into the directory portion (with the
// trailing slash kept, or empty) and the remaining filename prefix.
func splitPath(s string) (dir, prefix string) {
	if i := strings.LastIndexByte(s, '/'); i >= 0 {
		return s[:i+1], s[i+1:]
	}
	return "", s
}

// expandUser resolves a leading ~ or ~/ against the home directory.
func expandUser(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return home + strings.TrimPrefix(p, "~")
		}
	}
	return p
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
