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
	Item list.Item
	Prev *FileViewHistoryNode
}

type FileViewItem struct {
	File *api.RemoteFile
}

func (item FileViewItem) Title() string       { return item.File.Info.FileName }
func (item FileViewItem) Description() string { return item.File.Info.FilePath }
func (item FileViewItem) FilterValue() string { return item.File.Info.FileName }

type FileViewVolumeItem struct {
	Volume *api.RemoteVolume
}

func (item FileViewVolumeItem) Title() string       { return item.Volume.Info.Name }
func (item FileViewVolumeItem) Description() string { return item.Volume.Info.OsPath }
func (item FileViewVolumeItem) FilterValue() string { return item.Volume.Info.Name }

type FileView struct {
	list    list.Model
	style   lipgloss.Style
	prev    *FileViewHistoryNode
	host    *api.RemoteHost
	network *api.Network
	toCopy  *api.RemoteFile
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
			if fileItem, ok := item.(*FileViewItem); ok {
				file := fileItem.File
				if file.Info.FileType == api.DIRECTORY {
					prev := FileViewHistoryNode{Item: item, Prev: model.prev}
					model.prev = &prev
					cmd = model.resolveFileChildren(file)
				}
			} else if volumeItem, ok := item.(*FileViewVolumeItem); ok {
				prev := FileViewHistoryNode{Item: item, Prev: model.prev}
				model.prev = &prev
				cmd = model.resolveVolumeChildren(volumeItem.Volume)
			}
		case tea.KeyBackspace: // TODO. from settings?
			// Exit to the root directory of the selected host.
			if msg.Alt {
				cmd = func() tea.Msg { return UpdateHostMsg{Host: model.host} }
				// Exit from the selected directory.
			} else if model.prev.Prev != nil {
				model.prev = model.prev.Prev
				item := model.prev.Item
				if item == nil {
					cmd = func() tea.Msg { return UpdateHostMsg{Host: model.host} }
				} else if fileItem, ok := item.(*FileViewItem); ok {
					cmd = model.resolveFileChildren(fileItem.File)
				} else if volumeItem, ok := item.(*FileViewVolumeItem); ok {
					cmd = model.resolveVolumeChildren(volumeItem.Volume)
				}
			}
		// Marks the file for copying.
		case tea.KeyCtrlC:
			item := model.list.SelectedItem()
			if _, ok := item.(*FileViewItem); ok {
				model.toCopy = item.(*FileViewItem).File
			}
		// Starts the file copying.
		case tea.KeyCtrlV:
			if model.toCopy != nil && model.prev.Item != nil {
				var parent string
				cmd = func() tea.Msg {
					item := model.prev.Item
					if fileItem, ok := item.(*FileViewItem); ok {
						parent = fileItem.File.Info.FilePath
					} else if volumeItem, ok := item.(*FileViewVolumeItem); ok {
						parent = volumeItem.Volume.Info.LocalPath
					}

					file := model.toCopy
					file.CopyTo( // TODO. show error.
						model.network.Transport(),
						api.RemoteFile{
							Host: *model.host,
							Info: api.FileInfo{
								FileName: file.Info.FileName,
								FilePath: parent + file.Info.FileName,
								FileType: file.Info.FileType,
								FileSize: file.Info.FileSize,
							},
						},
					)
					return nil
				}
			}
		}
	case UpdateHostMsg:
		model.prev = &FileViewHistoryNode{}
		model.host = msg.Host
		cmd = model.resolveHostVolumes(msg.Host)
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

func (model FileView) resolveFileChildren(file *api.RemoteFile) tea.Cmd {
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

func (model FileView) resolveVolumeChildren(volume *api.RemoteVolume) tea.Cmd {
	return func() tea.Msg {
		// TODO. Show error.
		children, _ := volume.Children(model.network.Transport(), 0, 100) // TODO. from settings?
		items := make([]list.Item, len(children))
		for index, file := range children {
			items[index] = &FileViewItem{File: &file}
		}
		return UpdateFilesMsg{Items: items}
	}
}

func (model FileView) resolveHostVolumes(host *api.RemoteHost) tea.Cmd {
	return func() tea.Msg {
		// TODO. Show error.
		volumes, _ := host.Volumes(model.network.Transport())
		items := make([]list.Item, len(volumes))
		for index, file := range volumes {
			items[index] = &FileViewVolumeItem{Volume: &file}
		}
		return UpdateFilesMsg{Items: items}
	}
}
