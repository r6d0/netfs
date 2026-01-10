package main

// import (
// 	"fmt"
// 	"netfs/api"
// 	"netfs/api/transport"
// 	"time"
// )

// func main() {
// 	network, err := api.NewNetwork(api.NetworkConfig{Port: 8989, Protocol: transport.HTTP, Timeout: time.Second * 5})
// 	if err != nil {
// 		panic(err)
// 	}

// 	hosts, err := network.Hosts()
// 	fmt.Println(hosts)
// }

import (
	"netfs/api"
	"netfs/api/transport"
	"netfs/ui/console"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// type model struct {
// 	HostsView console.HostsViewModel
// }

// func (m model) Init() tea.Cmd {
// 	return tea.Batch(m.HostsView.Init())
// }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	return m, nil
// }

// func (m model) View() string {
// 	// s := "What should we buy at the market?\n\n"

// 	// var style = lipgloss.NewStyle().
// 	// 	Bold(true).
// 	// 	Foreground(lipgloss.Color("#9c9c9cff"))

// 	// for i, choice := range m.choices {
// 	// 	checked := " "
// 	// 	if _, ok := m.selected[i]; ok {
// 	// 		checked = "x"
// 	// 	}

// 	// 	s += fmt.Sprintf("[%s] %s\n", checked, style.Render(choice))
// 	// }

// 	// s += "\nPress q to quit.\n"

// 	// focusedModelStyle := lipgloss.NewStyle().
// 	// 	Padding(1).
// 	// 	Align(lipgloss.Left, lipgloss.Center).
// 	// 	BorderStyle(lipgloss.NormalBorder()).
// 	// 	BorderForeground(lipgloss.Color("ff"))

// 	// r := lipgloss.JoinVertical(lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Center, focusedModelStyle.Render(s), focusedModelStyle.Render(s)), lipgloss.JoinHorizontal(lipgloss.Right, focusedModelStyle.Render(s), focusedModelStyle.Render(s)))

// 	// return r

// 	return ""
// }

func main() {
	network, err := api.NewNetwork(api.NetworkConfig{Port: 8989, Protocol: transport.HTTP, Timeout: time.Second * 1})
	if err == nil {
		program := tea.NewProgram(console.NewConsoleViewModel(network), tea.WithAltScreen())
		_, err = program.Run()
	}

	if err != nil {
		panic(err)
	}
}
