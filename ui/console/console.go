package console

import (
	"fmt"
	"netfs/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const NEW_LINE = "\n"

type ConsoleView struct {
	HostsView tea.Model
}

func (model ConsoleView) Init() tea.Cmd {
	return tea.Batch(model.HostsView.Init())
}

func (model ConsoleView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return model, tea.Quit
		}
	case tea.WindowSizeMsg:
		fmt.Println("Width: ", msg.Width, " Height: ", msg.Height)
	}

	var cmd tea.Cmd
	model.HostsView, cmd = model.HostsView.Update(msg)
	return model, cmd
}

func (model ConsoleView) View() string {
	return lipgloss.JoinVertical(lipgloss.Center, model.HostsView.View())
}

func NewConsoleViewModel(network *api.Network) tea.Model {
	return ConsoleView{HostsView: NewHostsView(network, 30, 100)}
}
