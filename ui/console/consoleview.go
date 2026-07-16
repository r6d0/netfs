package console

import (
	"netfs/api"
	"time"

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
const TaskActiveKeyMsg = "alt+t"

type ConsoleActiveView uint8

const (
	Host ConsoleActiveView = iota
	File
	Task
)

// The event sends after changing the terminal size.
type ResizeMsg struct {
	Width  int
	Height int
}

// The event sends every N seconds.
type RefreshMsg struct{}

// The event sends after switching to another view.
type ChangeActiveViewMsg struct {
	View ConsoleActiveView
}

// The main view of the UI.
type ConsoleView struct {
	hostsView  tea.Model
	fileView   tea.Model
	taskView   tea.Model
	activeView ConsoleActiveView
	style      lipgloss.Style
}

func (model ConsoleView) Init() tea.Cmd {
	return tea.Sequence(
		model.hostsView.Init(),
		model.fileView.Init(),
		model.taskView.Init(),
		func() tea.Msg { return ChangeActiveViewMsg{View: Host} },
		tea.Every(3*time.Second, func(t time.Time) tea.Msg { return RefreshMsg{} }), // TODO. 3*time.Second - from settings
	)
}

func (model ConsoleView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var hostViewCmd tea.Cmd
	var fileViewCmd tea.Cmd
	var taskViewCmd tea.Cmd

	switch msg := msg.(type) {
	case RefreshMsg:
		cmd = tea.Every(3*time.Second, func(t time.Time) tea.Msg { return RefreshMsg{} }) // TODO. 3*time.Second - from settings
		model.hostsView, hostViewCmd = model.hostsView.Update(msg)
		model.fileView, fileViewCmd = model.fileView.Update(msg)
		model.taskView, taskViewCmd = model.taskView.Update(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case QuitKeyMsg:
			return model, tea.Quit
		case HostActiveKeyMsg:
			return model, func() tea.Msg { return ChangeActiveViewMsg{View: Host} }
		case FileActiveKeyMsg:
			return model, func() tea.Msg { return ChangeActiveViewMsg{View: File} }
		case TaskActiveKeyMsg:
			return model, func() tea.Msg { return ChangeActiveViewMsg{View: Task} }
		}

		switch model.activeView {
		case Host:
			model.hostsView, hostViewCmd = model.hostsView.Update(msg)
		case File:
			model.fileView, fileViewCmd = model.fileView.Update(msg)
		case Task:
			model.taskView, taskViewCmd = model.taskView.Update(msg)
		}

	case ChangeActiveViewMsg:
		switch msg.View {
		case Host:
			model.activeView = Host
		case File:
			model.activeView = File
		case Task:
			model.activeView = Task
		}
		model.hostsView, hostViewCmd = model.hostsView.Update(msg)
		model.fileView, fileViewCmd = model.fileView.Update(msg)
		model.taskView, taskViewCmd = model.taskView.Update(msg)

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
		fileViewHeight := int((height / 100.0) * 70.0)

		model.hostsView, hostViewCmd = model.hostsView.Update(ResizeMsg{Width: int(hostViewWidth), Height: int(height)})
		model.fileView, fileViewCmd = model.fileView.Update(ResizeMsg{Width: fileViewWidth, Height: fileViewHeight})
		model.taskView, taskViewCmd = model.taskView.Update(ResizeMsg{Width: fileViewWidth, Height: int(height) - fileViewHeight})
	default:
		model.hostsView, hostViewCmd = model.hostsView.Update(msg)
		model.fileView, fileViewCmd = model.fileView.Update(msg)
		model.taskView, taskViewCmd = model.taskView.Update(msg)

		return model, tea.Sequence(cmd, hostViewCmd, fileViewCmd, taskViewCmd)
	}

	return model, tea.Sequence(cmd, hostViewCmd, fileViewCmd, taskViewCmd)
}

func (model ConsoleView) View() string {
	return model.style.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			model.hostsView.View(),
			lipgloss.JoinVertical(
				lipgloss.Left,
				model.fileView.View(),
				model.taskView.View(),
			),
		),
	)
}

// The function returns new instance of ConsoleView.
func NewConsoleViewModel(network *api.Network) tea.Model {
	style := lipgloss.
		NewStyle().
		Align(lipgloss.Left, lipgloss.Left)

	return ConsoleView{
		hostsView: NewHostView(network),
		fileView:  NewFileView(network),
		taskView:  NewTaskView(network),
		style:     style,
	}
}
