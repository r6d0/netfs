package console

import (
	"fmt"
	"netfs/api"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HostResponse struct {
	hosts []api.RemoteHost
	err   error
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type HostView struct {
	list         list.Model
	network      *api.Network
	defaultStyle lipgloss.Style
}

func (model HostView) Init() tea.Cmd {
	return func() tea.Msg {
		hosts, err := model.network.Hosts()
		newHosts := []api.RemoteHost{}
		if err == nil && len(hosts) > 0 {
			for index := range 100 {
				newHosts = append(newHosts, api.RemoteHost{Name: fmt.Sprintf("(%d) ", index+1) + hosts[0].Name, IP: hosts[0].IP})
			}
		}
		return HostResponse{hosts: newHosts, err: err}
	}
}

func (model HostView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var refreshListCmd tea.Cmd

	switch msg := msg.(type) {
	case HostResponse:
		items := make([]list.Item, len(msg.hosts))
		for index, host := range msg.hosts {
			items[index] = item{title: host.Name, desc: host.IP.String()}
		}
		refreshListCmd = model.list.SetItems(items)
	case ResizeMsg:
		frameX, frameY := model.defaultStyle.GetFrameSize()
		width := msg.Width - frameX
		height := msg.Height - frameY
		model.defaultStyle = model.
			defaultStyle.
			Width(width).
			Height(height)

		model.list.SetSize(width, height)
	}

	var cmd tea.Cmd
	model.list, cmd = model.list.Update(msg)
	return model, tea.Sequence(refreshListCmd, cmd)
}

func (model HostView) View() string {
	return model.defaultStyle.Render(model.list.View())
}

func NewHostView(network *api.Network) HostView {
	lst := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	lst.DisableQuitKeybindings()
	lst.SetShowFilter(false)
	lst.SetShowHelp(false)
	lst.SetShowTitle(false)
	lst.SetShowStatusBar(false)

	defaultStyle := lipgloss.
		NewStyle().
		Align(lipgloss.Left, lipgloss.Left).
		BorderForeground(lipgloss.Color("ff")).
		BorderStyle(lipgloss.NormalBorder())

	return HostView{list: lst, network: network, defaultStyle: defaultStyle}
}
