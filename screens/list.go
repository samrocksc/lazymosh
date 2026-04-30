package screens

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"lazymosh/config"
	"lazymosh/log"
	"lazymosh/pkg"
	"lazymosh/style"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListModel is the main server list screen.
type ListModel struct {
	servers    []config.Server
	selected   int
	width      int
	connecting bool
	errMsg     string
	successMsg string
}

func NewListScreen() tea.Model {
	return ListModel{}
}

func (m ListModel) Init() tea.Cmd {
	return func() tea.Msg {
		store, err := config.Load()
		if err != nil {
			log.Error("load servers: %v", err)
			return LoadErrMsg{Err: err}
		}
		log.Info("loaded %d servers from %s", len(store.Servers), config.Path())
		return LoadMsg{Servers: store.Servers}
	}
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case LoadMsg:
		m.servers = msg.Servers
		return m, nil
	case LoadErrMsg:
		m.errMsg = msg.Err.Error()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m ListModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "a":
		return m, func() tea.Msg { return pkg.NavigateMsg{Screen: pkg.ScreenAdd} }

	case "e":
		if len(m.servers) == 0 {
			return m, nil
		}
		return m, func() tea.Msg {
			return pkg.NavigateMsg{Screen: pkg.ScreenEdit, Server: m.servers[m.selected], Index: m.selected}
		}

	case "d":
		if len(m.servers) == 0 {
			return m, nil
		}
		return m.handleDelete()

	case "enter":
		if len(m.servers) == 0 || m.connecting {
			return m, nil
		}
		return m, m.connect(m.servers[m.selected])

	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
		return m, nil
	case "down", "j":
		if m.selected < len(m.servers)-1 {
			m.selected++
		}
		return m, nil
	}
	return m, nil
}

func (m ListModel) handleDelete() (tea.Model, tea.Cmd) {
	if m.selected < 0 || m.selected >= len(m.servers) {
		return m, nil
	}
	log.Debug("delete: index %d (%s)", m.selected, m.servers[m.selected].Name)
	store, err := config.Load()
	if err != nil {
		m.errMsg = err.Error()
		return m, nil
	}
	var next []config.Server
	for i, s := range store.Servers {
		if i != m.selected {
			next = append(next, s)
		}
	}
	store.Servers = next
	if err := config.Save(store); err != nil {
		m.errMsg = err.Error()
		return m, nil
	}
	log.Info("deleted server (now %d remaining)", len(next))
	m.servers = next
	if m.selected >= len(m.servers) && m.selected > 0 {
		m.selected--
	}
	m.successMsg = "Server deleted"
	return m, nil
}

func (m ListModel) connect(srv config.Server) tea.Cmd {
	return func() tea.Msg {
		host := srv.Host
		if srv.Port != 22 {
			host = fmt.Sprintf("%s:%d", srv.Host, srv.Port)
		}
		log.Debug("connecting: mosh %s@%s", srv.User, host)

		// Try mosh first
		cmd := exec.Command("mosh", srv.User+"@"+host)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setctty: true,
			Setsid:  true,
		}
		err := cmd.Start()
		if err == nil {
			log.Info("mosh started — handing off (pid %d)", cmd.Process.Pid)
			os.Exit(0)
			return nil
		}
		log.Warn("mosh failed (%v), falling back to ssh", err)

		// mosh failed — fall back to ssh
		sshCmd := exec.Command("ssh", srv.User+"@"+host)
		sshCmd.Stdin = os.Stdin
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr
		sshCmd.SysProcAttr = &syscall.SysProcAttr{
			Setctty: true,
			Setsid:  true,
		}
		if err := sshCmd.Start(); err != nil {
			return pkg.ConnectErrMsg{Err: fmt.Errorf("ssh fallback failed: %w", err)}
		}
		os.Exit(0)
		return nil
	}
}

func (m ListModel) View() string {
	if m.width < 10 {
		m.width = 80
	}
	div := style.RenderDivider(m.width)

	lines := []string{
		style.StyleHeader.Render(fmt.Sprintf("Servers (%d)", len(m.servers))),
		div,
		fmt.Sprintf("%-4s %-20s %-25s %-10s %-8s",
			style.StyleTableHeader.Render(""),
			style.StyleTableHeader.Render("NAME"),
			style.StyleTableHeader.Render("HOST"),
			style.StyleTableHeader.Render("PORT"),
			style.StyleTableHeader.Render("LOCALITY"),
		),
	}

	for i, srv := range m.servers {
		rowStyle := style.StyleTableRow
		prefix := "   "
		if i == m.selected {
			rowStyle = lipgloss.Style{}.Foreground(style.ColorPrimary).Bold(true)
			prefix = " ▶ "
		}
		port := fmt.Sprintf("%d", srv.Port)
		if srv.Port == 0 {
			port = "22"
		}
		lines = append(lines, fmt.Sprintf("%s%-20s %-25s %-10s %-8s",
			prefix,
			rowStyle.Render(trunc(srv.Name, 20)),
			rowStyle.Render(trunc(srv.Host, 25)),
			rowStyle.Render(port),
			rowStyle.Render(trunc(srv.Locality, 8)),
		))
	}

	if len(m.servers) == 0 {
		lines = append(lines, "", style.StyleMuted.Render("  No servers yet. Press [A] to add one."))
	}

	lines = append(lines, "")
	lines = append(lines, style.StyleMuted.Render("  [Enter] mosh/ssh   [A] add   [E] edit   [D] delete   [Esc] quit"))

	if m.errMsg != "" {
		lines = append(lines, "", style.Error(m.errMsg))
	}
	if m.successMsg != "" {
		lines = append(lines, "", style.Success(m.successMsg))
	}

	return strings.Join(lines, "\n")
}

func trunc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// Messages (package-local)
type LoadMsg struct{ Servers []config.Server }
type LoadErrMsg struct{ Err error }
