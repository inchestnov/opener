# opener

`opener` is a small macOS CLI that gives you one interface for opening files, directories, applications, and URLs — behind a layer of named aliases, each of which knows how to open its targets and how to tab-complete them.

## Installation

```bash
go install github.com/inchestnov/opener/cmd/opener@latest
```

This puts `opener` in `$(go env GOPATH)/bin` — make sure that's on your `PATH`.

<details>
<summary>Building locally instead</summary>

If you're working on `opener` itself, build from a checkout of this repo:

```bash
make build      # builds ./bin/opener
```

or directly with Go:

```bash
go build -o bin/opener ./cmd/opener
```

</details>

## Usage

Every invocation is an **alias** followed by one or more **targets**:

```bash
opener <alias> <target>...
```

The alias decides how the targets open. With the [example config](examples/opener.yaml):

```bash
opener open docs.pdf                # `open` — hands the file to the system default
opener code ~/src/opener           # `code` — opens the directory in VS Code
opener web https://github.com      # `web`  — opens the URL in Chrome
opener edit main.go notes.md       # `edit` — opens both files in nvim
```

There is no bare `opener <target>` form — define an `open` alias (`cmd: open`) for that.

> [!TIP]
> Typing `opener` in full gets old fast — add a shell alias:
>
> ```bash
> alias o="opener"
> o code ~/src/opener
> ```

### Shell completion

`opener` generates completion scripts via `opener completion <shell>`.

- `opener <TAB>` completes **alias names** (never files).
- `opener <alias> <TAB>` completes **targets** from that alias's `source` (see
  [Configuration](#configuration)) — full paths and URLs, or paths relative to
  the alias's [`base`](#base) if it has one. An alias with no `source` falls
  back to plain file completion; an unknown alias completes nothing.

```bash
# zsh — add to a directory on your $fpath, e.g.
opener completion zsh > "${fpath[1]}/_opener"

# bash
opener completion bash > /opt/homebrew/etc/bash_completion.d/opener

# fish
opener completion fish > ~/.config/fish/completions/opener.fish
```

Run `opener completion <shell> --help` for per-shell details. If you aliased
`opener` to `o`, tell the shell the alias inherits the completion — zsh:
`compdef o=opener`.

> [!NOTE]
> A `command` source runs on every `<TAB>` (with a 2-second timeout). Keep it
> fast.

## Configuration

`opener` reads `~/.opener.yaml` if it exists. Without it, `opener` has no
aliases and every invocation is an "unknown alias" error. A full example lives
in [`examples/opener.yaml`](examples/opener.yaml).

The file has two top-level keys: `aliases` and `sources`.

### `aliases`

Each alias opens its targets as a macOS application (`app`) or a CLI command
(`cmd`):

```yaml
aliases:
  code:
    app: "Visual Studio Code"
  edit:
    cmd: nvim
```

```bash
opener code ~/src/opener   # open -a "Visual Studio Code" ~/src/opener
opener edit README.md      # nvim README.md
```

`cmd` is split into words the way a shell would (single/double quotes
honored) and run directly — **no shell is ever invoked** — with the targets
appended. This lets you bake in fixed flags:

```yaml
aliases:
  chrome:
    cmd: "open -a 'Google Chrome'"
  code:
    cmd: "code -n"
```

An alias that isn't in your config is an error:

```bash
$ opener foo .
opener: unknown alias: foo
```

**Targets are passed through verbatim at open time.** `opener` never validates
or rewrites them — `opener edit opener` runs `nvim opener` literally, whether
or not a file named `opener` exists. The `source` below only affects
completion. The one exception is [`base`](#base), which rewrites targets by
design.

### `base`

Aliases whose targets all live under one directory get long, repetitive
completions: every candidate carries the same absolute prefix, so
`opener code cat<TAB>` matches nothing — completion only matches from the
start of the candidate, and every candidate starts with `/Users/you/…`.

`base` fixes both ends at once:

```yaml
aliases:
  code:
    app: "Visual Studio Code"
    base: $SRC_ROOT
    source:
      kind: dirs-with      # no `roots:` — the walk starts at base
      marker: .git
      depth: 3
```

`base` is where the alias's paths are rooted, in all three senses:

- **Discovering**, a source with no `roots` of its own walks `base`, so you
  name the directory once. A *relative* root (`core-apps`) is resolved
  against `base`, not the current directory; an absolute root overrides it.
- **Completing**, candidates under `base` are offered relative to it —
  `catalog` rather than `/Users/you/src/catalog` — so the short name
  you'd naturally type is the thing that matches. Candidates that don't live
  under `base` (URLs, or paths from a source spanning several roots) keep
  their full form and stay completable that way.
- **Opening**, the target is joined back onto `base`.

A target that already says where it lives is left alone, so an alias with a
`base` can still open anything else:

| target | opens |
| --- | --- |
| `catalog` | `~/src/catalog` |
| `core-apps/api` | `~/src/core-apps/api` |
| `/etc/hosts`, `~/notes.md` | as written |
| `./local`, `../sibling`, `.` | as written |
| `https://example.com` | as written |

Note that only a leading `./` or `../` escapes — a dotfile like `.zshrc`
is an ordinary relative target and joins onto `base` as you'd expect.

> [!WARNING]
> Write `base: $HOME` or `base: "~"`, never a bare `base: ~` — YAML reads
> that as null. `opener` rejects it rather than silently ignoring the
> setting.

### `sources`

A source describes how to discover tab-completion candidates for an alias's
targets. Define reusable ones under `sources:` and reference them by name, or
inline one directly on an alias:

```yaml
sources:
  git-repos:
    kind: dirs-with
    roots: ["~/src", "~/work"]
    marker: .git

aliases:
  code:
    app: "Visual Studio Code"
    source: git-repos           # by name
  edit:
    cmd: nvim
    source:                     # inline
      kind: files
      extensions: [go, md]
```

| kind | fields | completes |
| --- | --- | --- |
| `list` | `items` | the listed paths / URLs |
| `files` | `roots` (default `["."]`), `extensions`, `depth` (default 2) | files under the roots |
| `dirs` | `roots`, `depth` (default 1) | directories under the roots |
| `dirs-with` | `roots`, `marker`, `depth` (default 1) | directories directly containing `marker` (e.g. `.git`) |
| `command` | `run`, `cwd` (optional) | each line of the command's stdout |

Notes:

- `base`, `roots`, and `items` are expanded for `$VAR` / `${VAR}` and a
  leading `~`. An **unset** variable is left in place verbatim rather than
  becoming the empty string, so a typo surfaces as a path that visibly does
  not exist instead of silently collapsing `$SRC_ROT/repo` into `/repo`.
- A root that is absolute or starts with `~` yields **absolute** candidates; a
  relative root (`.`, `sub/`) yields candidates relative to the current
  directory — or to the alias's [`base`](#base), if it has one.
- `depth` counts levels below a root (`1` = direct children). A negative depth
  is unlimited.
- `files` / `dirs` never descend into hidden directories (names starting with
  `.`).
- `command` runs via `sh -c`, so pipes and globs work; `stderr` is discarded
  and it is killed after 2 seconds.
- Completion candidates are full values (full paths, full URLs) so that, say,
  a repo named `opener` under two different roots is unambiguous — unless the
  alias sets a [`base`](#base), which shortens the ones beneath it.
