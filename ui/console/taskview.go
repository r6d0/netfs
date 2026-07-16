package console

import (
	"fmt"
	"io"
	"netfs/api"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const COLUMN_PROGRESS_WIDTH = 4

type UpdateTaskMsg struct {
	Items []list.Item
}

type TaskViewItem struct {
	Task *api.RemoteCopyTask
}

func (item TaskViewItem) Title() string       { return item.Task.Source.Info.Name }
func (item TaskViewItem) Description() string { return item.Task.Source.Info.Name }
func (item TaskViewItem) FilterValue() string { return item.Task.Source.Info.Name }

type TaskViewItemDelegate struct {
	columnTitleStyle    lipgloss.Style
	columnCountStyle    lipgloss.Style
	columnProgressStyle lipgloss.Style
	itemStyle           lipgloss.Style
	itemSelectedStyle   lipgloss.Style
}

func (delegate TaskViewItemDelegate) Render(writer io.Writer, model list.Model, index int, item list.Item) {
	style := delegate.itemStyle
	if model.Index() == index {
		style = delegate.itemSelectedStyle
	}

	taskItem := item.(*TaskViewItem)
	title := taskItem.Task.Source.Host.Name + "/../" + taskItem.Task.Source.Info.Name + " to " + taskItem.Task.Target.Host.Name + "/../" + taskItem.Task.Target.Info.Name
	count := fmt.Sprintf("%d/%d ", taskItem.Task.Current, taskItem.Task.Count)
	progress := fmt.Sprintf("  %d", taskItem.Task.Progress) + "%"

	style = style.Width(model.Width())
	delegate.columnTitleStyle = delegate.columnTitleStyle.Width(model.Width() - (lipgloss.Width(count) + lipgloss.Width(progress)))

	writer.Write([]byte(
		style.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				delegate.columnTitleStyle.Render(title),
				delegate.columnCountStyle.Render(count),
				delegate.columnProgressStyle.Render(progress),
			),
		),
	))
}

func (TaskViewItemDelegate) Height() int { return 1 }

func (TaskViewItemDelegate) Spacing() int { return 0 }

func (TaskViewItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

type TaskView struct {
	list     list.Model
	style    lipgloss.Style
	host     *api.RemoteHost
	network  *api.Network
	delegate *TaskViewItemDelegate
}

func (model TaskView) Init() tea.Cmd {
	return nil
}

func (model TaskView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var listCmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeActiveHostMsg:
		model.host = msg.Host
		cmd = model.resolveTasks()
	case UpdateTaskMsg:
		cmd = model.list.SetItems(msg.Items)
	case ChangeActiveViewMsg:
		if msg.View == Task {
			model.style = model.style.BorderForeground(lipgloss.Color("#3b82f6"))
		} else {
			model.style = model.style.BorderForeground(lipgloss.Color("#ffffff"))
		}
	case RefreshMsg:
		if model.host != nil {
			cmd = model.resolveTasks()
		}
	case ResizeMsg:
		frameX, frameY := model.style.GetFrameSize()
		width := msg.Width - frameX
		height := msg.Height - frameY
		model.style = model.
			style.
			Width(width).
			Height(height)

		delegate := model.delegate
		delegate.itemStyle = delegate.itemStyle.Width(width)
		delegate.itemSelectedStyle = delegate.itemSelectedStyle.Width(width)

		model.list.SetSize(width, height)
	}
	model.list, listCmd = model.list.Update(msg)

	return model, tea.Sequence(cmd, listCmd)
}

func (model TaskView) View() string {
	return model.style.Render(model.list.View())
}

func (model TaskView) resolveTasks() tea.Cmd {
	return func() tea.Msg {
		// TODO. show error
		tasks, err := model.host.Tasks(model.network.Transport())
		if err == nil {
			items := make([]list.Item, len(tasks))
			for index := range items {
				items[index] = &TaskViewItem{Task: &tasks[index]}
			}
			return UpdateTaskMsg{Items: items}
		}

		return UpdateTaskMsg{Items: []list.Item{}}
	}
}

func NewTaskView(network *api.Network) tea.Model {
	delegate := TaskViewItemDelegate{
		columnTitleStyle:    lipgloss.NewStyle().AlignHorizontal(lipgloss.Left),
		columnCountStyle:    lipgloss.NewStyle().AlignHorizontal(lipgloss.Right),
		columnProgressStyle: lipgloss.NewStyle().AlignHorizontal(lipgloss.Right),
		itemStyle:           lipgloss.NewStyle(),
		itemSelectedStyle:   lipgloss.NewStyle().Background(lipgloss.Color("#3b82f6")),
	}

	lst := list.New([]list.Item{}, delegate, 0, 0)
	lst.DisableQuitKeybindings()
	lst.SetShowFilter(false)
	lst.SetShowHelp(false)
	lst.SetShowTitle(false)
	lst.SetShowStatusBar(false)
	lst.SetShowPagination(false)

	return &TaskView{
		network:  network,
		list:     lst,
		delegate: &delegate,
		style: lipgloss.
			NewStyle().
			Align(lipgloss.Left, lipgloss.Left).
			BorderForeground(lipgloss.Color("#ffffff")).
			BorderStyle(lipgloss.NormalBorder()),
	}
}
