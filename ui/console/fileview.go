package console

import (
	"netfs/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FileView struct {
	defaultStyle lipgloss.Style
}

func (model FileView) Init() tea.Cmd {
	return nil
}

func (model FileView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ResizeMsg:
		frameX, frameY := model.defaultStyle.GetFrameSize()
		model.defaultStyle = model.
			defaultStyle.
			Width(msg.Width - frameX).
			Height(msg.Height - frameY)
	}

	return model, nil
}

func (model FileView) View() string {
	return model.defaultStyle.Render("HELLO")
}

func NewFileView(network *api.Network) tea.Model {
	defaultStyle := lipgloss.
		NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("ff")).
		BorderStyle(lipgloss.NormalBorder())

	return FileView{defaultStyle: defaultStyle}
}
