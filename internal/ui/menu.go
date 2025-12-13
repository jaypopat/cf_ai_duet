package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

type MenuChoice int

const (
	ChoiceNone MenuChoice = iota
	ChoiceCreate
	ChoiceJoin
)

type MenuModel struct {
	choice    int
	selected  MenuChoice
	roomID    string
	inputMode bool
	input     string
	width     int
	height    int
}

var (
	accentColor = lipgloss.Color("#00FFB3")
	dimColor    = lipgloss.Color("#4A4A4A")
	textColor   = lipgloss.Color("#E0E0E0")
	bgAccent    = lipgloss.Color("#0A0A0A")
	borderColor = lipgloss.Color("#333333")

	titleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	dividerStyle = lipgloss.NewStyle().
			Foreground(borderColor)

	optionStyle = lipgloss.NewStyle().
			Foreground(textColor)

	selectedStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			Padding(0, 1).
			Width(30)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true)

	inputTextStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	placeholderStyle = lipgloss.NewStyle().
				Foreground(dimColor).
				Italic(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	keyStyle = lipgloss.NewStyle().
			Foreground(accentColor)
)

func NewMenuModel() MenuModel {
	return MenuModel{
		choice:    0,
		selected:  ChoiceNone,
		inputMode: false,
		input:     "",
	}
}

// RunMenu wraps the Bubble Tea program for the SSH session.
func RunMenu(session ssh.Session) (MenuModel, error) {
	model := NewMenuModel()
	program := tea.NewProgram(model,
		tea.WithInput(session),
		tea.WithOutput(session),
		tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		return model, err
	}

	return finalModel.(MenuModel), nil
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.inputMode {
			switch msg.String() {
			case "enter":
				trimmed := strings.TrimSpace(m.input)
				if len(trimmed) > 0 {
					m.roomID = trimmed
					m.selected = ChoiceJoin
					return m, tea.Quit
				}
			case "esc":
				m.inputMode = false
				m.input = ""
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case "ctrl+u":
				m.input = ""
			default:
				chunk := msg.String()
				if len(chunk) == 0 {
					return m, nil
				}
				clean := strings.TrimSpace(strings.ToLower(chunk))
				if clean == "" {
					return m, nil
				}
				for _, r := range clean {
					if len(m.input) >= 32 {
						break
					}
					if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
						m.input += string(r)
					}
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.choice > 0 {
				m.choice--
			}
		case "down", "j":
			if m.choice < 1 {
				m.choice++
			}
		case "1":
			m.choice = 0
			m.selected = ChoiceCreate
			return m, tea.Quit
		case "2":
			m.choice = 1
			m.inputMode = true
		case "enter":
			if m.choice == 0 {
				m.selected = ChoiceCreate
				return m, tea.Quit
			} else {
				m.inputMode = true
			}
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	var b strings.Builder

	width := m.width
	if width == 0 {
		width = 80
	}

	verticalPadding := "\n\n\n\n"

	if m.inputMode {
		b.WriteString(verticalPadding)

		title := titleStyle.Render("DUET")
		b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, title))
		b.WriteString("\n")

		subtitle := subtitleStyle.Render("ssh pair programming")
		b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, subtitle))
		b.WriteString("\n\n")

		divider := dividerStyle.Render("─────────────────────────")
		b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, divider))
		b.WriteString("\n\n")

		label := inputLabelStyle.Render("Room ID")
		b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, label))
		b.WriteString("\n")

		var inputContent string
		if len(m.input) == 0 {
			inputContent = placeholderStyle.Render("paste or type") + cursorStyle.Render(" █")
		} else {
			inputContent = inputTextStyle.Render(m.input) + cursorStyle.Render("█")
		}

		inputBox := inputBoxStyle.Render(inputContent)
		b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, inputBox))
		b.WriteString("\n\n")

		keys := []string{
			keyStyle.Render("↵") + " " + helpStyle.Render("join"),
			keyStyle.Render("esc") + " " + helpStyle.Render("back"),
			keyStyle.Render("^U") + " " + helpStyle.Render("clear"),
		}
		separator := lipgloss.NewStyle().Foreground(dimColor).Render("·")
		help := strings.Join(keys, "  "+separator+"  ")
		b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, help))

		return b.String()
	}

	b.WriteString(verticalPadding)

	logo := titleStyle.Render("DUET")
	b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, logo))
	b.WriteString("\n")

	tagline := subtitleStyle.Render("ssh pair programming")
	b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, tagline))
	b.WriteString("\n\n")

	// Minimal divider
	divider := dividerStyle.Render("─────────────────────────")
	b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, divider))
	b.WriteString("\n\n")

	// Menu options - super clean
	var option1, option2 string

	if m.choice == 0 {
		option1 = selectedStyle.Render("→ Create Session")
		option2 = optionStyle.Render("  Join Session")
	} else {
		option1 = optionStyle.Render("  Create Session")
		option2 = selectedStyle.Render("→ Join Session")
	}

	b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, option1))
	b.WriteString("\n")
	b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, option2))
	b.WriteString("\n\n")

	// Minimal keybinds
	keys := []string{
		keyStyle.Render("↑↓") + " " + helpStyle.Render("select"),
		keyStyle.Render("↵") + " " + helpStyle.Render("confirm"),
		keyStyle.Render("q") + " " + helpStyle.Render("quit"),
	}
	separator := lipgloss.NewStyle().Foreground(dimColor).Render("·")
	help := strings.Join(keys, "  "+separator+"  ")
	b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, help))

	return b.String()
}

func (m MenuModel) GetChoice() MenuChoice {
	return m.selected
}

func (m MenuModel) GetRoomID() string {
	return m.roomID
}
