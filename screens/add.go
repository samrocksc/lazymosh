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

// AddModel is the "add server" screen.
type AddModel struct {
	nameField     string
	hostField     string
	userField     string
	portField     string
	localityField string

	focusIndex int // 0=name 1=host 2=user 3=port 4=locality
	width     int
	errMsg    string
	saving    bool
}

const fieldLabelWidth = 10

func NewAddScreen() tea.Model {
	return AddModel{portField: "22", focusIndex: 0}
}

func (m AddModel) Init() tea.Cmd { return nil }

func (m AddModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case SaveErrMsg:
		m.errMsg = msg.Err.Error()
		m.saving = false
		return m, nil
	case SaveSuccessMsg:
		// Background save succeeded — navigate back to list so it reloads from disk
		return m, func() tea.Msg { return pkg.NavigateMsg{Screen: pkg.ScreenList} }
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m AddModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.saving {
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

	default:
		if msg.Type == tea.KeyRunes {
			m.insert(string(msg.Runes))
		}
		return m, nil
	}
}

func (m *AddModel) insert(s string) {
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

func (m *AddModel) backspace() {
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

func (m AddModel) save() tea.Cmd {
	if m.nameField == "" || m.hostField == "" || m.userField == "" {
		m.errMsg = "name, host, and user are required"
		return nil
	}

	port := 22
	if m.portField != "" {
		fmt.Sscanf(m.portField, "%d", &port)
	}

	srv := config.Server{
		ID:       genID(),
		Name:     m.nameField,
		Host:     m.hostField,
		Port:     port,
		User:     m.userField,
		Locality: m.localityField,
	}

	m.saving = true
	return func() tea.Msg {
		store, err := config.Load()
		if err != nil {
			return SaveErrMsg{Err: err}
		}
		store.Servers = append(store.Servers, srv)
		if err := config.Save(store); err != nil {
			return SaveErrMsg{Err: err}
		}
		return SaveSuccessMsg{}
	}
}

func (m AddModel) View() string {
	if m.width < 10 {
		m.width = 80
	}
	div := style.RenderDivider(m.width)

	lines := []string{
		style.StyleHeader.Render("Add Server"),
		div,
		"  " + m.fieldLine("name", "Name", m.nameField, 0),
		"  " + m.fieldLine("host", "Host", m.hostField, 1),
		"  " + m.fieldLine("user", "User", m.userField, 2),
		"  " + m.fieldLine("port", "Port", m.portField, 3),
		"  " + m.fieldLine("locality", "Locality", m.localityField, 4),
		"",
		style.StyleMuted.Render("  [Tab/↑/↓] navigate   [Enter] next field / save   [Esc] cancel"),
		style.StyleMuted.Render("  Locality is optional (e.g. eu-berlin, us-east)"),
	}

	if m.errMsg != "" {
		lines = append(lines, "", style.Error(m.errMsg))
	}

	return strings.Join(lines, "\n")
}

func (m AddModel) fieldLine(key, label, value string, idx int) string {
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
		style.StyleMuted.Render("["+key+"]"),
	)
}

func genID() string {
	store, err := config.Load()
	if err != nil {
		store = &config.Store{}
	}
	return fmt.Sprintf("%d", len(store.Servers)+1)
}

// Package-local messages
type SaveErrMsg struct{ Err error }
type SaveSuccessMsg struct{}
