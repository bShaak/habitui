# Habitui

Habitui is a terminal-based utility that helps you track and manage your habits.

## Features

- **Simple Interface**: Easy-to-use terminal interface for quick habit tracking.
- **Customizable Habits**: Add, edit, and remove habits as needed.
- **Daily Tracking**: Mark habits as completed for each day.
- **Statistics**: View your progress and habit completion rates over time.
- **Data Persistence**: Save your habit data locally for future reference.

## Configuration

Create a `habitui.config` file to customize colors. Available color keys:

| Key | Default | Used For |
|-----|---------|----------|
| `rosewater` | `#f5e0dc` | UI accents |
| `flamingo` | `#f2cdcd` | UI accents |
| `mauve` | `#cba6f7` | Default habit color |
| `red` | `#f38ba8` | Habit color |
| `peach` | `#fab387` | Habit color |
| `yellow` | `#f9e2af` | Habit color |
| `green` | `#a6e3a1` | Habit color, success states |
| `teal` | `#94e2d5` | UI accents |
| `sky` | `#89dceb` | UI accents |
| `sapphire` | `#74c7ec` | UI accents |
| `blue` | `#89b4fa` | Habit color |
| `lavender` | `#b4befe` | UI accents |
| `text` | `#cdd6f4` | Main text |
| `subtext` | `#bac2de` | Secondary text |
| `overlay` | `#9399b2` | Muted text |
| `surface1` | `#45475a` | Backgrounds |
| `surface2` | `#585b70` | Elevated backgrounds |
| `pink` | `#f5c2e7` | Habit color |

### Example

```json
{
  "base": {
    "red": "#ff6b6b",
    "green": "#51cf66",
    "text": "#ffffff"
  }
}
```

## TODO

- [x] add ability to customize habits (add, delete, modify)
- [x] Add data persistance to local db
- [x] Add create habit view when pressing 'a'
- [x] Delete habits
- [x] Add daily tracking.
- [x] Show progress in daily view
- [x] Edit habits
- [x] Add calendar week view in tui
- [x] Add different colours for different habits
- [x] Add different themes and ability to switch between them
- [x] Populate the calendar view based on schedule
- [ ] Add statistics for your habits completions
- [ ] Add progress bars for weekly habit goals, inferred from habits schedule
- [ ] Habit data and stats for selected habit in daily view
- [ ] monthly view
- [ ] make project simple to install, on linux and mac
