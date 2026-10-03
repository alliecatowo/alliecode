package types

import "testing"

func TestRenderNodeWithChildrenAppendsWithoutMutatingOriginal(t *testing.T) {
	original := RenderNode{Kind: RenderNodeSection, Children: []RenderNode{{Kind: RenderNodeLine, Text: "a"}}}
	next := original.WithChildren(RenderNode{Kind: RenderNodeLine, Text: "b"})

	if len(original.Children) != 1 {
		t.Fatalf("expected original node untouched, got %d children", len(original.Children))
	}
	if len(next.Children) != 2 {
		t.Fatalf("expected copied node with appended child, got %d children", len(next.Children))
	}
	if next.Children[1].Text != "b" {
		t.Fatalf("expected appended child text, got %q", next.Children[1].Text)
	}
}
