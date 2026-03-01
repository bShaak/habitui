package view

import (
	"context"
	"log"
	"time"

	types "github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	rosewater = lipgloss.Color("#f5e0dc")
	flamingo  = lipgloss.Color("#f2cdcd")
	mauve     = lipgloss.Color("#cba6f7")
	red       = lipgloss.Color("#f38ba8")
	peach     = lipgloss.Color("#fab387")
	yellow    = lipgloss.Color("#f9e2af")
	green     = lipgloss.Color("#a6e3a1")
	teal      = lipgloss.Color("#94e2d5")
	sky       = lipgloss.Color("#89dceb")
	sapphire  = lipgloss.Color("#74c7ec")
	blue      = lipgloss.Color("#89b4fa")
	lavender  = lipgloss.Color("#b4befe")
	text      = lipgloss.Color("#cdd6f4")
	subtext   = lipgloss.Color("#bac2de")
	overlay   = lipgloss.Color("#9399b2")
	surface1  = lipgloss.Color("#45475a")
	surface2  = lipgloss.Color("#585b70")
	pink      = lipgloss.Color("#f5c2e7")
)

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
