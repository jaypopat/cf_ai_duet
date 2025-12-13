package ui

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/jaypopat/duet/internal/pty"
	"github.com/jaypopat/duet/internal/room"
)

type SessionModel struct {
	room       *room.Room
	client     *room.Client
	ptyHandler *pty.Handler
	isHost     bool
}

func NewSessionModel(r *room.Room, c *room.Client, h *pty.Handler, isHost bool) SessionModel {
	return SessionModel{
		room:       r,
		client:     c,
		ptyHandler: h,
		isHost:     isHost,
	}

}

func (m SessionModel) Init() tea.Cmd {
	return nil
}

func (m SessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, nil
	}

	return m, nil
}

func (m SessionModel) View() string {
	if m.isHost {
		return fmt.Sprintf("Duet Session - Room ID: %s (Host)\nClients: %d\n\nDisconnect to exit",
			m.room.ID,
			m.room.ClientCount())
	}
	return fmt.Sprintf("Duet Session - Room ID: %s (Guest)\n\nDisconnect to exit",
		m.room.ID)
}

type ptyClientWriter struct {
	h        *pty.Handler
	clientID string
}

func (w ptyClientWriter) Write(p []byte) (int, error) {
	if w.h == nil {
		return len(p), nil
	}
	if err := w.h.WriteFromClient(w.clientID, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func HandleSession(s ssh.Session, r *room.Room, c *room.Client, h *pty.Handler, isHost bool) error {
	fmt.Fprintf(s, "\r\n[Duet] Room: %s | You: %s\r\n", r.ID, c.ID)

	if isHost && h != nil {
		go h.BroadcastToClients()
	}

	_, err := io.Copy(ptyClientWriter{h: h, clientID: c.ID}, s)
	return err
}
