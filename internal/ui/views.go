package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

func (m *Model) viewLaunch() string {
	logo := m.styles.logoStyle.Render(asciiLogo)

	createBtn := m.styles.buttonStyle.Render("Create Room  (c)")
	joinBtn := m.styles.buttonStyle.Render("Join Room    (J)")

	if m.selected == 0 {
		createBtn = m.styles.buttonActive.Render("Create Room  (c)")
	} else {
		joinBtn = m.styles.buttonActive.Render("Join Room    (J)")
	}

	buttons := lipgloss.JoinVertical(lipgloss.Center, createBtn, joinBtn)
	help := m.styles.helpStyle.Render("↑/↓ select • enter confirm • q quit")
	content := lipgloss.JoinVertical(lipgloss.Center, logo, buttons, help)

	return lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) viewCreate() string {
	title := m.styles.titleStyle.Render("Create Room")
	prompt := m.styles.textStyle.Render("Enter a description for your room:")
	input := m.styles.inputBoxStyle.Render(m.input.View())
	help := m.styles.helpStyle.Render("enter create • esc back")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title, "", prompt, "", input, help,
	)

	view := lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, content)

	return view
}

func (m *Model) viewJoin() string {
	title := m.styles.titleStyle.Render("Join Room")
	prompt := m.styles.textStyle.Render("Enter the room ID:")
	input := m.styles.inputBoxStyle.Render(m.input.View())
	help := m.styles.helpStyle.Render("enter join • esc back")

	// if room doesnt exist we show the toast
	var errorLine string
	if len(m.toasts) > 0 {
		errorLine = m.styles.errorStyle.Render("▸ " + m.toasts[len(m.toasts)-1].text)
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		title, "", prompt, "", input, "", errorLine, help,
	)

	view := lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, content)

	return view
}

func (m *Model) viewRoomCreated() string {
	title := m.styles.titleStyle.Render("Room Created!")

	// Room code box - for easy copying
	codeLabel := m.styles.dimStyle.Render("Share this code with others to join:")
	codeBox := m.styles.baseStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 3).
		Bold(true).
		Foreground(colorSuccess).
		Render(m.roomID)

	hint := m.styles.dimStyle.Render("(select and copy the code above)")
	help := m.styles.helpStyle.Render("enter → enter room • esc back")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title, "", codeLabel, "", codeBox, "", hint, "", help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) viewRoom() string {
	if m.width < MinWidthForSidebar || m.height < MinHeightForSidebar {
		return m.viewResizePrompt()
	}

	sidebarW, terminalW, aiSidebarW, mainHeight := m.roomLayout()

	sidebar := m.renderSidebar(sidebarW, mainHeight)
	terminal := m.renderTerminal(terminalW, mainHeight)

	var main string
	if m.showAISidebar {
		aiPanel := m.renderAISidebar(aiSidebarW, mainHeight)
		main = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, terminal, aiPanel)
	} else {
		main = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, terminal)
	}

	// bottom bar (vim-like): input bar or toasts
	bottom := m.renderBottomBar()
	bottom = m.styles.bottomBarStyle.Width(m.width).Render(bottom)

	return lipgloss.JoinVertical(lipgloss.Left, main, bottom)
}

func (m *Model) renderSidebar(w, h int) string {
	var b strings.Builder

	youLabel := m.styles.dimStyle.Render("you: ")
	youName := m.styles.accentStyle.Bold(true).Render(m.username)
	b.WriteString(youLabel + youName + "\n\n")

	roomLabel := m.styles.dimStyle.Render("room: ")
	roomID := m.styles.textStyle.Render(truncate(m.roomID, w-8))
	b.WriteString(roomLabel + roomID + "\n")

	if m.currentRoom != nil && m.currentRoom.Description != "" {
		desc := m.currentRoom.Description
		if len(desc) > w-4 {
			desc = truncate(desc, w-4)
		}
		descText := m.styles.dimStyle.Render("      " + "\"" + desc + "\"")
		b.WriteString(descText + "\n")
	}
	b.WriteString(m.styles.dimStyle.Render(strings.Repeat("─", w-2)) + "\n\n")

	// Users
	usersLabel := m.styles.dimStyle.Render(fmt.Sprintf("connected (%d):", len(m.users)))
	b.WriteString(usersLabel + "\n")
	for _, u := range m.users {
		b.WriteString(m.styles.textStyle.Render("  • "+u) + "\n")
	}

	// Typing indicator
	if m.typingUser != "" {
		b.WriteString("\n")
		typingText := fmt.Sprintf("✎ %s is typing...", m.typingUser)
		b.WriteString(m.styles.accentStyle.Render(typingText) + "\n")
	}
	b.WriteString(m.styles.dimStyle.Render(strings.Repeat("─", w-2)) + "\n\n")

	// Keybinds
	keysLabel := m.styles.dimStyle.Render("keys:")
	b.WriteString(keysLabel + "\n")
	b.WriteString(m.styles.textStyle.Render("  ctrl+g  AI prompt") + "\n")
	b.WriteString(m.styles.textStyle.Render("  ctrl+a  toggle AI") + "\n")
	b.WriteString(m.styles.textStyle.Render("  ctrl+j/k scroll AI") + "\n")
	b.WriteString(m.styles.textStyle.Render("  ctrl+r  run command") + "\n")
	b.WriteString(m.styles.textStyle.Render("  ctrl+l  leave room") + "\n")

	return m.styles.sidebarStyle.Width(w).Height(h).Render(b.String())
}

func (m *Model) renderTerminal(w, h int) string {
	header := m.styles.titleStyle.Render("shared terminal")
	content := m.termContent
	if content == "" {
		content = m.styles.dimStyle.Render("Starting terminal...")
	}

	return m.styles.terminalStyle.Width(w).Height(h).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, "", content),
	)
}

func (m *Model) renderBottomBar() string {
	// Right side: Mode status (always visible) similar to vim mode indicator
	modeText := m.getModeStatus()
	right := m.styles.accentStyle.Bold(true).Render(modeText)
	rightWidth := lipgloss.Width(right)

	//  Priority: Toasts > Input > Help
	var left string
	if len(m.toasts) > 0 {
		var parts []string
		for _, t := range m.toasts {
			parts = append(parts, t.text)
		}
		toastText := "▸ " + strings.Join(parts, " • ")
		left = m.styles.accentStyle.Bold(true).Render(truncate(toastText, m.width-rightWidth-2))
	} else if m.inputMode != ModeNormal {
		left = m.cmdInput.View()
	} else {
		helpText := "ctrl+g AI • ctrl+a toggle AI • ctrl+r sandbox"
		left = m.styles.dimStyle.Render(truncate(helpText, m.width-rightWidth-2))
	}

	leftWidth := lipgloss.Width(left)
	padding := max(0, m.width-leftWidth-rightWidth)

	return left + strings.Repeat(" ", padding) + right
}

func (m *Model) getModeStatus() string {
	switch m.inputMode {
	case ModeAI:
		return "-- AI --"
	case ModeSandbox:
		return "-- RUN --"
	default:
		return "-- NORMAL --"
	}
}

func (m *Model) viewResizePrompt() string {
	title := m.styles.titleStyle.Render("Terminal Too Small")
	msg := m.styles.textStyle.Render(fmt.Sprintf(
		"Please resize your terminal to at least %dx%d",
		MinWidthForSidebar, MinHeightForSidebar,
	))
	current := m.styles.dimStyle.Render(fmt.Sprintf("Current: %dx%d", m.width, m.height))
	hint := m.styles.dimStyle.Render("(or press ctrl+a to hide AI sidebar)")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title, "", msg, current, "", hint,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) renderAISidebar(w, h int) string {
	var b strings.Builder

	header := m.styles.titleStyle.Render("AI Assistant")
	b.WriteString(header + "\n")
	b.WriteString(m.styles.dimStyle.Render(strings.Repeat("─", w-4)) + "\n\n")

	if m.aiLoading {
		loadingText := fmt.Sprintf("%s Thinking...", m.aiSpinner.View())
		b.WriteString(m.styles.accentStyle.Render(loadingText) + "\n\n")
		b.WriteString(m.aiViewport.View())
	} else if len(m.getAIMessages()) == 0 {
		emptyMsg := m.styles.dimStyle.Render("No messages yet.\nPress ctrl+g to ask AI.")
		b.WriteString(emptyMsg)
	} else {
		b.WriteString(m.aiViewport.View())
	}

	// Scroll indicator (0 -100)
	if len(m.getAIMessages()) > 0 {
		scrollInfo := fmt.Sprintf(" %.0f%% ", m.aiViewport.ScrollPercent()*100)
		b.WriteString("\n" + m.styles.dimStyle.Render(scrollInfo))
	}

	return m.styles.aiSidebarStyle.Width(w).Height(h).Render(b.String())
}

// formatting content for viewport with proper line tracking
func (m *Model) buildAIContent(maxWidth int) (string, int) {
	if maxWidth <= 0 {
		maxWidth = 40
	}

	var b strings.Builder
	wrapWidth := maxWidth - 4 // account for indent

	msgs := m.getAIMessages()
	currentLine := 0
	lastPromptOffset := 0

	for i, msg := range msgs {
		var prefix string
		var isUser bool

		if msg.Role == "user" {
			username := msg.UserID
			if username == "" {
				username = "you"
			}
			prefix = m.styles.accentStyle.Render(username + ": ")
			isUser = true
			// Track the line offset where this user prompt starts
			lastPromptOffset = currentLine
		} else {
			prefix = m.styles.dimStyle.Render("AI: ")
			isUser = false
		}

		// Word wrap the message text using reflow
		wrapped := wordwrap.String(msg.Text, wrapWidth)
		lines := strings.Split(wrapped, "\n")

		for j, line := range lines {
			if j == 0 {
				// First line includes prefix
				if isUser {
					b.WriteString(prefix + m.styles.accentStyle.Render(line))
				} else {
					b.WriteString(prefix + m.styles.textStyle.Render(line))
				}
			} else {
				// Continuation lines - indent to align with text
				indent := "    "
				if isUser {
					b.WriteString(indent + m.styles.accentStyle.Render(line))
				} else {
					b.WriteString(indent + m.styles.textStyle.Render(line))
				}
			}
			b.WriteString("\n")
			currentLine++
		}

		// Blank line between messages (except after last)
		if i < len(msgs)-1 {
			b.WriteString("\n")
			currentLine++
		}
	}

	return b.String(), lastPromptOffset
}
