package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/jaypopat/duet/internal/pty"
	"github.com/jaypopat/duet/internal/room"
	"github.com/jaypopat/duet/internal/ui"
)

type Server struct {
	addr        string
	hostKeyPath string
	roomManager *room.Manager
	logger      *log.Logger
}

func New(addr, hostKeyPath string) *Server {
	return &Server{
		addr:        addr,
		hostKeyPath: hostKeyPath,
		roomManager: room.NewManager(),
		logger: log.NewWithOptions(os.Stderr, log.Options{
			Prefix: "duet",
		}),
	}
}

func (s *Server) Start() error {
	srv, err := wish.NewServer(
		wish.WithAddress(s.addr),
		wish.WithHostKeyPath(s.hostKeyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(s.teaHandler),
			logging.Middleware(),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.ListenAndServe()
	}()

	s.logger.Info("Starting SSH server", "address", s.addr)

	select {
	case err := <-serveErr:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	s.logger.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("Server stopped")
	return nil
}

func (s *Server) teaHandler(sshSession ssh.Session) (tea.Model, []tea.ProgramOption) {
	fmt.Fprintf(sshSession, "\033[2J\033[H\033[?25l")

	for {
		menuModel, err := ui.RunMenu(sshSession)
		if err != nil {
			s.logger.Error("Menu error", "error", err)
			return ui.NewMenuModel(), []tea.ProgramOption{}
		}

		switch menuModel.GetChoice() {
		case ui.ChoiceCreate:
			_, _ = s.handleCreateRoom(sshSession)
			continue
		case ui.ChoiceJoin:
			_, _ = s.handleJoinRoom(sshSession, menuModel.GetRoomID())
			continue
		default:
			return ui.NewMenuModel(), []tea.ProgramOption{}
		}
	}
}

func (s *Server) handleCreateRoom(sshSession ssh.Session) (tea.Model, []tea.ProgramOption) {
	hostID := sshSession.User()

	r, err := s.roomManager.CreateRoom(hostID)
	if err != nil {
		s.logger.Error("Failed to create room", "error", err)
		ui.ShowError(sshSession, fmt.Sprintf("Failed to create room: %v", err))
		time.Sleep(2 * time.Second)
		return ui.NewMenuModel(), []tea.ProgramOption{}
	}

	s.logger.Info("Room created", "roomID", r.ID, "host", hostID)

	client := &room.Client{
		ID:      sshSession.User() + "-host",
		Session: sshSession,
		IsHost:  true,
	}
	r.AddClient(client)

	ptyHandler, err := pty.NewHandler(r)
	if err != nil {
		s.logger.Error("Failed to create PTY handler", "error", err)
		ui.ShowError(sshSession, fmt.Sprintf("Failed to create PTY handler: %v", err))
		time.Sleep(2 * time.Second)
		return ui.NewMenuModel(), []tea.ProgramOption{}
	}

	if err := ptyHandler.StartMaster(); err != nil {
		s.logger.Error("Failed to start PTY master", "error", err)
		ui.ShowError(sshSession, fmt.Sprintf("Failed to start shell: %v", err))
		time.Sleep(2 * time.Second)
		return ui.NewMenuModel(), []tea.ProgramOption{}
	}
	r.PTYHandler = ptyHandler

	ui.ShowRoomCreated(sshSession, r.ID)

	err = ui.HandleSession(sshSession, r, client, ptyHandler, true)
	if err != nil {
		s.logger.Error("Host session error", "error", err)
	}

	r.RemoveClient(client.ID)
	if r.ClientCount() == 0 {
		ptyHandler.Close()
		s.roomManager.DeleteRoom(r.ID)
		s.logger.Info("Room closed", "roomID", r.ID)
	}

	return ui.NewMenuModel(), []tea.ProgramOption{}
}

func (s *Server) handleJoinRoom(sshSession ssh.Session, roomID string) (tea.Model, []tea.ProgramOption) {
	if roomID == "" {
		ui.ShowError(sshSession, "No room ID provided")
		return ui.NewMenuModel(), []tea.ProgramOption{}
	}

	r, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		s.logger.Error("Room not found", "roomID", roomID)
		ui.ShowError(sshSession, fmt.Sprintf("Room %s not found", roomID))
		time.Sleep(2 * time.Second)
		return ui.NewMenuModel(), []tea.ProgramOption{}
	}

	s.logger.Info("Client joining room", "roomID", roomID, "user", sshSession.User())

	client := &room.Client{
		ID:      sshSession.User() + "-guest",
		Session: sshSession,
		IsHost:  false,
	}
	r.AddClient(client)

	var ptyHandler *pty.Handler
	if r.PTYHandler != nil {
		ptyHandler = r.PTYHandler.(*pty.Handler)
	} else {
		s.logger.Error("No PTY handler in room", "roomID", roomID)
		ui.ShowError(sshSession, "Room not ready yet (host hasn't started PTY)")
		time.Sleep(2 * time.Second)
		return ui.NewMenuModel(), []tea.ProgramOption{}
	}

	ui.ShowJoining(sshSession, roomID)

	err = ui.HandleSession(sshSession, r, client, ptyHandler, false)
	if err != nil {
		s.logger.Error("Guest session error", "error", err)
	}

	r.RemoveClient(client.ID)
	s.logger.Info("Client disconnected", "roomID", r.ID, "clientID", client.ID)

	return ui.NewMenuModel(), []tea.ProgramOption{}
}

func (s *Server) GetRoomManager() *room.Manager {
	return s.roomManager
}
