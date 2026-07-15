package view

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"
	"github.com/bShaak/habitui/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	screenMain screen = iota
	screenCalendar
	screenStats
	screenCreateHabit
	screenEditHabit
)

var (
	currentTheme  theme.Theme
	themeConfig   *theme.Config
)

var (
	text       lipgloss.Color
	subtext    lipgloss.Color
	muted      lipgloss.Color
	surface    lipgloss.Color
	surfaceAlt lipgloss.Color
	primary    lipgloss.Color
	red        lipgloss.Color
	orange     lipgloss.Color
	yellow     lipgloss.Color
	green      lipgloss.Color
	blue       lipgloss.Color
	purple     lipgloss.Color
	pink       lipgloss.Color
)

func initTheme() {
	config, err := theme.LoadConfig()
	if err != nil {
		log.Printf("Error loading theme config: %s, using default", err)
		config = &theme.Config{}
	}
	themeConfig = config
	currentTheme = theme.GetTheme(config)
	applyThemeColors()
}

func applyThemeColors() {
	t := currentTheme.Base
	text = lipgloss.Color(t.Text)
	subtext = lipgloss.Color(t.Subtext)
	muted = lipgloss.Color(t.Muted)
	surface = lipgloss.Color(t.Surface)
	surfaceAlt = lipgloss.Color(t.SurfaceAlt)
	primary = lipgloss.Color(t.Primary)
	red = lipgloss.Color(t.Red)
	orange = lipgloss.Color(t.Orange)
	yellow = lipgloss.Color(t.Yellow)
	green = lipgloss.Color(t.Green)
	blue = lipgloss.Color(t.Blue)
	purple = lipgloss.Color(t.Purple)
	pink = lipgloss.Color(t.Pink)
}

func cycleTheme(m Model) Model {
	if themeConfig == nil {
		themeConfig = &theme.Config{}
	}
	next := theme.NextThemeName(currentTheme.Name)
	themeConfig.Theme = next
	currentTheme = theme.GetTheme(themeConfig)
	applyThemeColors()
	m.styles = newStyles(m.lg)
	m.statusMsg = "Theme: " + currentTheme.Name
	if err := theme.SaveConfig(themeConfig); err != nil {
		log.Printf("Error saving theme config: %s", err)
	}
	return m
}

func getHabitColor(colorName string) lipgloss.Color {
	colorKey := colorName
	if colorKey == "" {
		colorKey = "red"
	}
	colorKey = strings.ToLower(colorKey)

	colorMap := map[string]lipgloss.Color{
		"red":    red,
		"blue":   blue,
		"green":  green,
		"yellow": yellow,
		"orange": orange,
		"purple": purple,
		"pink":   pink,
	}

	c, ok := colorMap[colorKey]
	if !ok {
		c = red
	}
	return c
}

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help,
	ContentBox,
	Title lipgloss.Style
}

func newStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 2, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(primary).
		Bold(true).
		MarginLeft(2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primary).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(pink)
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(muted)
	s.ContentBox = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primary).
		Padding(1, 2).
		MarginLeft(2)
	s.Title = lg.NewStyle().
		Foreground(primary).
		Bold(true).
		MarginLeft(2)
	return &s
}

type Model struct {
	habits            []models.Habit
	completions       []models.Completion
	streakCompletions []models.Completion
	statsCompletions  []models.Completion
	store             storage.Store
	form              *huh.Form
	formFields        *habitFormFields
	lg                *lipgloss.Renderer
	styles            *Styles
	screen            screen
	cursor            int
	weekStart         time.Time
	weekCompletions   []models.Completion
	calendarCol       int
	scrollOffset      int
	statsTab          int
	width             int
	height            int
	confirmingDelete  bool
	statusMsg         string
	viewDay time.Time
}

func (m Model) appBoundaryView(title string) string {
	return lipgloss.PlaceHorizontal(
		0,
		lipgloss.Left,
		m.styles.HeaderText.Render(title),
		lipgloss.WithWhitespaceForeground(primary),
	)
}

func (m Model) renderTitle() string {
	return m.styles.Title.Render("Habitui")
}

func (m Model) Close() error {
	if m.store == nil {
		return nil
	}
	return m.store.Close()
}

func InitViewState() Model {
	initTheme()
	store, err := storage.OpenSQLite()
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	habits, err := store.ListHabits(context.Background())
	if err != nil {
		log.Fatalf("Error fetching habits: %s", err)
	}

	completions, err := store.GetCompletionsByDate(context.Background(), time.Now())
	if err != nil {
		log.Fatalf("Error fetching completions: %s", err)
	}

	now := time.Now()
	weekStart := getMonday(now)
	weekEnd := weekStart.AddDate(0, 0, 6)
	weekCompletions, err := store.GetCompletionsByDateRange(context.Background(), weekStart, weekEnd)
	if err != nil {
		log.Fatalf("Error fetching week completions: %s", err)
	}

	streakCompletions, err := loadStreakCompletions(store, now)
	if err != nil {
		log.Fatalf("Error fetching streak completions: %s", err)
	}

	lg := lipgloss.DefaultRenderer()

	return Model{
		habits:            habits,
		completions:       completions,
		streakCompletions: streakCompletions,
		store:             store,
		lg:                lg,
		styles:            newStyles(lg),
		screen:            screenMain,
		cursor:            0,
		weekStart:         weekStart,
		weekCompletions:   weekCompletions,
		calendarCol:       0,
		viewDay:           startOfDay(now),
	}
}

const (
	streakLookbackYears    = 5
	dayRefreshTickInterval = 5 * time.Minute
)

type dayRefreshTickMsg time.Time

func dayRefreshTick() tea.Cmd {
	return tea.Tick(dayRefreshTickInterval, func(t time.Time) tea.Msg {
		return dayRefreshTickMsg(t)
	})
}

func loadStreakCompletions(store storage.Store, now time.Time) ([]models.Completion, error) {
	streakStart := startOfDay(now.AddDate(-streakLookbackYears, 0, 0))
	return store.GetCompletionsByDateRange(context.Background(), streakStart, now)
}

func (m Model) Init() tea.Cmd {
	return dayRefreshTick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.scrollOffset = clampScroll(m.scrollOffset, 0, maxScroll(m))
		if m.form != nil {
			applyFormSize(m.form, m.width, m.height)
			// Don't forward to huh — it pads all groups to the tallest page height.
			return m, nil
		}
		return m.updateScreen(msg)
	case tea.FocusMsg:
		// Terminal regained focus (e.g. morning after overnight sleep).
		return refreshIfDayChanged(m), nil
	case dayRefreshTickMsg:
		// Catch midnight rollover when the terminal stays focused and idle.
		return refreshIfDayChanged(m), dayRefreshTick()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.confirmingDelete {
				m.confirmingDelete = false
				return m, nil
			}
			completions, err := m.store.GetCompletionsByDate(context.Background(), time.Now())
			if err != nil {
				log.Printf("Error fetching today's completions: %s", err)
			}
			m.completions = completions
			m = refreshStreakCompletions(m)
			m.form = nil
			m.formFields = nil
			m.confirmingDelete = false
			m.scrollOffset = 0
			m.screen = screenMain
			return m, nil
		case "pgup", "ctrl+u":
			if m.form == nil {
				m.scrollOffset = clampScroll(m.scrollOffset-pageScrollAmount(m), 0, maxScroll(m))
				return m, nil
			}
		case "pgdown", "ctrl+d":
			if m.form == nil {
				m.scrollOffset = clampScroll(m.scrollOffset+pageScrollAmount(m), 0, maxScroll(m))
				return m, nil
			}
		}
	}
	return m.updateScreen(msg)
}

func (m Model) updateScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenCalendar:
		return updateCalendar(m, msg)
	case screenStats:
		return updateStats(m, msg)
	case screenCreateHabit:
		return updateCreateHabit(m, msg)
	case screenEditHabit:
		return updateEditHabit(m, msg)
	default:
		return updateMain(m, msg)
	}
}

func (m Model) View() string {
	content := m.screenContent()
	// Forms manage their own height/scrolling via huh; don't double-clip them.
	if m.form != nil {
		return content
	}
	return applyViewport(content, m.height, m.scrollOffset)
}

func (m Model) screenContent() string {
	switch m.screen {
	case screenCalendar:
		return viewCalendar(m)
	case screenStats:
		return viewStats(m)
	case screenCreateHabit:
		return viewCreateHabit(m)
	case screenEditHabit:
		return viewEditHabit(m)
	default:
		return viewMain(m)
	}
}

func pageScrollAmount(m Model) int {
	if m.height <= 2 {
		return 1
	}
	return m.height / 2
}

func maxScroll(m Model) int {
	if m.height <= 0 {
		return 0
	}
	lines := strings.Split(m.screenContent(), "\n")
	if len(lines) <= m.height {
		return 0
	}
	return len(lines) - m.height
}

func clampScroll(offset, min, max int) int {
	if offset < min {
		return min
	}
	if offset > max {
		return max
	}
	return offset
}

func applyViewport(content string, height, offset int) string {
	if height <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	if len(lines) <= height {
		return content
	}
	if offset < 0 {
		offset = 0
	}
	maxOff := len(lines) - height
	if offset > maxOff {
		offset = maxOff
	}
	visible := lines[offset : offset+height]
	return strings.Join(visible, "\n")
}

func refreshStreakCompletions(m Model) Model {
	completions, err := loadStreakCompletions(m.store, time.Now())
	if err != nil {
		log.Printf("Error fetching streak completions: %s", err)
		return m
	}
	m.streakCompletions = completions
	return m
}

func needsDayRefresh(viewDay, now time.Time) bool {
	if viewDay.IsZero() {
		return true
	}
	return !startOfDay(viewDay).Equal(startOfDay(now))
}

func refreshIfDayChanged(m Model) Model {
	now := time.Now()
	if !needsDayRefresh(m.viewDay, now) {
		return m
	}
	return refreshForDay(m, now)
}

func refreshForDay(m Model, now time.Time) Model {
	completions, err := m.store.GetCompletionsByDate(context.Background(), now)
	if err != nil {
		log.Printf("Error fetching today's completions: %s", err)
		// Leave viewDay unchanged so the next focus/tick retries.
		return m
	}
	m.completions = completions

	weekStart := getMonday(now)
	weekEnd := weekStart.AddDate(0, 0, 6)
	weekCompletions, err := m.store.GetCompletionsByDateRange(context.Background(), weekStart, weekEnd)
	if err != nil {
		log.Printf("Error fetching week completions: %s", err)
	} else {
		m.weekStart = weekStart
		m.weekCompletions = weekCompletions
	}

	m = refreshStreakCompletions(m)
	m.viewDay = startOfDay(now)
	return m
}

func returnToMain(m Model) Model {
	m.form = nil
	m.formFields = nil
	m.screen = screenMain
	return m
}
