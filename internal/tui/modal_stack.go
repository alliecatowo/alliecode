package tui

import "strings"

type modalSurface string

const (
	modalSurfaceNone         modalSurface = "none"
	modalSurfaceReference    modalSurface = "reference"
	modalSurfaceSlash        modalSurface = "slash"
	modalSurfaceModelPicker  modalSurface = "model-picker"
	modalSurfaceCommandPanel modalSurface = "command-panel"
	modalSurfaceSearch       modalSurface = "search"
	modalSurfacePermission   modalSurface = "permission"
)

type modalLayer struct {
	surface  modalSurface
	priority int
}

type modalStackState struct {
	layers []modalLayer
}

func (a *App) syncModalStack() {
	layers := make([]modalLayer, 0, 6)
	if a.refAuto.active {
		layers = append(layers, modalLayer{surface: modalSurfaceReference, priority: 10})
	}
	if a.slashAutocomplete.isVisible() {
		layers = append(layers, modalLayer{surface: modalSurfaceSlash, priority: 20})
	}
	if a.modelPickerActive() {
		layers = append(layers, modalLayer{surface: modalSurfaceModelPicker, priority: 30})
	}
	if a.commandPanel.active {
		layers = append(layers, modalLayer{surface: modalSurfaceCommandPanel, priority: 40})
	}
	if a.state == stateSearch {
		layers = append(layers, modalLayer{surface: modalSurfaceSearch, priority: 50})
	}
	if a.state == statePermissionPrompt {
		layers = append(layers, modalLayer{surface: modalSurfacePermission, priority: 60})
	}
	a.modalStack.layers = layers
}

func (a *App) activeModalSurface() modalSurface {
	a.syncModalStack()
	if len(a.modalStack.layers) == 0 {
		return modalSurfaceNone
	}
	top := a.modalStack.layers[0]
	for _, layer := range a.modalStack.layers[1:] {
		if layer.priority >= top.priority {
			top = layer
		}
	}
	return top.surface
}

func (a *App) modalStackSignature() string {
	a.syncModalStack()
	if len(a.modalStack.layers) == 0 {
		return string(modalSurfaceNone)
	}
	parts := make([]string, 0, len(a.modalStack.layers))
	for _, layer := range a.modalStack.layers {
		parts = append(parts, string(layer.surface))
	}
	return strings.Join(parts, ">")
}
