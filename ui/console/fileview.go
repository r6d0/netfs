package console

import (
	"netfs/api"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UpdateFilesMsg struct {
	Items []list.Item
}

type FileViewHistoryNode struct {
	File *api.RemoteFile
	Prev *FileViewHistoryNode
}

type FileViewItem struct {
	File *api.RemoteFile
}

func (item FileViewItem) Title() string       { return item.File.Info.FileName }
func (item FileViewItem) Description() string { return item.File.Info.FilePath }
func (item FileViewItem) FilterValue() string { return item.File.Info.FileName }

type FileView struct {
	list    list.Model
	style   lipgloss.Style
	prev    *FileViewHistoryNode
	host    *api.RemoteHost
	network *api.Network
}

func (model FileView) Init() tea.Cmd {
	return nil
}

func (model FileView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		// Enter to the selected directory.
		case tea.KeyEnter: // TODO. from settings?
			item := model.list.SelectedItem()
			file := item.(*FileViewItem).File
			if file.Info.FileType == api.DIRECTORY {
				prev := FileViewHistoryNode{File: file, Prev: model.prev}
				model.prev = &prev
				cmd = model.refreshFilesList(file)
			}
		case tea.KeyBackspace: // TODO. from settings?
			// Exit to the root directory of the selected host.
			if msg.Alt {
				file := model.host.Root()
				model.prev = &FileViewHistoryNode{File: file}
				cmd = model.refreshFilesList(file)
				// Exit from the selected directory.
			} else if model.prev.Prev != nil {
				model.prev = model.prev.Prev
				cmd = model.refreshFilesList(model.prev.File)
			}
		}
	case UpdateHostMsg:
		file := msg.Host.Root()
		model.host = msg.Host
		model.prev = &FileViewHistoryNode{File: file}
		cmd = model.refreshFilesList(file)
	case UpdateFilesMsg:
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

func (model FileView) View() string {
	return model.style.Render(model.list.View())
}

func NewFileView(network *api.Network) tea.Model {
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

	return FileView{list: lst, style: style, network: network}
}

func (model FileView) refreshFilesList(file *api.RemoteFile) tea.Cmd {
	return func() tea.Msg {
		// TODO. Show error.
		children, _ := file.Children(model.network.Transport(), 0, 100) // TODO. from settings?
		items := make([]list.Item, len(children))
		for index, file := range children {
			items[index] = &FileViewItem{File: &file}
		}
		return UpdateFilesMsg{Items: items}
	}
}
