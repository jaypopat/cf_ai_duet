package pty

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/jaypopat/duet/internal/room"
)

// manages PTY sessions and broadcasting
type Handler struct {
	room       *room.Room
	ptmx       *os.File
	mu         sync.Mutex
	lastWriter string
}

// creates a new PTY handler
func NewHandler(r *room.Room) (*Handler, error) {
	return &Handler{
		room: r,
	}, nil
}

// starts the master PTY with a shell
func (h *Handler) StartMaster() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var cmd *exec.Cmd

	if shell := os.Getenv("SHELL"); shell != "" {
		cmd = exec.Command(shell)
	} else {
		cmd = exec.Command("/bin/sh")
	}
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	h.ptmx = ptmx
	h.room.MasterPath = ptmx.Name()

	return nil
}

// broadcasts PTY output to all connected clients
func (h *Handler) BroadcastToClients() {
	h.mu.Lock()
	ptmx := h.ptmx
	h.mu.Unlock()
	if ptmx == nil {
		return
	}

	buf := make([]byte, 1024)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		clients := h.room.GetClients()
		for _, client := range clients {
			if client.Session != nil {
				client.Session.Write(buf[:n])
			}
		}
	}
}

// writes input from a client to the master PTY
func (h *Handler) WriteFromClient(clientID string, data []byte) error {
	h.mu.Lock()
	ptmx := h.ptmx
	roomID := h.room.ID
	shouldUpdateTitle := clientID != "" && clientID != h.lastWriter
	if shouldUpdateTitle {
		h.lastWriter = clientID
	}
	h.mu.Unlock()

	if ptmx == nil {
		return nil
	}

	if shouldUpdateTitle {
		for _, c := range h.room.GetClients() {
			if c.Session != nil {
				fmt.Fprintf(c.Session, "\033]0;Duet %s - %s typing\007", roomID, clientID)
			}
		}
	}

	_, err := ptmx.Write(data)
	return err
}

// closes the PTY
func (h *Handler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.ptmx != nil {
		err := h.ptmx.Close()
		h.ptmx = nil
		return err
	}
	return nil
}
