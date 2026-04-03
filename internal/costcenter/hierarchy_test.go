package costcenter

import (
	"strings"
	"testing"
)

func TestNewCostCenterHierarchy(t *testing.T) {
	h := NewCostCenterHierarchy()
	if h == nil {
		t.Fatal("expected non-nil hierarchy")
	}
	if h.centers == nil {
		t.Fatal("expected initialized centers map")
	}
}

func TestAddCenter(t *testing.T) {
	h := NewCostCenterHierarchy()
	center := &CostCenter{ID: "cc-1", Name: "Engineering", Budget: 1000}
	if err := h.AddCenter(center); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := h.GetCenter("cc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Engineering" {
		t.Errorf("name = %q, want %q", got.Name, "Engineering")
	}
}

func TestAddCenterAutoID(t *testing.T) {
	h := NewCostCenterHierarchy()
	center := &CostCenter{Name: "AutoID"}
	if err := h.AddCenter(center); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if center.ID == "" {
		t.Error("expected auto-generated ID")
	}
	if !strings.HasPrefix(center.ID, "cc-") {
		t.Errorf("ID = %q, want cc- prefix", center.ID)
	}
}

func TestAddCenterNilTags(t *testing.T) {
	h := NewCostCenterHierarchy()
	center := &CostCenter{ID: "cc-1"}
	if err := h.AddCenter(center); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if center.Tags == nil {
		t.Error("expected Tags to be initialized")
	}
}

func TestGetCenterNotFound(t *testing.T) {
	h := NewCostCenterHierarchy()
	_, err := h.GetCenter("missing")
	if err == nil {
		t.Error("expected error for missing center")
	}
}

func TestGetChildren(t *testing.T) {
	h := NewCostCenterHierarchy()
	parent := &CostCenter{ID: "cc-parent", Name: "Parent"}
	child := &CostCenter{ID: "cc-child", Name: "Child", ParentID: "cc-parent"}
	h.AddCenter(parent)
	h.AddCenter(child)

	children := h.GetChildren("cc-parent")
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0].Name != "Child" {
		t.Errorf("child name = %q, want %q", children[0].Name, "Child")
	}
}

func TestGetChildrenNotFound(t *testing.T) {
	h := NewCostCenterHierarchy()
	children := h.GetChildren("missing")
	if children != nil {
		t.Errorf("expected nil, got %v", children)
	}
}

func TestGetRootCenters(t *testing.T) {
	h := NewCostCenterHierarchy()
	root := &CostCenter{ID: "cc-root", Name: "Root"}
	child := &CostCenter{ID: "cc-child", Name: "Child", ParentID: "cc-root"}
	h.AddCenter(root)
	h.AddCenter(child)

	roots := h.GetRootCenters()
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].Name != "Root" {
		t.Errorf("root name = %q, want %q", roots[0].Name, "Root")
	}
}

func TestUpdateSpend(t *testing.T) {
	h := NewCostCenterHierarchy()
	center := &CostCenter{ID: "cc-1", Name: "Eng", Budget: 1000, ActualSpend: 100}
	h.AddCenter(center)

	if err := h.UpdateSpend("cc-1", 50); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := h.GetCenter("cc-1")
	if got.ActualSpend != 150 {
		t.Errorf("ActualSpend = %v, want 150", got.ActualSpend)
	}
}

func TestUpdateSpendNotFound(t *testing.T) {
	h := NewCostCenterHierarchy()
	err := h.UpdateSpend("missing", 50)
	if err == nil {
		t.Error("expected error for missing center")
	}
}

func TestUpdateSpendCascadesToParent(t *testing.T) {
	h := NewCostCenterHierarchy()
	parent := &CostCenter{ID: "cc-parent", Name: "Parent", Budget: 5000}
	child := &CostCenter{ID: "cc-child", Name: "Child", ParentID: "cc-parent", Budget: 1000}
	h.AddCenter(parent)
	h.AddCenter(child)

	if err := h.UpdateSpend("cc-child", 200); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parentCenter, _ := h.GetCenter("cc-parent")
	if parentCenter.ActualSpend != 200 {
		t.Errorf("parent ActualSpend = %v, want 200", parentCenter.ActualSpend)
	}
}

func TestGetBudgetUtilization(t *testing.T) {
	h := NewCostCenterHierarchy()
	center := &CostCenter{ID: "cc-1", Name: "Eng", Budget: 1000, ActualSpend: 250}
	h.AddCenter(center)

	util, err := h.GetBudgetUtilization("cc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if util != 25.0 {
		t.Errorf("utilization = %v, want 25.0", util)
	}
}

func TestGetBudgetUtilizationNotFound(t *testing.T) {
	h := NewCostCenterHierarchy()
	_, err := h.GetBudgetUtilization("missing")
	if err == nil {
		t.Error("expected error for missing center")
	}
}

func TestGetBudgetUtilizationZeroBudget(t *testing.T) {
	h := NewCostCenterHierarchy()
	center := &CostCenter{ID: "cc-1", Budget: 0}
	h.AddCenter(center)

	util, err := h.GetBudgetUtilization("cc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if util != 0 {
		t.Errorf("utilization = %v, want 0", util)
	}
}

func TestGetHierarchyReport(t *testing.T) {
	h := NewCostCenterHierarchy()
	h.AddCenter(&CostCenter{ID: "cc-1", Name: "Eng", Budget: 1000, ActualSpend: 500})
	h.AddCenter(&CostCenter{ID: "cc-2", Name: "QA", ParentID: "cc-1", Budget: 500, ActualSpend: 100})

	reports := h.GetHierarchyReport()
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}

	var engReport *HierarchyReport
	for i := range reports {
		if reports[i].ID == "cc-1" {
			engReport = &reports[i]
		}
	}
	if engReport == nil {
		t.Fatal("expected Eng report")
	}
	if engReport.Utilization != 50.0 {
		t.Errorf("Eng utilization = %v, want 50.0", engReport.Utilization)
	}
	if engReport.ChildrenCount != 1 {
		t.Errorf("Eng children = %d, want 1", engReport.ChildrenCount)
	}
}
