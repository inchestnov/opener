# 02: Alias resolution mode

**What to build:** Support `opener <alias> <target> [target...]`, where the alias is looked up in config and launched either as a macOS GUI application (`open -a "App" target...`) or as a CLI command run directly (`exec.Command(command, targets...)`). An alias not present in config produces a clear error and a nonzero exit code.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] One positional arg → automatic mode (ticket 01's path); two or more args → alias mode, where the first arg is the alias name and the rest are targets
- [ ] `opener ide .` (with `aliases.ide.app: "Visual Studio Code"` configured) opens the directory in VS Code via `open -a "Visual Studio Code" .`
- [ ] `opener editor README.md` (with `aliases.editor.command: "nvim"`) runs `nvim README.md` directly, no shell
- [ ] Multiple targets after an alias are all passed through to the same app/command invocation, e.g. `opener editor a.md b.md`
- [ ] `opener foo .` for an alias not in config prints `unknown alias: foo` and exits nonzero
- [ ] Unit tests cover alias resolution: found+app-type, found+command-type, not-found, and multi-target passthrough — all against in-memory config, no process execution
