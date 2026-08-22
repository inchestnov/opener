# opener

`opener` is a small macOS CLI that gives you one interface for opening files, directories, and applications — hiding the differences between specific macOS apps behind a layer of aliases and user configuration.

```bash
opener <target>
opener <alias> <target>
```

`opener` doesn't implement its own file-opening mechanism. It's a thin wrapper around macOS's native `open` command and Launch Services, plus a configurable layer of aliases on top.

## Install

```bash
go install github.com/inchestnov/opener/cmd/opener@latest
```

This puts `opener` in `$(go env GOPATH)/bin` — make sure that's on your `PATH`.

### Building locally instead

If you're working on `opener` itself, build from a checkout of this repo:

```bash
make build      # builds ./bin/opener
```

or directly with Go:

```bash
go build -o bin/opener ./cmd/opener
```

## Usage

Out of the box, `opener` just uses the system default:

```bash
opener docs.pdf              # Open using default app
opener ~/go/projects/opener  # Open directory in Finder
opener https://github.com    # Open URL in default browser
```

A target that doesn't exist and isn't a URL is an error, and so is an executable file — `opener` won't guess at running it for you; use `./script.sh` directly.

That's overridable via `~/.opener.yaml`. Give a target a named alias, and point the alias at a CLI program:

```yaml
# ~/.opener.yaml
aliases:
  ide:
    cmd: nvim
```

```bash
opener ide ~/go/projects/opener   # nvim ~/go/projects/opener
```

or at a macOS application:

```yaml
# ~/.opener.yaml
aliases:
  ide:
    app: "Visual Studio Code"
```

```bash
opener ide ~/go/projects/opener   # opens in Visual Studio Code
```

An alias that isn't in your config is an error:

```bash
$ opener foo .
opener: unknown alias: foo
```

Typing `opener` in full gets old fast — add a shell alias:

```bash
alias o="opener"
o docs.pdf
```

## Configuration

`opener` reads `~/.opener.yaml` if it exists. The file is entirely optional — without it, existing files, directories, and URLs fall back to the system `open <target>`.

### `cmd` — run a CLI program directly

```yaml
aliases:
  editor:
    cmd: nvim
```

```bash
opener editor README.md   # nvim README.md
```

Targets are appended as-is; `cmd` here is a bare executable, run directly via `exec.Command`, never through a shell.

### `app` — open in a macOS application

```yaml
aliases:
  browser:
    app: "Google Chrome"
```

```bash
opener browser https://github.com   # open -a "Google Chrome" https://github.com
```

### `pattern` — match files by extension or glob

```yaml
open:
  patterns:
    - pattern: "*.pdf"
      app: "Google Chrome"
```

```bash
opener document.pdf   # now opens in Chrome instead of the default app
```

`pattern` can be a glob (`*.pdf`) or a bare extension (`.pdf`, treated the same as `*.pdf`), matched case-insensitively against the filename. Patterns are checked in order; the first match wins.

To match several extensions with one rule, use a brace group:

```yaml
open:
  patterns:
    - pattern: "*.{jpg,png,gif}"
      app: "Preview"
```

### `cmd` for pattern rules — a full command line

```yaml
open:
  patterns:
    - pattern: ".pdf"
      cmd: "open -a 'Google Chrome'"
```

Unlike the alias/`open.directory` form of `cmd` above, this one takes a whole command line rather than a bare executable — useful when you need fixed flags baked in. It's split into words the way a shell would (quotes honored), then run directly; no shell is ever invoked. Targets are appended to the end.

### `templates` — a link to a fixed file or project

```yaml
templates:
  payment-service:
    path: ~/repos/payment-service
    app: "Visual Studio Code"
```

```bash
opener payment-service   # opens ~/repos/payment-service in Visual Studio Code
```

A template is a named shortcut, not a real target: `opener <name>` ignores
whatever exists on disk at `<name>` and opens `path` instead. Checked before
automatic-mode's usual file/directory/URL resolution, so a template name
always wins over an on-disk file of the same name.

`app`/`cmd` are optional. Set them to pin exactly how `path` opens, same as
an alias. Leave both unset and `path` is resolved exactly as if you'd typed
it directly — so a template can just pin *what* opens, and let `open.patterns`
/ `open.directory` (or the system default) decide *how*:

```yaml
templates:
  payment-service:
    path: ~/repos/payment-service   # a directory -> opens in Finder (or open.directory, if set)
  invoice:
    path: ~/Downloads/invoice.pdf   # a file -> matched against open.patterns, same as any .pdf
```

Any combination works: a file pinned to an app, a directory pinned to a CLI
command, or either left to fall back to whatever automatic mode would
otherwise do.

### `open.directory` — override how directories open

```yaml
open:
  directory:
    app: "Visual Studio Code"
```

```bash
opener ~/go/projects/opener   # opens in Visual Studio Code instead of Finder
```

Same `app`/`cmd` forms as an alias.
