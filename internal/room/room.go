package room

import (
	"os"
	"sync"

	"github.com/charmbracelet/ssh"
)

// represents a connected user
type Client struct {
	ID      string
	Session ssh.Session
	IsHost  bool
}

// represents a pairing session
type Room struct {
	ID          string
	Host        string
	Connections []*Client
	PTYSession  *os.File
	PTYHandler  any
	MasterPath  string
	mu          sync.RWMutex
}

func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Connections = append(r.Connections, client)
}

func (r *Room) RemoveClient(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, c := range r.Connections {
		if c.ID == clientID {
			r.Connections = append(r.Connections[:i], r.Connections[i+1:]...)
			break
		}
	}
}

func (r *Room) GetClients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	clients := make([]*Client, len(r.Connections))
	copy(clients, r.Connections)
	return clients
}

func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Connections)
}
