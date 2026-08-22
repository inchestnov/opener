# opener

`opener` is a small macOS CLI that gives you one interface for opening files, directories, and applications — hiding the differences between specific macOS apps behind a layer of aliases and user configuration.

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

Out of the box, `opener` just uses the system default:

```bash
opener docs.pdf              # Open using default app
opener ~/go/projects/opener  # Open directory in Finder
opener https://github.com    # Open URL in default browser
```

> [!TIP]
> Typing `opener` in full gets old fast — add a shell alias:
>
> ```bash
> alias o="opener"
> o docs.pdf
> ```

## Configuration

`opener` reads `~/.opener.yaml` if it exists. The file is entirely optional — without it, existing files, directories, and URLs fall back to the system `open <target>`.

#### `open.patterns` — match files by extension or glob

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

A pattern rule's `cmd` takes a whole command line rather than a bare executable — useful when you need fixed flags baked in:

```yaml
open:
  patterns:
    - pattern: ".pdf"
      cmd: "open -a 'Google Chrome'"
```

It's split into words the way a shell would (quotes honored), then run directly; no shell is ever invoked. Targets are appended to the end.

#### `open.directory` — override how directories open

```yaml
open:
  directory:
    app: "Visual Studio Code"
```

```bash
opener ~/go/projects/opener   # opens in Visual Studio Code instead of Finder
```

#### `aliases` — named shortcuts for `opener <alias> <target>...`

Point an alias at a CLI program:

```yaml
aliases:
  ide:
    cmd: nvim
```

```bash
opener ide ~/go/projects/opener   # nvim ~/go/projects/opener
```

or at a macOS application:

```yaml
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

#### `templates` — a link to a fixed file or project

```yaml
templates:
  bashrc:
    path: ~/.bashrc
    cmd: vim
```

```bash
opener bashrc   # vim ~/.bashrc
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
  vimrc:
    path: ~/.vimrc                 # a file -> matched against open.patterns, same as any other file
  nvim-config:
    path: ~/.config/nvim           # a directory -> opens in Finder (or open.directory, if set)
```

Any combination works: a file pinned to an app, a directory pinned to a CLI
command, or either left to fall back to whatever automatic mode would
otherwise do.
