package tui

import "github.com/alliecatowo/alliecode/internal/types"

func renderIntentsPlain(intents []types.RenderIntent) string {
	return renderNodeTreePlain(buildIntentRenderTree(intents))
}

func renderIntentPlain(intent types.RenderIntent) string {
	return renderNodeTreePlain(types.RenderNode{Kind: types.RenderNodeRoot, Children: []types.RenderNode{buildIntentNode(intent)}})
}
