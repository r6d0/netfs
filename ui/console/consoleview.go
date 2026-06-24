package console

import (
	"netfs/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TODO. from settings?
const EnterKeyMsg = "enter"
const BackspaceKeyMsg = "backspace"
const AltBackspaceKeyMsg = "alt+backspace"
const QuitKeyMsg = "alt+q"
const HostActiveKeyMsg = "alt+h"
const FileActiveKeyMsg = "alt+f"

type ConsoleActiveView uint8

const (
	Host ConsoleActiveView = iota
	File
)

// The event sends after changing the terminal size.
type ResizeMsg struct {
	Width  int
	Height int
}

// The event sends every N seconds.
type RefreshMsg struct{}

// The event sends after switching to another view.
type ChangeActiveView struct {
	View ConsoleActiveView
}

// The main view of the UI.
type ConsoleView struct {
	hostsView  tea.Model
	fileView   tea.Model
	activeView ConsoleActiveView
	style      lipgloss.Style
}

func (model ConsoleView) Init() tea.Cmd {
	return tea.Sequence(
		model.hostsView.Init(),
		model.fileView.Init(),
		func() tea.Msg { return ChangeActiveView{View: Host} },
	)
}

func (model ConsoleView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var hostViewCmd tea.Cmd
	var fileViewCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case QuitKeyMsg:
			return model, tea.Quit
		case HostActiveKeyMsg:
			return model, func() tea.Msg { return ChangeActiveView{View: Host} }
		case FileActiveKeyMsg:
			return model, func() tea.Msg { return ChangeActiveView{View: File} }
		}

		switch model.activeView {
		case Host:
			model.hostsView, hostViewCmd = model.hostsView.Update(msg)
		case File:
			model.fileView, fileViewCmd = model.fileView.Update(msg)
		}

	case ChangeActiveView:
		switch msg.View {
		case Host:
			model.activeView = Host
		case File:
			model.activeView = File
		}
		model.hostsView, hostViewCmd = model.hostsView.Update(msg)
		model.fileView, fileViewCmd = model.fileView.Update(msg)

	case tea.WindowSizeMsg:
		frameX, frameY := model.style.GetFrameSize()
		width := float32(msg.Width - frameX)
		height := float32(msg.Height - frameY)
		model.style = model.
			style.
			Width(int(width)).
			Height(int(height))

		// TODO. from settings?
		hostViewWidth := (width / 100.0) * 30.0
		fileViewWidth := int(width - hostViewWidth)

		model.hostsView, hostViewCmd = model.hostsView.Update(ResizeMsg{Width: int(hostViewWidth), Height: int(height)})
		model.fileView, fileViewCmd = model.fileView.Update(ResizeMsg{Width: fileViewWidth, Height: int(height)})
	default:
		model.hostsView, hostViewCmd = model.hostsView.Update(msg)
		model.fileView, fileViewCmd = model.fileView.Update(msg)

		return model, tea.Sequence(hostViewCmd, fileViewCmd)
	}

	return model, tea.Sequence(hostViewCmd, fileViewCmd)
}

func (model ConsoleView) View() string {
	return model.style.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			model.hostsView.View(),
			model.fileView.View(),
		),
	)
}

// The function returns new instance of ConsoleView.
func NewConsoleViewModel(network *api.Network) tea.Model {
	style := lipgloss.
		NewStyle().
		Align(lipgloss.Left, lipgloss.Left)

	return ConsoleView{hostsView: NewHostView(network), fileView: NewFileView(network), style: style}
}
