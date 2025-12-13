package room

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomExists   = errors.New("room already exists")
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

func (m *Manager) CreateRoom(hostPubKey string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	roomID := generateRoomID()

	if _, exists := m.rooms[roomID]; exists {
		return nil, ErrRoomExists
	}

	room := &Room{
		ID:          roomID,
		Host:        hostPubKey,
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

func (m *Manager) DeleteRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, roomID)
}

func (m *Manager) ListRooms() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.rooms))
	for id, _ := range m.rooms {
		ids = append(ids, id)
	}
	return ids
}

func (m *Manager) RoomCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

func generateRoomID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
