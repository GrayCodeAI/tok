package replication

import (
	"testing"
)

func TestNewReplicationManager(t *testing.T) {
	mgr := NewReplicationManager()
	if mgr == nil {
		t.Error("Expected non-nil manager")
	}
}

func TestReplicationManagerAddReplica(t *testing.T) {
	mgr := NewReplicationManager()

	replica := Replica{ID: "r1", URL: "http://localhost:8080", Status: "active"}
	err := mgr.AddReplica(nil, replica)
	if err != nil {
		t.Errorf("AddReplica failed: %v", err)
	}

	replicas := mgr.GetReplicas()
	if len(replicas) != 1 {
		t.Errorf("Expected 1 replica, got %d", len(replicas))
	}
}

func TestReplicationManagerRemoveReplica(t *testing.T) {
	mgr := NewReplicationManager()

	mgr.AddReplica(nil, Replica{ID: "r1", URL: "http://localhost:8080", Status: "active"})
	mgr.RemoveReplica(nil, "r1")

	replicas := mgr.GetReplicas()
	if len(replicas) != 0 {
		t.Errorf("Expected 0 replicas, got %d", len(replicas))
	}
}

func TestReplicationManagerReplicate(t *testing.T) {
	mgr := NewReplicationManager()

	err := mgr.Replicate([]byte("test data"))
	if err != nil {
		t.Errorf("Replicate failed: %v", err)
	}
}
