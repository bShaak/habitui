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

Requires Go 1.23+.

```bash
go install github.com/bShaak/habitui/cmd/habitui@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`, then run:

```bash
habitui
```

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
| `t` | Cycle color theme |
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

| Theme | Mode |
|-------|------|
| `catppuccin-mocha` | Dark (default) |
| `catppuccin-latte` | Light |
| `dracula` | Dark |
| `nord` | Dark |
| `gruvbox-dark` | Dark |
| `gruvbox-light` | Light |
| `tokyo-night` | Dark |
| `rose-pine` | Dark |

### Color slots

| Key | Used for |
|-----|----------|
| `text` | Main text |
| `subtext` | Secondary labels |
| `muted` | Help text / dim UI |
| `surface` | Selection backgrounds |
| `surface_alt` | Elevated surfaces |
| `primary` | Headers and borders |
| `red` | Errors / habit color |
| `orange` | Streaks / habit color |
| `yellow` | Partial completion / habit color |
| `green` | Success / habit color |
| `blue` | Calendar accents / habit color |
| `purple` | Default habit color |
| `pink` | Highlights / habit color |

If `~/.habitui/habitui.config` is missing, Habitui also checks `./habitui.config` in the current directory (useful while developing).

## License

[MIT](LICENSE)
