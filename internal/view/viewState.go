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
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

type model struct {
	habits    	[]types.Habit
	completions []types.Completion
	store  			*storage.SQLiteStore
	form 				*huh.Form
	lg         	*lipgloss.Renderer
	styles     	*Styles
	getView    	func(m model) string
	getUpdate  	func(m model,msg tea.Msg) (tea.Model, tea.Cmd)
	cursor     	int
}

func (m model) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		80,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
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

	lg := lipgloss.DefaultRenderer()

	return model{
		habits:    habits,
		completions: completions,
		form: nil,
		store:     store,
		lg:        lg,
		styles:    NewStyles(lg),
		getView:   GetMainView,
		getUpdate: GetMainUpdate,
		cursor:    0,
	}
}

// // Init
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
