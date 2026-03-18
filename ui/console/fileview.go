package console

import (
	"io"
	"netfs/api"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const COUNT_MAX_LEN = 5
const COLUMN_TYPE_WIDTH = 5
const COLUMN_SIZE_WIDTH = 15
const TOO_LONG_LINE_POSTFIX = "..."

var TOO_LONG_LINE_POSTFIX_WIDTH = lipgloss.Width(TOO_LONG_LINE_POSTFIX)

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

func (item FileViewItem) Title() string       { return item.File.Info.Name }
func (item FileViewItem) Description() string { return item.File.Info.Path }
func (item FileViewItem) FilterValue() string { return item.File.Info.Name }

type FileViewVolumeItem struct {
	Volume *api.RemoteVolume
}

func (item FileViewVolumeItem) Title() string       { return item.Volume.Info.Name }
func (item FileViewVolumeItem) Description() string { return item.Volume.Info.OsPath }
func (item FileViewVolumeItem) FilterValue() string { return item.Volume.Info.Name }

type FileViewItemDelegate struct {
	headerStyle       lipgloss.Style
	footerStyle       lipgloss.Style
	columnTypeStyle   lipgloss.Style
	columnNameStyle   lipgloss.Style
	columnSizeStyle   lipgloss.Style
	itemStyle         lipgloss.Style
	itemSelectedStyle lipgloss.Style
}

func (delegate FileViewItemDelegate) Render(writer io.Writer, model list.Model, index int, item list.Item) {
	style := delegate.itemStyle
	if model.Index() == index {
		style = delegate.itemSelectedStyle
	}

	typeColumn := ""
	nameColumn := ""
	sizeColumn := ""

	if volumeItem, ok := item.(*FileViewVolumeItem); ok {
		typeColumn = "-"
		nameColumn = volumeItem.Volume.Info.Name
		sizeColumn = "-"
	} else if fileItem, ok := item.(*FileViewItem); ok {
		typeColumn = fileItem.File.Info.Type.String()
		nameColumn = fileItem.File.Info.Name
		sizeColumn = fileItem.File.Info.Size.String()
	}

	nameWidth := delegate.columnNameStyle.GetWidth()
	if lipgloss.Width(nameColumn) > nameWidth {
		nameColumn = lipgloss.
			NewStyle().
			MaxWidth(nameWidth-TOO_LONG_LINE_POSTFIX_WIDTH).
			Render(nameColumn) + TOO_LONG_LINE_POSTFIX
	}

	writer.Write(
		[]byte(
			style.Render(
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					delegate.columnTypeStyle.Render(typeColumn),
					delegate.columnNameStyle.Render(nameColumn),
					delegate.columnSizeStyle.Render(sizeColumn),
				),
			),
		),
	)
}

func (FileViewItemDelegate) Height() int { return 1 }

func (FileViewItemDelegate) Spacing() int { return 0 }

func (FileViewItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

type FileViewHeader struct {
	delegate *FileViewItemDelegate
}

func (*FileViewHeader) Init() tea.Cmd {
	return nil
}

func (header *FileViewHeader) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return header, nil
}

func (header *FileViewHeader) View() string {
	delegate := header.delegate
	return delegate.headerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			delegate.columnTypeStyle.Render("TYPE"),
			delegate.columnNameStyle.Render("NAME"),
			delegate.columnSizeStyle.Render("SIZE"),
		),
	)
}

type FileViewFooter struct {
	count    int
	delegate *FileViewItemDelegate
}

func (*FileViewFooter) Init() tea.Cmd {
	return nil
}

func (footer *FileViewFooter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UpdateFilesMsg:
		footer.count = len(msg.Items)
	}

	return footer, nil
}

func (footer *FileViewFooter) View() string {
	delegate := footer.delegate

	count := strconv.Itoa(footer.count)
	countLen := len(count)
	buffer := strings.Builder{}
	buffer.WriteString("COUNT: ")
	for countLen < COUNT_MAX_LEN {
		countLen++
		buffer.WriteString(" ")
	}
	buffer.WriteString(count)

	return delegate.footerStyle.Render(buffer.String())
}

type FileView struct {
	header   tea.Model
	footer   tea.Model
	list     list.Model
	delegate *FileViewItemDelegate
	style    lipgloss.Style
	prev     *FileViewHistoryNode
	host     *api.RemoteHost
	network  *api.Network
	toCopy   *api.RemoteFile
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
				if file.Info.Type == api.DIRECTORY {
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
						parent = fileItem.File.Info.Path
					} else if volumeItem, ok := item.(*FileViewVolumeItem); ok {
						parent = volumeItem.Volume.Info.LocalPath
					}

					file := model.toCopy
					file.CopyTo( // TODO. show error.
						model.network.Transport(),
						api.RemoteFile{
							Host: *model.host,
							Info: api.FileInfo{
								Name: file.Info.Name,
								Path: parent + file.Info.Name,
								Type: file.Info.Type,
								Size: file.Info.Size,
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

		delegate := model.delegate
		delegate.headerStyle = delegate.headerStyle.Width(width)
		delegate.footerStyle = delegate.footerStyle.Width(width)
		delegate.columnTypeStyle = delegate.columnTypeStyle.Width(COLUMN_TYPE_WIDTH)
		delegate.columnNameStyle = delegate.columnNameStyle.Width(width - (COLUMN_TYPE_WIDTH + COLUMN_SIZE_WIDTH))
		delegate.columnSizeStyle = delegate.columnSizeStyle.Width(COLUMN_SIZE_WIDTH)
		delegate.itemStyle = delegate.itemStyle.Width(width)
		delegate.itemSelectedStyle = delegate.itemSelectedStyle.Width(width)

		headerSize := lipgloss.Height(model.header.View())
		footerSize := lipgloss.Height(model.footer.View())
		model.list.SetSize(width, height-headerSize-footerSize)
	}

	var headerCmd tea.Cmd
	model.header, headerCmd = model.header.Update(msg)

	var footerCmd tea.Cmd
	model.footer, footerCmd = model.footer.Update(msg)

	var listCmd tea.Cmd
	model.list, listCmd = model.list.Update(msg)
	return model, tea.Sequence(cmd, headerCmd, footerCmd, listCmd)
}

func (model FileView) View() string {
	return model.style.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			model.header.View(),
			model.list.View(),
			model.footer.View(),
		),
	)
}

func NewFileView(network *api.Network) tea.Model {
	delegate := &FileViewItemDelegate{
		headerStyle:       lipgloss.NewStyle().Height(1).BorderForeground(lipgloss.Color("#ffffff")).BorderStyle(lipgloss.NormalBorder()).BorderBottom(true),
		footerStyle:       lipgloss.NewStyle().AlignHorizontal(lipgloss.Right).Height(1).BorderForeground(lipgloss.Color("#ffffff")).BorderStyle(lipgloss.NormalBorder()).BorderTop(true),
		columnTypeStyle:   lipgloss.NewStyle().AlignHorizontal(lipgloss.Left),
		columnNameStyle:   lipgloss.NewStyle().AlignHorizontal(lipgloss.Left),
		columnSizeStyle:   lipgloss.NewStyle().AlignHorizontal(lipgloss.Right),
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
		BorderForeground(lipgloss.Color("#ffffff")).
		BorderStyle(lipgloss.NormalBorder())

	return FileView{
		header:   &FileViewHeader{delegate: delegate},
		footer:   &FileViewFooter{delegate: delegate},
		list:     lst,
		delegate: delegate,
		style:    style,
		network:  network,
	}
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
