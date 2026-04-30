package main

import (
	"lazymosh/cli"
	"lazymosh/config"
	"lazymosh/log"
	"lazymosh/pkg"
	"lazymosh/screens"
	"lazymosh/style"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RootModel holds all application state.
type RootModel struct {
	currentScreen pkg.Screen
	servers       []config.Server
	screenModel   tea.Model
	width         int
	height        int
	errMsg        string
}

func NewRootModel() RootModel {
	return RootModel{
		currentScreen: pkg.ScreenList,
		screenModel:   screens.NewListScreen(),
	}
}

func (m RootModel) Init() tea.Cmd {
	log.Info("lazymosh started — config: %s", config.Path())
	return nil
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			log.Info("quit")
			return m, tea.Quit
		}

	case pkg.NavigateMsg:
		m.currentScreen = msg.Screen
		switch msg.Screen {
		case pkg.ScreenList:
			m.screenModel = screens.NewListScreen()
		case pkg.ScreenAdd:
			m.screenModel = screens.NewAddScreen()
		case pkg.ScreenEdit:
			m.screenModel = screens.NewEditScreen(msg.Server.(config.Server), msg.Index)
		}
		return m, nil

	case pkg.ReloadMsg:
		log.Debug("reload: fetching server list")
		store, err := config.Load()
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.servers = store.Servers
		m.screenModel = screens.NewListScreen()
		return m, nil
	}

	// Pass updates to active screen
	updated, cmd := m.screenModel.Update(msg)
	m.screenModel = updated
	return m, cmd
}

func (m RootModel) View() string {
	nav := m.renderNav()
	content := m.screenModel.View()
	return nav + "\n" + content
}

func (m RootModel) renderNav() string {
	if m.width < 10 {
		m.width = 80
	}
	div := style.RenderDivider(m.width - style.AppMargin*2)

	items := []string{
		"[L]ist",
		"[A]dd",
		"[Esc]Back",
	}

	navLine := ""
	for _, item := range items {
		itemStyle := lipgloss.Style{}.Foreground(style.ColorMuted)
		navLine += itemStyle.Render(item) + "  "
	}

	header := lipgloss.Style{}.
		Foreground(style.ColorPrimary).
		Bold(true).
		Render("lazymosh")

	pad := strings.Repeat(" ", style.AppMargin)
	return pad + div + "\n" + pad + header + "  " + navLine + "\n" + pad + div
}

func main() {
	cfg, exit := cli.Parse()
	if exit {
		return
	}

	log.Debug("verbosity: %s, config: %s", cfg.Verbosity, config.Path())

	p := tea.NewProgram(NewRootModel())
	if err := p.Start(); err != nil {
		log.Fatal("tea program error: %v", err)
	}
}
