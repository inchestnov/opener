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
```

That's overridable via `~/.opener.yaml`. Give a target a named alias, and point the alias at a CLI program:

```yaml
# ~/.opener.yaml
aliases:
  ide:
    command: nvim
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

`opener` reads `~/.opener.yaml` if it exists. The file is entirely optional — without it, everything falls back to the system `open <target>`.

### `command` — run a CLI program directly

```yaml
aliases:
  editor:
    command: nvim
```

```bash
opener editor README.md   # nvim README.md
```

Targets are appended as-is; `command` is run directly via `exec.Command`, never through a shell.

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

### `cmd` — a full command line for pattern rules

```yaml
open:
  patterns:
    - pattern: ".pdf"
      cmd: "open -a 'Google Chrome'"
```

Like `command`, but for a whole command line rather than a bare executable — useful when you need fixed flags baked in. `cmd` is split into words the way a shell would (quotes honored), then run directly; no shell is ever invoked. Targets are appended to the end.

### `open.directory` — override how directories open

```yaml
open:
  directory:
    app: "Visual Studio Code"
```

```bash
opener ~/go/projects/opener   # opens in Visual Studio Code instead of Finder
```

Same `app`/`command` forms as an alias.
