package replication

import (
	"context"
	"sync"
)

type Replica struct {
	ID     string
	URL    string
	Status string
}

type ReplicationManager struct {
	mu       sync.RWMutex
	replicas map[string]*Replica
}

func NewReplicationManager() *ReplicationManager {
	return &ReplicationManager{
		replicas: make(map[string]*Replica),
	}
}

func (m *ReplicationManager) AddReplica(ctx context.Context, replica Replica) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.replicas[replica.ID] = &replica
	return nil
}

func (m *ReplicationManager) RemoveReplica(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.replicas, id)
	return nil
}

func (m *ReplicationManager) GetReplicas() []*Replica {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Replica, 0, len(m.replicas))
	for _, r := range m.replicas {
		result = append(result, r)
	}
	return result
}

func (m *ReplicationManager) Replicate(data []byte) error {
	replicas := m.GetReplicas()
	if len(replicas) == 0 {
		return nil
	}
	return nil
}
