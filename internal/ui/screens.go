package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

const defaultWidth = 80

func centerLine(s ssh.Session, txt string) {
	fmt.Fprintf(s, "%s\n", lipgloss.PlaceHorizontal(defaultWidth, lipgloss.Center, txt))
}

// ShowError renders a simple error message to the SSH session.
func ShowError(s ssh.Session, msg string) {
	fmt.Fprintf(s, "\033[2J\033[H")
	fmt.Fprintf(s, "\n\n")
	centerLine(s, lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6666")).Bold(true).Render("❌ "+msg))
	fmt.Fprintf(s, "\n")
}

// ShowRoomCreated renders the room created banner.
func ShowRoomCreated(s ssh.Session, roomID string) {
	fmt.Fprintf(s, "\033[2J\033[H")
	fmt.Fprintf(s, "\n\n")
	centerLine(s, titleStyle.Render("DUET"))
	centerLine(s, subtitleStyle.Render("ssh pair programming"))
	centerLine(s, dividerStyle.Render("─────────────────────────"))
	centerLine(s, selectedStyle.Render("Room Created"))
	fmt.Fprintf(s, "\n")
	centerLine(s, subtitleStyle.Render("Share this ID to pair:"))
	centerLine(s, "┌──────────────┐")
	centerLine(s, fmt.Sprintf("│ %s │", selectedStyle.Render(roomID)))
	centerLine(s, "└──────────────┘")
	fmt.Fprintf(s, "\n")
	centerLine(s, subtitleStyle.Render("Starting shared terminal in 3s..."))
	centerLine(s, helpStyle.Render("Disconnect to exit"))
	time.Sleep(3 * time.Second)
}

// ShowJoining renders the joining banner.
func ShowJoining(s ssh.Session, roomID string) {
	fmt.Fprintf(s, "\033[2J\033[H")
	fmt.Fprintf(s, "\n\n")
	centerLine(s, titleStyle.Render("DUET"))
	centerLine(s, subtitleStyle.Render("ssh pair programming"))
	centerLine(s, dividerStyle.Render("─────────────────────────"))
	centerLine(s, selectedStyle.Render("Joining Room"))
	fmt.Fprintf(s, "\n")
	centerLine(s, selectedStyle.Render(roomID))
	centerLine(s, subtitleStyle.Render("Connecting to shared terminal..."))
	centerLine(s, helpStyle.Render("Disconnect to exit"))
	time.Sleep(2 * time.Second)
}
