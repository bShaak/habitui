# Habitui

Habitui is a terminal habit tracker. Add habits, check them off each day, browse a week calendar, and review streaks and stats — all from your terminal.

## Features

- Daily habit list with vim-style navigation
- Per-habit schedule, color, icon, and times-per-day goal
- Week calendar for reviewing and toggling past days
- Stats for the last 7 days, 30 days, and year
- Streaks on the main view (shown at 3+ days)
- Local SQLite storage under `~/.habitui/`

## Install

Requires Go 1.23+.

```bash
go install github.com/bShaak/habitui/cmd/habitui@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`, then run:

```bash
habitui
```

No C compiler is required — Habitui uses a pure-Go SQLite driver.

### From source

```bash
git clone https://github.com/bShaak/habitui.git
cd habitui
make install   # or: make run
```

| Command | Description |
|---------|-------------|
| `make run` | Run without installing |
| `make build` | Build to `bin/habitui` |
| `make test` | Run tests |
| `make install` | Install to `$(go env GOPATH)/bin` |
| `make clean` | Remove `bin/` |

## Usage

| Key | Action |
|-----|--------|
| `j` / `k` | Move selection |
| `enter` | Toggle completion for today |
| `a` | Add habit |
| `e` | Edit selected habit |
| `x` | Delete selected habit (`y` to confirm) |
| `c` | Week calendar |
| `s` | Statistics |
| `esc` | Back to main view |
| `q` / `ctrl+c` | Quit |

### Calendar

| Key | Action |
|-----|--------|
| `h` / `l` | Previous / next day |
| `j` / `k` | Previous / next habit |
| `H` / `L` | Previous / next week |
| `enter` | Toggle completion for the selected day |

### Stats

| Key | Action |
|-----|--------|
| `←` / `→` | Switch period tabs |

## Data

Everything lives in `~/.habitui/`:

| Path | Purpose |
|------|---------|
| `~/.habitui/habit.db` | Habits and completions |
| `~/.habitui/habitui.config` | Optional color overrides |

## Configuration

Optional color overrides use Catppuccin Mocha defaults. Create `~/.habitui/habitui.config`:

```json
{
  "base": {
    "red": "#ff6b6b",
    "green": "#51cf66",
    "text": "#ffffff"
  }
}
```

| Key | Default | Used for |
|-----|---------|----------|
| `rosewater` | `#f5e0dc` | UI accents |
| `flamingo` | `#f2cdcd` | UI accents |
| `mauve` | `#cba6f7` | Default habit color |
| `red` | `#f38ba8` | Habit color |
| `peach` | `#fab387` | Habit color / streak accent |
| `yellow` | `#f9e2af` | Habit color / partial completion |
| `green` | `#a6e3a1` | Habit color / success |
| `teal` | `#94e2d5` | UI accents |
| `sky` | `#89dceb` | UI accents |
| `sapphire` | `#74c7ec` | UI accents |
| `blue` | `#89b4fa` | Habit color |
| `lavender` | `#b4befe` | Headers and borders |
| `text` | `#cdd6f4` | Main text |
| `subtext` | `#bac2de` | Secondary text |
| `overlay` | `#9399b2` | Muted text / help |
| `surface1` | `#45475a` | Selection backgrounds |
| `surface2` | `#585b70` | Elevated backgrounds |
| `pink` | `#f5c2e7` | Habit color |

If `~/.habitui/habitui.config` is missing, Habitui also checks `./habitui.config` in the current directory (useful while developing).

## License

[MIT](LICENSE)
