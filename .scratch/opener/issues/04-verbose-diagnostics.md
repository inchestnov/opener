# 04: Verbose diagnostics

**What to build:** A `-v`/`--verbose` flag, position-independent relative to the target/alias, that prints the full resolution decision trail to stderr without changing actual behavior. Diagnostics flow through a `Logger` interface (`Debug(format string, args ...any)`) injected via an `Options{Verbose bool}` value — no direct `fmt.Println` in resolver/launcher code.

**Blocked by:** 03 (needs every pipeline stage — target, type, file type, alias, config rule, fallback — present to log)

**Status:** ready-for-agent

- [ ] `-v`/`--verbose` work in any position: `opener -v document.pdf`, `opener document.pdf -v`, `opener --verbose document.pdf`
- [ ] Same works for directory and alias invocations: `opener -v .`, `opener -v ide .`, `opener editor . --verbose`
- [ ] Verbose trace for automatic/pdf mode matches verbose.md's sequence: target → target type → file type → config check → rule found/fallback → resolved application/command → launch command
- [ ] Verbose trace for directory mode shows "no custom directory rule found" / default Finder path when applicable
- [ ] Verbose trace for alias mode shows alias name, alias type (application/command), and resolved launch command
- [ ] All verbose output goes to stderr; stdout and program behavior are identical with and without `-v`
- [ ] Without `-v`, only essential errors are printed — no diagnostic noise
- [ ] `Logger` has a no-op implementation used when `Verbose == false`; resolver/launcher only ever call `Logger.Debug`, never print directly
