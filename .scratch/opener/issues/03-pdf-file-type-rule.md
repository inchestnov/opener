# 03: PDF file-type rule + config override

**What to build:** In automatic mode, detect a target's file extension and, for `.pdf`, check `open.files.pdf` in config for an override (App or Command), reusing the launch strategies built in ticket 02. Without an override, PDFs still fall back to plain `open` (ticket 01's behavior).

**Blocked by:** 02 (reuses the App/Command launch strategy)

**Status:** ready-for-agent

- [x] `opener document.pdf` with no `open.files.pdf` rule opens via plain `open document.pdf`
- [x] `opener document.pdf` with `open.files.pdf.app: "Google Chrome"` opens via `open -a "Google Chrome" document.pdf`
- [x] Changing the config to `open.files.pdf.app: "Safari"` correctly switches the resolved application
- [x] `open.files.pdf.command: "..."` form (CLI command override for PDFs) also resolves and launches directly, mirroring alias command-type behavior
- [x] Extension matching is case-insensitive (`.PDF` behaves the same as `.pdf`)
- [x] Files with no matching `open.files.<ext>` entry remain on the ticket 01 fallback path (no regression)
- [x] Unit tests cover file-type resolution (pdf vs. other extensions) and config rule resolution/override for `open.files.pdf`

**Note:** the implementation ended up extension-generic rather than pdf-hardcoded — `open.files.<ext>` is matched for any extension key present in config, not just `pdf` (a natural consequence of reusing `OpenConfig.Files map[string]Rule`, which already existed generically since ticket 01). Kept deliberately, per user decision, rather than artificially restricted to `pdf` only. Covered by `TestResolve_NonPDFExtensionRuleOverride`.
