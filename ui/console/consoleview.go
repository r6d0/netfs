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

type ResizeMsg struct {
	Width  int
	Height int
}

type ConsoleView struct {
	hostsView    tea.Model
	fileView     tea.Model
	activeView   ConsoleActiveView
	defaultStyle lipgloss.Style
}

func (model ConsoleView) Init() tea.Cmd {
	return tea.Sequence(model.hostsView.Init(), model.fileView.Init())
}

func (model ConsoleView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var hostViewCmd tea.Cmd
	var fileViewCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch model.activeView {
		case Host:
			model.hostsView, hostViewCmd = model.hostsView.Update(msg)
		case File:
			model.fileView, fileViewCmd = model.fileView.Update(msg)
		}

		switch msg.String() {
		case QuitKeyMsg:
			return model, tea.Quit
		case HostActiveKeyMsg:
			model.activeView = Host
			return model, nil
		case FileActiveKeyMsg:
			model.activeView = File
			return model, nil
		}

	case UpdateFilesMsg:
		model.activeView = File
		model.fileView, fileViewCmd = model.fileView.Update(msg)
		return model, nil

	case tea.WindowSizeMsg:
		frameX, frameY := model.defaultStyle.GetFrameSize()
		width := float32(msg.Width - frameX)
		height := float32(msg.Height - frameY)
		model.defaultStyle = model.
			defaultStyle.
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
	// return ""
	return model.defaultStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			model.hostsView.View(),
			model.fileView.View(),
		),
	)
}

func NewConsoleViewModel(network *api.Network) tea.Model {
	defaultStyle := lipgloss.
		NewStyle().
		Align(lipgloss.Left, lipgloss.Left)

	return ConsoleView{hostsView: NewHostView(network), fileView: NewFileView(network), activeView: Host, defaultStyle: defaultStyle}
}
