package types

// RenderNodeKind identifies a structured intermediate render node.
type RenderNodeKind string

const (
	RenderNodeRoot    RenderNodeKind = "root"
	RenderNodeSection RenderNodeKind = "section"
	RenderNodeLine    RenderNodeKind = "line"
	RenderNodeTable   RenderNodeKind = "table"
	RenderNodeRow     RenderNodeKind = "row"
)

// RenderNode is a lightweight render-tree node consumed by the TUI pipeline.
type RenderNode struct {
	Kind     RenderNodeKind `json:"kind"`
	Text     string         `json:"text,omitempty"`
	Cells    []string       `json:"cells,omitempty"`
	Children []RenderNode   `json:"children,omitempty"`
}

// WithChildren returns a node copy with appended child nodes.
func (n RenderNode) WithChildren(children ...RenderNode) RenderNode {
	if len(children) == 0 {
		return n
	}
	n.Children = append(append([]RenderNode(nil), n.Children...), children...)
	return n
}
