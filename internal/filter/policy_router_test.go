package filter

import "testing"

func TestPolicyRouterRoute_ExplicitIntentWins(t *testing.T) {
	r := NewPolicyRouter()
	route := r.Route("panic: crash", "review")
	if route.QueryIntent != "review" {
		t.Fatalf("got %q, want review", route.QueryIntent)
	}
}

func TestPolicyRouterRoute_InferDebug(t *testing.T) {
	r := NewPolicyRouter()
	route := r.Route("ERROR: build failed with panic: nil pointer", "")
	if route.QueryIntent != "debug" {
		t.Fatalf("got %q, want debug", route.QueryIntent)
	}
}

func TestPolicyRouterRoute_InferDeploy(t *testing.T) {
	r := NewPolicyRouter()
	route := r.Route("kubectl rollout restart deployment api", "")
	if route.QueryIntent != "deploy" {
		t.Fatalf("got %q, want deploy", route.QueryIntent)
	}
}

func TestPolicyRouterRoute_InferReview(t *testing.T) {
	r := NewPolicyRouter()
	route := r.Route("diff --git a/main.go b/main.go", "")
	if route.QueryIntent != "review" {
		t.Fatalf("got %q, want review", route.QueryIntent)
	}
}

func TestPolicyRouterRoute_InferBuild(t *testing.T) {
	r := NewPolicyRouter()
	route := r.Route("build failed during compile and linker stage", "")
	if route.QueryIntent != "build" {
		t.Fatalf("got %q, want build", route.QueryIntent)
	}
}

func TestPolicyRouterRoute_InferSearch(t *testing.T) {
	r := NewPolicyRouter()
	route := r.Route("rg \"NewPipelineCoordinator\" internal/filter", "")
	if route.QueryIntent != "search" {
		t.Fatalf("got %q, want search", route.QueryIntent)
	}
}
