# 05: Polish — README + error/exit-code audit

**What to build:** Close out the first version: document the tool end-to-end in the README, and make sure every error path (bad config, unknown alias, launch failure) produces a clear message and a nonzero exit code, not just the ones already covered by earlier tickets' acceptance criteria.

**Blocked by:** 04

**Status:** ready-for-agent

- [x] `README.md` documents install, usage (automatic mode, alias mode), the full config format (`aliases`, `open.directory`, `open.files.pdf`), and the worked examples from the spec (Finder, PDF, VS Code, Chrome, PDF-in-Chrome, Neovim)
- [x] Every error path (unknown alias, unreadable/malformed `~/.opener.yaml`, launch failure) prints an understandable message and exits via `os.Exit(1)` from `main.go`
- [x] Manual run-through of every Definition-of-Done scenario from the spec succeeds against a built binary (`go build`)
- [x] No hardcoded references to specific IDEs/file managers/browsers anywhere in the code — all such names come from config

**Note:** `internal/opener/resolver.go`'s directory branch logs `"using default application: Finder"` in verbose mode. This is a literal string from verbose.md's own worked example (a diagnostic label), not a behavioral choice — directories always launch via plain `open <target>`, letting macOS pick the app; no code branches on "Finder" specifically. Kept as an intentional exception rather than removed.
