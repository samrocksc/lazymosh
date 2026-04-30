package screens

import (
	"fmt"
	"strings"

	"lazymosh/config"
	"lazymosh/pkg"
	"lazymosh/style"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EditModel is the "edit server" screen.
type EditModel struct {
	index         int
	original      config.Server
	nameField     string
	hostField     string
	userField     string
	portField     string
	localityField string

	focusIndex     int
	width          int
	errMsg         string
	successMsg     string
	saving         bool
	confirmDelete  bool
}

func NewEditScreen(srv config.Server, index int) tea.Model {
	return EditModel{
		index:         index,
		original:      srv,
		nameField:     srv.Name,
		hostField:     srv.Host,
		userField:     srv.User,
		portField:     fmt.Sprintf("%d", srv.Port),
		localityField: srv.Locality,
		focusIndex:    0,
	}
}

func (m EditModel) Init() tea.Cmd { return nil }

func (m EditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case SaveErrMsg:
		m.errMsg = msg.Err.Error()
		m.saving = false
		return m, nil
	case SaveSuccessMsg:
		return m, func() tea.Msg { return pkg.ReloadMsg{} }
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m EditModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	if m.confirmDelete {
		switch msg.String() {
		case "y", "Y":
			return m, m.deleteServer()
		case "n", "N", "esc":
			m.confirmDelete = false
			return m, nil
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		return m, func() tea.Msg { return pkg.NavigateMsg{Screen: pkg.ScreenList} }

	case "tab":
		m.focusIndex = (m.focusIndex + 1) % 5
		return m, nil

	case "shift+tab":
		m.focusIndex = (m.focusIndex + 4) % 5
		return m, nil

	case "up", "k":
		m.focusIndex = (m.focusIndex + 4) % 5
		return m, nil

	case "down", "j":
		m.focusIndex = (m.focusIndex + 1) % 5
		return m, nil

	case "enter":
		if m.focusIndex < 4 {
			m.focusIndex++
			return m, nil
		}
		return m, m.save()

	case "backspace":
		m.backspace()
		return m, nil

	case "d":
		m.confirmDelete = true
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			m.insert(string(msg.Runes))
		}
		return m, nil
	}
}

func (m *EditModel) insert(s string) {
	switch m.focusIndex {
	case 0:
		m.nameField += s
	case 1:
		m.hostField += s
	case 2:
		m.userField += s
	case 3:
		m.portField += s
	case 4:
		m.localityField += s
	}
}

func (m *EditModel) backspace() {
	switch m.focusIndex {
	case 0:
		if len(m.nameField) > 0 {
			m.nameField = m.nameField[:len(m.nameField)-1]
		}
	case 1:
		if len(m.hostField) > 0 {
			m.hostField = m.hostField[:len(m.hostField)-1]
		}
	case 2:
		if len(m.userField) > 0 {
			m.userField = m.userField[:len(m.userField)-1]
		}
	case 3:
		if len(m.portField) > 0 {
			m.portField = m.portField[:len(m.portField)-1]
		}
	case 4:
		if len(m.localityField) > 0 {
			m.localityField = m.localityField[:len(m.localityField)-1]
		}
	}
}

func (m EditModel) save() tea.Cmd {
	if m.nameField == "" || m.hostField == "" || m.userField == "" {
		m.errMsg = "name, host, and user are required"
		return nil
	}

	port := 22
	fmt.Sscanf(m.portField, "%d", &port)

	m.saving = true
	return func() tea.Msg {
		store, err := config.Load()
		if err != nil {
			return SaveErrMsg{Err: err}
		}
		store.Servers[m.index] = config.Server{
			ID:       m.original.ID,
			Name:     m.nameField,
			Host:     m.hostField,
			Port:     port,
			User:     m.userField,
			Locality: m.localityField,
		}
		if err := config.Save(store); err != nil {
			return SaveErrMsg{Err: err}
		}
		return SaveSuccessMsg{}
	}
}

func (m EditModel) deleteServer() tea.Cmd {
	m.saving = true
	return func() tea.Msg {
		store, err := config.Load()
		if err != nil {
			return SaveErrMsg{Err: err}
		}
		store.Servers = append(store.Servers[:m.index], store.Servers[m.index+1:]...)
		if err := config.Save(store); err != nil {
			return SaveErrMsg{Err: err}
		}
		return pkg.ReloadMsg{}
	}
}

func (m EditModel) View() string {
	if m.width < 10 {
		m.width = 80
	}
	div := style.RenderDivider(m.width)

	lines := []string{
		style.StyleHeader.Render("Edit Server"),
		div,
		style.StyleMuted.Render(fmt.Sprintf("  ID: %s", m.original.ID)),
		"  " + m.fieldLine("name", "Name", m.nameField, 0),
		"  " + m.fieldLine("host", "Host", m.hostField, 1),
		"  " + m.fieldLine("user", "User", m.userField, 2),
		"  " + m.fieldLine("port", "Port", m.portField, 3),
		"  " + m.fieldLine("locality", "Locality", m.localityField, 4),
		"",
		style.StyleMuted.Render("  [Tab/↓/↑] navigate   [Enter] save   [D] delete   [Esc] cancel"),
	}

	if m.confirmDelete {
		lines = append(lines, "", style.Warn("  Confirm delete? [Y]es / [N]o"))
	}

	if m.errMsg != "" {
		lines = append(lines, "", style.Error(m.errMsg))
	}

	return strings.Join(lines, "\n")
}

func (m EditModel) fieldLine(key, label, value string, idx int) string {
	styleLabel := lipgloss.Style{}.Foreground(style.ColorMuted).Width(fieldLabelWidth)
	styleValue := lipgloss.Style{}.
			Foreground(style.ColorText).
			Width(30)

	v := value
	if v == "" {
		v = "_"
	}

	prefix := "  "
	if idx == m.focusIndex {
		prefix = " ▶ "
		styleValue = lipgloss.Style{}.
				Foreground(style.ColorPrimary).
				Width(30)
	}

	return fmt.Sprintf("%s%s  %s %s",
		prefix,
		styleLabel.Render(label),
		styleValue.Render(v),
		lipgloss.Style{}.Foreground(style.ColorMuted).Render("["+key+"]"),
	)
}
