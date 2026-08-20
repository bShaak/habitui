# Habitui

Habitui is a terminal habit tracker. Add habits, check them off each day, and review streaks and stats.

## Features

- Daily habit list with vim-style navigation
- Per-habit schedule, color, icon, and times-per-day goal
- Weekly calendar for reviewing and toggling past days
- Stats for the last 7 days, 30 days, and year
- Streaks on the main view (shown at 3+ days)
- Multiple color themes (dark and light)
- Local SQLite storage under `~/.habitui/`

## Install

Requires [Go 1.23+](https://go.dev/doc/install) and Git. Habitui is pure Go (no C compiler). It builds for macOS and Linux on amd64 and arm64.

Check that Go is new enough (`go version` should report `go1.23` or later). On macOS, `brew install go git` is enough. On Linux, the default distro package is often too old (for example Ubuntu 24.04’s `golang-go` is 1.22) — use the [official installer](https://go.dev/doc/install) or a versioned package such as `golang-1.23`.

```bash
go install github.com/bShaak/habitui/cmd/habitui@latest
```

If `@latest` cannot resolve (no module proxy / no Git tags), install from the default branch instead:

```bash
go install github.com/bShaak/habitui/cmd/habitui@main
```

`go install` writes the binary to `$(go env GOPATH)/bin` (or `$(go env GOBIN)` if set). That directory is not on `PATH` by default. This line works in bash and zsh; add it to `~/.zshrc` (macOS default) or `~/.bashrc` (most Linux distros), then open a new terminal:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Then run:

```bash
habitui
```

### From source

```bash
git clone https://github.com/bShaak/habitui.git
cd habitui
go install ./cmd/habitui   # or: make install
```

`make` is optional. On Linux, install it with your package manager if you want the targets below (`sudo apt install make` or `sudo dnf install make`). On macOS it ships with Xcode Command Line Tools (`xcode-select --install`).

| Command        | Description                       |
| -------------- | --------------------------------- |
| `make run`     | Run without installing            |
| `make build`   | Build to `bin/habitui`            |
| `make test`    | Run tests                         |
| `make install` | Install to `$(go env GOPATH)/bin` |
| `make clean`   | Remove `bin/`                     |

## CLI

`habitui` launches the TUI by default. For automation (e.g. Reminder Agent), use the `cli` subcommand against the same SQLite database:

```bash
habitui cli list --json
habitui cli complete --id 1 --date 2026-08-03 --json
```

| Flag / env | Description |
| ---------- | ----------- |
| `--json` | Machine-readable JSON on stdout |
| `--db path` | SQLite path (default: `~/.habitui/habit.db`) |
| `HABITUI_DB` | Same as `--db` when the flag is omitted |
| `--date YYYY-MM-DD` | Day to list or complete (default: today) |

`list` includes `due` (scheduled that day) and `complete` (goal met). On `complete`, `already_complete` is true only when the habit was already at/above goal (no write); a new completion always has `already_complete: false` and a `completion` object. Habit create/edit/delete stays in the TUI.

## Usage

| Key            | Action                                 |
| -------------- | -------------------------------------- |
| `j` / `k`      | Move selection                         |
| `enter`        | Toggle completion for today            |
| `a`            | Add habit                              |
| `e`            | Edit selected habit                    |
| `x`            | Delete selected habit (`y` to confirm) |
| `t`            | Cycle color theme                      |
| `c`            | Week calendar                          |
| `s`            | Statistics                             |
| `esc`          | Back to main view                      |
| `q` / `ctrl+c` | Quit                                   |

### Calendar

| Key       | Action                                 |
| --------- | -------------------------------------- |
| `h` / `l` | Previous / next day                    |
| `j` / `k` | Previous / next habit                  |
| `H` / `L` | Previous / next week                   |
| `enter`   | Toggle completion for the selected day |

### Stats

| Key       | Action             |
| --------- | ------------------ |
| `←` / `→` | Switch period tabs |

## Data

Everything lives in `~/.habitui/`:

| Path                        | Purpose                  |
| --------------------------- | ------------------------ |
| `~/.habitui/habit.db`       | Habits and completions   |
| `~/.habitui/habitui.config` | Optional color overrides |

## Configuration

Themes and optional color overrides live in `~/.habitui/habitui.config`. Press `t` on the main view to cycle themes (selection is saved automatically).

```json
{
  "theme": "catppuccin-mocha",
  "base": {
    "primary": "#b4befe",
    "green": "#51cf66",
    "text": "#ffffff"
  }
}
```

### Built-in themes

| Theme              | Mode           |
| ------------------ | -------------- |
| `catppuccin-mocha` | Dark (default) |
| `catppuccin-latte` | Light          |
| `dracula`          | Dark           |
| `nord`             | Dark           |
| `gruvbox-dark`     | Dark           |
| `gruvbox-light`    | Light          |
| `tokyo-night`      | Dark           |
| `rose-pine`        | Dark           |

### Color slots

| Key           | Used for                         |
| ------------- | -------------------------------- |
| `text`        | Main text                        |
| `subtext`     | Secondary labels                 |
| `muted`       | Help text / dim UI               |
| `surface`     | Selection backgrounds            |
| `surface_alt` | Elevated surfaces                |
| `primary`     | Headers and borders              |
| `red`         | Errors / habit color             |
| `orange`      | Streaks / habit color            |
| `yellow`      | Partial completion / habit color |
| `green`       | Success / habit color            |
| `blue`        | Calendar accents / habit color   |
| `purple`      | habit color                      |
| `pink`        | Highlights / habit color         |

If `~/.habitui/habitui.config` is missing, Habitui also checks `./habitui.config` in the current directory (useful while developing).

## License

[MIT](LICENSE)
