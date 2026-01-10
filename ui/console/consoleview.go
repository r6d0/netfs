package console

import (
	"netfs/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TODO. from settings?
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
	infoView     tea.Model
	activeView   ConsoleActiveView
	defaultStyle lipgloss.Style
}

func (model ConsoleView) Init() tea.Cmd {
	return tea.Sequence(model.hostsView.Init(), model.fileView.Init(), model.infoView.Init())
}

func (model ConsoleView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var hostViewCmd tea.Cmd
	var fileViewCmd tea.Cmd
	var infoViewCmd tea.Cmd

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
		fileViewHeight := (height / 100.0) * 80.0
		infoViewHeight := (height / 100.0) * 20.0
		hostViewHeight := int(fileViewHeight) + int(infoViewHeight)

		model.hostsView, hostViewCmd = model.hostsView.Update(ResizeMsg{Width: int(hostViewWidth), Height: hostViewHeight})
		model.fileView, fileViewCmd = model.fileView.Update(ResizeMsg{Width: fileViewWidth, Height: int(fileViewHeight)})
		model.infoView, infoViewCmd = model.infoView.Update(ResizeMsg{Width: fileViewWidth, Height: int(infoViewHeight)})
	default:
		model.hostsView, hostViewCmd = model.hostsView.Update(msg)
		model.fileView, fileViewCmd = model.fileView.Update(msg)
		model.infoView, infoViewCmd = model.infoView.Update(msg)

		return model, tea.Sequence(hostViewCmd, fileViewCmd, infoViewCmd)
	}

	return model, tea.Sequence(hostViewCmd, fileViewCmd, infoViewCmd)
}

func (model ConsoleView) View() string {
	// return ""
	return model.defaultStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			model.hostsView.View(),
			lipgloss.JoinVertical(
				lipgloss.Top,
				model.fileView.View(),
				model.infoView.View(),
			),
		),
	)
}

func NewConsoleViewModel(network *api.Network) tea.Model {
	defaultStyle := lipgloss.
		NewStyle().
		Align(lipgloss.Left, lipgloss.Left)

	return ConsoleView{hostsView: NewHostView(network), fileView: NewFileView(network), infoView: NewInfoView(), activeView: Host, defaultStyle: defaultStyle}
}
