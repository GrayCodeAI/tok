// Package costcenter provides cost center hierarchy management
package costcenter

import (
	"fmt"
	"sync"
	"time"
)

// CostCenter represents a cost center in the hierarchy
type CostCenter struct {
	ID          string
	Name        string
	Description string
	ParentID    string
	Children    []string
	Budget      float64
	ActualSpend float64
	Owner       string
	Tags        map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CostCenterHierarchy manages the cost center hierarchy
type CostCenterHierarchy struct {
	centers map[string]*CostCenter
	mu      sync.RWMutex
}

// NewCostCenterHierarchy creates a new hierarchy
func NewCostCenterHierarchy() *CostCenterHierarchy {
	return &CostCenterHierarchy{
		centers: make(map[string]*CostCenter),
	}
}

// AddCenter adds a cost center
func (h *CostCenterHierarchy) AddCenter(center *CostCenter) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if center.ID == "" {
		center.ID = generateID()
	}

	if center.Tags == nil {
		center.Tags = make(map[string]string)
	}

	h.centers[center.ID] = center

	// Add to parent's children
	if center.ParentID != "" {
		parent, ok := h.centers[center.ParentID]
		if ok {
			parent.Children = append(parent.Children, center.ID)
		}
	}

	return nil
}

// GetCenter returns a cost center by ID
func (h *CostCenterHierarchy) GetCenter(id string) (*CostCenter, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	center, ok := h.centers[id]
	if !ok {
		return nil, fmt.Errorf("cost center not found: %s", id)
	}

	return center, nil
}

// GetChildren returns all children of a cost center
func (h *CostCenterHierarchy) GetChildren(id string) []*CostCenter {
	h.mu.RLock()
	defer h.mu.RUnlock()

	center, ok := h.centers[id]
	if !ok {
		return nil
	}

	children := make([]*CostCenter, 0, len(center.Children))
	for _, childID := range center.Children {
		if child, ok := h.centers[childID]; ok {
			children = append(children, child)
		}
	}

	return children
}

// GetRootCenters returns all root cost centers
func (h *CostCenterHierarchy) GetRootCenters() []*CostCenter {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roots := make([]*CostCenter, 0)
	for _, center := range h.centers {
		if center.ParentID == "" {
			roots = append(roots, center)
		}
	}

	return roots
}

// UpdateSpend updates the actual spend for a cost center
func (h *CostCenterHierarchy) UpdateSpend(id string, amount float64) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	center, ok := h.centers[id]
	if !ok {
		return fmt.Errorf("cost center not found: %s", id)
	}

	center.ActualSpend += amount
	center.UpdatedAt = time.Now()

	// Update parent spend recursively
	h.updateParentSpend(center.ParentID, amount)

	return nil
}

func (h *CostCenterHierarchy) updateParentSpend(parentID string, amount float64) {
	if parentID == "" {
		return
	}

	parent, ok := h.centers[parentID]
	if !ok {
		return
	}

	parent.ActualSpend += amount
	parent.UpdatedAt = time.Now()

	h.updateParentSpend(parent.ParentID, amount)
}

// GetBudgetUtilization returns budget utilization for a cost center
func (h *CostCenterHierarchy) GetBudgetUtilization(id string) (float64, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	center, ok := h.centers[id]
	if !ok {
		return 0, fmt.Errorf("cost center not found: %s", id)
	}

	if center.Budget == 0 {
		return 0, nil
	}

	return (center.ActualSpend / center.Budget) * 100, nil
}

// GetHierarchyReport returns a report of the entire hierarchy
func (h *CostCenterHierarchy) GetHierarchyReport() []HierarchyReport {
	h.mu.RLock()
	defer h.mu.RUnlock()

	reports := make([]HierarchyReport, 0)

	for _, center := range h.centers {
		utilization := 0.0
		if center.Budget > 0 {
			utilization = (center.ActualSpend / center.Budget) * 100
		}

		reports = append(reports, HierarchyReport{
			ID:            center.ID,
			Name:          center.Name,
			ParentID:      center.ParentID,
			Budget:        center.Budget,
			ActualSpend:   center.ActualSpend,
			Utilization:   utilization,
			ChildrenCount: len(center.Children),
			Owner:         center.Owner,
		})
	}

	return reports
}

// HierarchyReport represents a cost center report
type HierarchyReport struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ParentID      string  `json:"parent_id"`
	Budget        float64 `json:"budget"`
	ActualSpend   float64 `json:"actual_spend"`
	Utilization   float64 `json:"utilization"`
	ChildrenCount int     `json:"children_count"`
	Owner         string  `json:"owner"`
}

func generateID() string {
	return fmt.Sprintf("cc-%d", time.Now().UnixNano())
}
