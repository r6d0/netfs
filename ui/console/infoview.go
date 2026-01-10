package console

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InfoView struct {
	defaultStyle lipgloss.Style
}

func (model InfoView) Init() tea.Cmd {
	return nil
}

func (model InfoView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (model InfoView) View() string {
	return model.defaultStyle.Render("HELLO")
}

func NewInfoView() tea.Model {
	defaultStyle := lipgloss.
		NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("ff")).
		BorderStyle(lipgloss.NormalBorder())

	return InfoView{defaultStyle: defaultStyle}
}
