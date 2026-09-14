package console

import (
	"io"
	"netfs/api"
	"netfs/ui/console/message"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ChangeActiveHostMsg struct {
	Host  *api.Host
	Alive bool
}

type UpdateHostsMsg struct {
	Items []list.Item
	Index int
}

type HostViewItem struct {
	Host  *api.Host
	Alive bool
}

func (item HostViewItem) Equal(other HostViewItem) bool {
	return item.Host.IP.Equal(other.Host.IP)
}

func (item HostViewItem) Title() string       { return item.Host.Name }
func (item HostViewItem) Description() string { return item.Host.IP.String() }
func (item HostViewItem) FilterValue() string { return item.Host.Name }

type HostViewItemDelegate struct {
	itemStyle         lipgloss.Style
	itemSelectedStyle lipgloss.Style
}

func (delegate HostViewItemDelegate) Render(writer io.Writer, model list.Model, index int, item list.Item) {
	style := delegate.itemStyle
	if model.Index() == index {
		style = delegate.itemSelectedStyle
	}

	hostItem := item.(*HostViewItem)
	if !hostItem.Alive {
		style = style.Foreground(lipgloss.Color("#777575"))
	}

	writer.Write(
		[]byte(
			style.Render(
				strings.Join([]string{hostItem.Host.Name, "(", hostItem.Host.IP.String(), ")"}, ""),
			),
		),
	)
}

func (HostViewItemDelegate) Height() int { return 1 }

func (HostViewItemDelegate) Spacing() int { return 0 }

func (HostViewItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

type HostView struct {
	list        list.Model
	style       lipgloss.Style
	activeStyle lipgloss.Style
	network     *api.Network
	active      bool
}

func (model HostView) Init() tea.Cmd {
	return model.refreshHosts()
}

func (model HostView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			item := model.list.SelectedItem()
			hostItem := item.(*HostViewItem)

			if hostItem.Alive {
				cmd = tea.Sequence(
					func() tea.Msg { return ChangeActiveHostMsg{Host: hostItem.Host, Alive: hostItem.Alive} },
					func() tea.Msg { return ChangeActiveViewMsg{View: File} },
				)
			}
		}

	case ChangeActiveViewMsg:
		model.active = (msg.View == Host)
	case UpdateHostsMsg:
		selectedHost := model.selectedItem()

		model.list.SetItems(msg.Items)
		for index, item := range msg.Items {
			hostItem := item.(*HostViewItem)
			if selectedHost != nil && hostItem.Host.IP.Equal(selectedHost.IP) {
				model.list.Select(index)
				break
			}
		}
	case message.RefreshMsg:
		cmd = model.refreshHosts()
	case message.ResizeMsg:
		frameX, frameY := model.style.GetFrameSize()
		width := msg.Width - frameX
		height := msg.Height - frameY

		model.style = model.
			style.
			Width(width).
			Height(height)

		model.activeStyle = model.
			activeStyle.
			Width(width).
			Height(height)

		model.list.SetSize(width, height)
	}

	var listCmd tea.Cmd
	model.list, listCmd = model.list.Update(msg)
	return model, tea.Sequence(cmd, listCmd)
}

func (model HostView) View() string {
	if model.active {
		return model.activeStyle.Render(model.list.View())
	}
	return model.style.Render(model.list.View())
}

func (model HostView) refreshHosts() tea.Cmd {
	return func() tea.Msg {
		items := make([]list.Item, len(model.list.Items()))

		hosts, err := model.network.Hosts()
		if err == nil && len(hosts) > 0 { // TODO. show error
			for index, item := range model.list.Items() {
				hostItem := item.(*HostViewItem)
				items[index] = &HostViewItem{Host: hostItem.Host, Alive: slices.ContainsFunc(hosts, hostItem.Host.Equal)}
			}

			for _, host := range hosts {
				if !slices.ContainsFunc(items, func(item list.Item) bool { return item.(*HostViewItem).Host.Equal(host) }) {
					items = append(items, &HostViewItem{Host: &host, Alive: true})
				}
			}
		}
		return UpdateHostsMsg{Items: items}
	}
}

func (model HostView) selectedItem() *api.Host {
	item := model.list.SelectedItem()
	if item != nil {
		if hostItem, ok := item.(*HostViewItem); ok {
			return hostItem.Host
		}
	}
	return nil
}

func NewHostView(network *api.Network) tea.Model {
	delegate := HostViewItemDelegate{
		itemStyle:         lipgloss.NewStyle(),
		itemSelectedStyle: lipgloss.NewStyle().Background(lipgloss.Color("#3b82f6")),
	}

	lst := list.New([]list.Item{}, delegate, 0, 0)
	lst.DisableQuitKeybindings()
	lst.SetShowFilter(false)
	lst.SetShowHelp(false)
	lst.SetShowTitle(false)
	lst.SetShowStatusBar(false)
	lst.SetShowPagination(false)

	style := lipgloss.
		NewStyle().
		Align(lipgloss.Left, lipgloss.Left).
		BorderForeground(lipgloss.Color("#fff")).
		BorderStyle(lipgloss.NormalBorder())

	activeStyle := lipgloss.
		NewStyle().
		Align(lipgloss.Left, lipgloss.Left).
		BorderForeground(lipgloss.Color("#3b82f6")).
		BorderStyle(lipgloss.NormalBorder())

	return &HostView{list: lst, network: network, style: style, activeStyle: activeStyle}
}
