package console

import (
	"netfs/api"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UpdateHostMsg struct {
	Host  *api.RemoteHost
	Error error
}

type UpdateHostsMsg struct {
	Items []list.Item
	Error error
}

type HostViewItem struct {
	Host *api.RemoteHost
}

func (item HostViewItem) Title() string       { return item.Host.Name }
func (item HostViewItem) Description() string { return item.Host.IP.String() }
func (item HostViewItem) FilterValue() string { return item.Host.Name }

type HostView struct {
	list    list.Model
	style   lipgloss.Style
	network *api.Network
}

func (model HostView) Init() tea.Cmd {
	return func() tea.Msg {
		hosts, err := model.network.Hosts()
		if err == nil && len(hosts) > 0 {
			items := make([]list.Item, len(hosts))
			for index, host := range hosts {
				items[index] = &HostViewItem{Host: &host}
			}
			return UpdateHostsMsg{Items: items}
		}
		return UpdateHostsMsg{Error: err}
	}
}

func (model HostView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter { // TODO. from settings?
			item := model.list.SelectedItem()
			cmd = func() tea.Msg { return UpdateHostMsg{Host: item.(*HostViewItem).Host} }
		}
	case UpdateHostsMsg:
		cmd = model.list.SetItems(msg.Items)
	case ResizeMsg:
		frameX, frameY := model.style.GetFrameSize()
		width := msg.Width - frameX
		height := msg.Height - frameY
		model.style = model.
			style.
			Width(width).
			Height(height)

		model.list.SetSize(width, height)
	}

	var listCmd tea.Cmd
	model.list, listCmd = model.list.Update(msg)
	return model, tea.Sequence(cmd, listCmd)
}

func (model HostView) View() string {
	return model.style.Render(model.list.View())
}

func NewHostView(network *api.Network) HostView {
	lst := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	lst.DisableQuitKeybindings()
	lst.SetShowFilter(false)
	lst.SetShowHelp(false)
	lst.SetShowTitle(false)
	lst.SetShowStatusBar(false)

	style := lipgloss.
		NewStyle().
		Align(lipgloss.Left, lipgloss.Left).
		BorderForeground(lipgloss.Color("ff")).
		BorderStyle(lipgloss.NormalBorder())

	return HostView{list: lst, network: network, style: style}
}
