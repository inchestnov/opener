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

```bash
opener <target>
```

A single argument opens `target`: `opener` figures out whether it's a file, a directory, or something else (a URL, a nonexistent path) and opens it accordingly.

```bash
opener document.pdf     # opens via the configured PDF rule, or the default app
opener .                # opens the current directory in Finder
opener ~/Downloads      # opens a directory in Finder
opener image.png        # no special rule -> system `open`, Launch Services decides
opener https://github.com   # not a local path -> passed straight to `open`
```

Override how a file type opens via `~/.opener.yaml`, matching a glob pattern against the filename:

```yaml
open:
  patterns:
    - pattern: "*.pdf"
      app: "Google Chrome"
```

```bash
opener document.pdf     # now opens in Chrome instead of the default app
```

`pattern` can also be a full command line instead of an app name, via `cmd` (word-split, never run through a shell):

```yaml
open:
  patterns:
    - pattern: ".pdf"
      cmd: "open -a 'Google Chrome'"
```

Patterns are checked in order; the first match wins.

For anything you want a shortcut to — not just file types — define an alias and pass it as the first argument, with the target(s) after it:

```yaml
aliases:
  ide:
    app: "Visual Studio Code"
```

```bash
opener ide .             # opens . with the app configured for "ide"
```

Aliases work with CLI commands too (`command: "nvim"`), and can take multiple targets: `opener editor a.md b.md`. An alias that isn't in your config is an error:

```bash
$ opener foo .
opener: unknown alias: foo
```
