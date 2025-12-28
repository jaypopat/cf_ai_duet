package room

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrRoomNotFound = errors.New("room not found")
)

type Manager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

func (m *Manager) CreateRoom(host, description string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	roomID := uuid.New().String()
	room := &Room{
		ID:          roomID,
		Description: description,
		Host:        host,
		Connections: make([]*Client, 0),
	}
	m.rooms[roomID] = room
	return room, nil
}
func (m *Manager) GetRoom(roomID string) (*Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, exists := m.rooms[roomID]
	if !exists {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

func (m *Manager) RoomCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

func (m *Manager) LeaveRoom(roomID, clientID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, exists := m.rooms[roomID]
	if !exists {
		return false
	}

	room.RemoveClient(clientID)

	if room.ClientCount() == 0 {
		if room.Terminal != nil {
			room.Terminal.Close()
			room.Terminal = nil
		}
		delete(m.rooms, roomID)
		return true
	}
	return false
}
