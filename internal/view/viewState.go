package view

import (
	"context"
	"log"
	"strings"
	"time"

	types "github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"
	"github.com/bShaak/habitui/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var currentTheme theme.Theme

var (
	rosewater lipgloss.Color
	flamingo  lipgloss.Color
	mauve     lipgloss.Color
	red       lipgloss.Color
	peach     lipgloss.Color
	yellow    lipgloss.Color
	green     lipgloss.Color
	teal      lipgloss.Color
	sky       lipgloss.Color
	sapphire  lipgloss.Color
	blue      lipgloss.Color
	lavender  lipgloss.Color
	text      lipgloss.Color
	subtext   lipgloss.Color
	overlay   lipgloss.Color
	surface1  lipgloss.Color
	surface2  lipgloss.Color
	pink      lipgloss.Color
)

func initTheme() {
	config, err := theme.LoadConfig()
	if err != nil {
		log.Printf("Error loading theme config: %s, using default", err)
		config = &theme.Config{}
	}
	currentTheme = theme.GetTheme(config)
	applyThemeColors()
}

func applyThemeColors() {
	t := currentTheme.Base
	rosewater = lipgloss.Color(t.Rosewater)
	flamingo = lipgloss.Color(t.Flamingo)
	mauve = lipgloss.Color(t.Mauve)
	red = lipgloss.Color(t.Red)
	peach = lipgloss.Color(t.Peach)
	yellow = lipgloss.Color(t.Yellow)
	green = lipgloss.Color(t.Green)
	teal = lipgloss.Color(t.Teal)
	sky = lipgloss.Color(t.Sky)
	sapphire = lipgloss.Color(t.Sapphire)
	blue = lipgloss.Color(t.Blue)
	lavender = lipgloss.Color(t.Lavender)
	text = lipgloss.Color(t.Text)
	subtext = lipgloss.Color(t.Subtext)
	overlay = lipgloss.Color(t.Overlay)
	surface1 = lipgloss.Color(t.Surface1)
	surface2 = lipgloss.Color(t.Surface2)
	pink = lipgloss.Color(t.Pink)
}

func GetHabitColor(colorName string) lipgloss.Color {
	colorKey := colorName
	if colorKey == "" {
		colorKey = "purple"
	}
	colorKey = strings.ToLower(colorKey)

	colorMap := map[string]lipgloss.Color{
		"red":    lipgloss.Color(currentTheme.Base.Red),
		"blue":   lipgloss.Color(currentTheme.Base.Blue),
		"green":  lipgloss.Color(currentTheme.Base.Green),
		"yellow": lipgloss.Color(currentTheme.Base.Yellow),
		"orange": lipgloss.Color(currentTheme.Base.Peach),
		"purple": lipgloss.Color(currentTheme.Base.Mauve),
		"pink":   lipgloss.Color(currentTheme.Base.Pink),
	}

	c, ok := colorMap[colorKey]
	if !ok {
		c = lipgloss.Color(currentTheme.Base.Mauve)
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

func getMonday(t time.Time) time.Time {
	weekday := t.Weekday()
	daysSinceMonday := (int(weekday) + 6) % 7
	return t.AddDate(0, 0, -daysSinceMonday)
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 2, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(lavender).
		Bold(true).
		MarginLeft(2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lavender).
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
		Foreground(overlay)
	s.ContentBox = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lavender).
		Padding(1, 2).
		MarginLeft(2)
	s.Title = lg.NewStyle().
		Foreground(lavender).
		Bold(true).
		MarginLeft(2)
	return &s
}

type model struct {
	habits          []types.Habit
	completions     []types.Completion
	store           *storage.SQLiteStore
	form            *huh.Form
	lg              *lipgloss.Renderer
	styles          *Styles
	getView         func(m model) string
	getUpdate       func(m model, msg tea.Msg) (tea.Model, tea.Cmd)
	cursor          int
	weekStart       time.Time
	weekCompletions []types.Completion
	calendarCol     int
	scrollOffset    int
	statsTab        int
}

func (m model) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		0,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceForeground(lavender),
	)
}

func (m model) renderTitle() string {
	return m.styles.Title.Render("Habitui")
}

func InitViewState() model {
	initTheme()
	store, err := storage.OpenSQLite()
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
		defer store.Close()
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

	lg := lipgloss.DefaultRenderer()

	return model{
		habits:          habits,
		completions:     completions,
		form:            nil,
		store:           store,
		lg:              lg,
		styles:          NewStyles(lg),
		getView:         GetMainView,
		getUpdate:       GetMainUpdate,
		cursor:          0,
		weekStart:       weekStart,
		weekCompletions: weekCompletions,
		calendarCol:     0,
	}
}

// Init
func (m model) Init() tea.Cmd {
	return nil
}

// Update
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			defer m.store.Close()
			return m, tea.Quit
		case "esc":
			completions, err := m.store.GetCompletionsByDate(context.Background(), time.Now())
			if err != nil {
				log.Printf("Error fetching today's completions: %s", err)
			}
			m.completions = completions
			m.getView = GetMainView
			m.getUpdate = GetMainUpdate
			return m, nil
		}
	}
	return m.getUpdate(m, msg)
}

// View
func (m model) View() string {
	return m.getView(m)
}
