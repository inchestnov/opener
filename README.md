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

### Automatic mode

```bash
opener <target>
```

A single argument opens `target` automatically: `opener` figures out whether it's a file, a directory, or something else (a URL, a nonexistent path) and picks how to open it.

```bash
opener document.pdf     # opens via the configured PDF rule, or the default app
opener .                # opens the current directory in Finder
opener ~/Downloads      # opens a directory in Finder
opener image.png        # no special rule -> system `open`, Launch Services decides
opener https://github.com   # not a local path -> passed straight to `open`
```

### Alias mode

```bash
opener <alias> <target> [target...]
```

Two or more arguments mean the first one is an alias name, looked up in your config, and the rest are targets passed to whatever that alias launches.

```bash
opener ide .                 # opens . with the app configured for "ide"
opener browser https://github.com
opener editor README.md a.md b.md   # multiple targets, all passed to the same command
```

An alias that isn't in your config is an error:

```bash
$ opener foo .
opener: unknown alias: foo
```

### Verbose mode

Add `-v` / `--verbose` (in any position) to see the full resolution trace on stderr — what target/type was detected, which config rule matched (or didn't), and the exact command that gets run. Verbose mode never changes behavior, only what gets printed.

```bash
opener -v document.pdf
opener document.pdf -v
opener -v ide .
```

### Other flags

```bash
opener --help
opener --version
```

## Configuration

`opener` reads `~/.opener.yaml` if it exists. The file is entirely optional — without it, `opener` falls back to plain `open <target>` for everything.

```yaml
aliases:
  ide:
    app: "Visual Studio Code"

  browser:
    app: "Google Chrome"

  editor:
    command: "nvim"

open:
  directory:
    app: "Visual Studio Code"   # optional override for automatic-mode directories

  files:
    pdf:
      app: "Google Chrome"
```

Two forms are available wherever an app/command choice is configurable — under `aliases.<name>`, `open.directory`, and `open.files.<ext>`:

- `app: "Application Name"` — a macOS GUI application, launched via `open -a "Application Name" <target>...`
- `command: "executable"` — a CLI program, run directly via `exec.Command("executable", <target>...)` (never through a shell)

If both are set on the same rule, `command` takes precedence.

`open.files` isn't limited to `pdf` — any extension key you add is matched (case-insensitively) against a target's file extension in automatic mode.

## Worked examples

**Finder**

```bash
opener .
opener ~/Downloads
```

**PDF, default app**

```bash
opener document.pdf
```

**PDF, forced into Chrome**

```yaml
open:
  files:
    pdf:
      app: "Google Chrome"
```

```bash
opener document.pdf
```

**VS Code via alias**

```yaml
aliases:
  ide:
    app: "Visual Studio Code"
```

```bash
opener ide .
```

**Chrome via alias**

```yaml
aliases:
  browser:
    app: "Google Chrome"
```

```bash
opener browser https://github.com
```

**Neovim via alias**

```yaml
aliases:
  editor:
    command: "nvim"
```

```bash
opener editor README.md
```

## Exit codes

`opener` exits non-zero on any error — an unknown alias, a config file that can't be read or parsed, or a failed launch — and prints a message describing what went wrong to stderr.

## Development

```bash
make test     # go test ./...
make build    # go build -o bin/opener ./cmd/opener
```
