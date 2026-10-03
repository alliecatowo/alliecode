// Package tui implements AllieCode's terminal user interface using Bubbletea.
package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var allieSignatureSpinner = spinner.Spinner{
	Frames: []string{
		"[A....]",
		"[.A...]",
		"[..A..]",
		"[...A.]",
		"[....A]",
		"[...A.]",
		"[..A..]",
		"[.A...]",
	},
	FPS: time.Millisecond * 90,
}

// SpinnerModel displays an animated thinking/loading indicator with a message.
type SpinnerModel struct {
	spinner spinner.Model
	message string
	active  bool
}

// NewSpinner creates a new spinner with the given status message.
func NewSpinner(message string) SpinnerModel {
	s := spinner.New()
	s.Spinner = allieSignatureSpinner
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("219")).Bold(true)
	return SpinnerModel{
		spinner: s,
		message: message,
		active:  true,
	}
}

// SetMessage updates the spinner's display message.
func (m *SpinnerModel) SetMessage(msg string) {
	m.message = msg
}

// SetActive enables or disables the spinner animation.
func (m *SpinnerModel) SetActive(active bool) {
	m.active = active
}

// Init implements tea.Model.
func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update implements tea.Model.
func (m SpinnerModel) Update(msg tea.Msg) (SpinnerModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m SpinnerModel) View() string {
	if !m.active {
		return ""
	}
	return m.spinner.View() + " " + m.message
}
