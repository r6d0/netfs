package console

import (
	"fmt"
	"netfs/api"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.
	NewStyle().
	Align(lipgloss.Left, lipgloss.Center).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("ff"))

type hostsResponse struct {
	hosts []api.RemoteHost
	err   error
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type HostsView struct {
	width   int
	height  int
	list    list.Model
	network *api.Network
}

func (model HostsView) Init() tea.Cmd {
	return func() tea.Msg {
		hosts, err := model.network.Hosts()

		newHosts := []api.RemoteHost{}
		for index := range 100 {
			newHosts = append(newHosts, api.RemoteHost{Name: fmt.Sprintf("(%d) ", index+1) + hosts[0].Name, IP: hosts[0].IP})
		}
		return hostsResponse{hosts: newHosts, err: err}
	}
}

func (model HostsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var refreshListCmd tea.Cmd

	switch msg := msg.(type) {
	case hostsResponse:
		items := make([]list.Item, len(msg.hosts))
		for index, host := range msg.hosts {
			items[index] = item{title: host.Name, desc: host.IP.String()}
		}
		refreshListCmd = model.list.SetItems(items)
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		model.list.SetSize(msg.Width-h, msg.Height-v)
		docStyle = docStyle.Width((msg.Width / 100) * model.width).Height((msg.Height / 100) * model.height)
	}

	var cmd tea.Cmd
	model.list, cmd = model.list.Update(msg)
	return model, tea.Sequence(refreshListCmd, cmd)
}

func (model HostsView) View() string {
	return docStyle.Render(model.list.View())
}

func NewHostsView(network *api.Network, width int, height int) HostsView {
	lst := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	lst.DisableQuitKeybindings()
	lst.SetShowFilter(false)
	lst.SetShowHelp(false)
	lst.SetShowTitle(false)
	lst.SetShowStatusBar(false)
	// lst.SetShowPagination(false)

	return HostsView{list: lst, network: network, width: width, height: height}
}
