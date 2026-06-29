package relationships_test

import (
	"context"
	"testing"
)

func TestGraph_SymmetricType_OneLink(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")

	rt, err := svc.CreateType(ctx, "Friend", "Friend")
	if err != nil {
		t.Fatalf("CreateType: %v", err)
	}

	if _, err := svc.AttachRelationship(ctx, alice, bob, rt.ID, ""); err != nil {
		t.Fatalf("AttachRelationship: %v", err)
	}

	graph, err := svc.Graph(ctx, 0)
	if err != nil {
		t.Fatalf("Graph: %v", err)
	}

	if len(graph.Links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(graph.Links))
	}

	link := graph.Links[0]

	if link.Source > link.Target {
		t.Errorf("source should be <= target; got source=%d target=%d", link.Source, link.Target)
	}

	if link.Type != "Friend" {
		t.Errorf("type: got %q, want %q", link.Type, "Friend")
	}

	if link.ReverseType != "Friend" {
		t.Errorf("reverse_type: got %q, want %q", link.ReverseType, "Friend")
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}
}

func TestGraph_AsymmetricType_OneLinkCorrectNames(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")

	rt, err := svc.CreateType(ctx, "Manager", "Reports to")
	if err != nil {
		t.Fatalf("CreateType: %v", err)
	}

	// Alice manages Bob.
	if _, err := svc.AttachRelationship(ctx, alice, bob, rt.ID, ""); err != nil {
		t.Fatalf("AttachRelationship: %v", err)
	}

	graph, err := svc.Graph(ctx, 0)
	if err != nil {
		t.Fatalf("Graph: %v", err)
	}

	if len(graph.Links) != 1 {
		t.Fatalf("expected exactly 1 link after dedup, got %d", len(graph.Links))
	}

	link := graph.Links[0]

	// Canonical: source = min(alice,bob), target = max(alice,bob).
	src, tgt := alice, bob
	if src > tgt {
		src, tgt = tgt, src
	}

	if link.Source != src || link.Target != tgt {
		t.Errorf("link endpoints: got {%d,%d}, want {%d,%d}", link.Source, link.Target, src, tgt)
	}

	// When alice < bob: type = "Manager" (alice→bob direction), reverse = "Reports to".
	// When alice > bob: type = "Reports to" (bob→alice direction) — ensure not both are equal.
	if link.Type == "" || link.ReverseType == "" {
		t.Errorf("type or reverse_type is empty: type=%q reverse_type=%q", link.Type, link.ReverseType)
	}

	// The two names must be opposites.
	if link.Type == link.ReverseType {
		t.Errorf("asymmetric type should have different type/reverse_type, got both %q", link.Type)
	}

	validPairs := (link.Type == "Manager" && link.ReverseType == "Reports to") ||
		(link.Type == "Reports to" && link.ReverseType == "Manager")

	if !validPairs {
		t.Errorf("unexpected type pair: type=%q reverse_type=%q", link.Type, link.ReverseType)
	}
}

func TestGraph_EgoMode_Triangle(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")
	carol := insertPerson(t, db, "Carol")

	rt, err := svc.CreateType(ctx, "Friend", "Friend")
	if err != nil {
		t.Fatalf("CreateType: %v", err)
	}

	// Alice—Bob, Alice—Carol, Bob—Carol.
	if _, err := svc.AttachRelationship(ctx, alice, bob, rt.ID, ""); err != nil {
		t.Fatalf("attach alice-bob: %v", err)
	}

	if _, err := svc.AttachRelationship(ctx, alice, carol, rt.ID, ""); err != nil {
		t.Fatalf("attach alice-carol: %v", err)
	}

	if _, err := svc.AttachRelationship(ctx, bob, carol, rt.ID, ""); err != nil {
		t.Fatalf("attach bob-carol: %v", err)
	}

	graph, err := svc.Graph(ctx, alice)
	if err != nil {
		t.Fatalf("Graph ego: %v", err)
	}

	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Links) != 3 {
		t.Errorf("expected 3 links (triangle), got %d", len(graph.Links))
	}
}

func TestGraph_UnlabeledPerson_EmptyGroup(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")

	rt, _ := svc.CreateType(ctx, "Friend", "Friend")
	if _, err := svc.AttachRelationship(ctx, alice, bob, rt.ID, ""); err != nil {
		t.Fatalf("AttachRelationship: %v", err)
	}

	graph, err := svc.Graph(ctx, 0)
	if err != nil {
		t.Fatalf("Graph: %v", err)
	}

	for _, node := range graph.Nodes {
		if node.Group != "" {
			t.Errorf("node %d: expected empty group (no labels), got %q", node.ID, node.Group)
		}
	}
}

func TestGraph_SelfPersonAnchor_GlobalView(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	// Insert a self person with no relationships.
	_, err := db.ExecContext(ctx, `INSERT INTO person (name, is_self) VALUES (?, 1)`, "Me")
	if err != nil {
		t.Fatalf("insert self person: %v", err)
	}

	graph, err := svc.Graph(ctx, 0)
	if err != nil {
		t.Fatalf("Graph: %v", err)
	}

	var found bool

	for _, node := range graph.Nodes {
		if node.IsSelf {
			found = true

			break
		}
	}

	if !found {
		t.Error("expected self person to appear in global graph nodes even with 0 edges")
	}
}
