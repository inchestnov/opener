# 01: Bootstrap + base open pipeline

**What to build:** Set up the Go project (module, structure, license) and get the thinnest end-to-end path working: a user runs `opener <target>` for a file, a directory, or a URL/nonexistent path, and it opens via the plain macOS `open` command, exactly as if the user had typed `open <target>` themselves. This is the foundation every later ticket extends.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] `go.mod` declares module `github.com/inchestnov/opener`, `go 1.26`; `LICENSE` (MIT) present
- [ ] Project structure matches `cmd/opener`, `internal/config`, `internal/opener`, `internal/cli`
- [ ] `opener .` and `opener ~/Downloads` open the directory in Finder
- [ ] `opener image.png` (or any file with no special rule) opens via system `open`
- [ ] `opener https://github.com` (or any nonexistent/non-file target) passes straight through to `open` without a resolution error
- [ ] `opener --help` and `opener --version` work
- [ ] `~/.opener.yaml` config is loaded via viper when present; when absent, resolution proceeds with an empty config (no error) and still falls back to system `open`
- [ ] No `sh -c`/shell involved in launching — commands run directly via `os/exec`
- [ ] Unit tests (stdlib `testing`, table-driven) cover target resolution (existing file / existing directory / nonexistent path) and config loading (missing file → empty config, present valid file → parsed struct)
