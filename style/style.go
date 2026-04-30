package style

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette — nightshade/mosh vibe
var (
	ColorBackground = lipgloss.Color("#0d0d0d")
	ColorSurface    = lipgloss.Color("#1a1a2e")
	ColorBorder     = lipgloss.Color("#2d2d44")
	ColorPrimary    = lipgloss.Color("#7c3aed") // violet
	ColorAccent     = lipgloss.Color("#a78bfa") // light violet
	ColorSuccess    = lipgloss.Color("#34d399")
	ColorDanger     = lipgloss.Color("#f87171")
	ColorWarning    = lipgloss.Color("#fbbf24")
	ColorMuted      = lipgloss.Color("#6b7280")
	ColorText       = lipgloss.Color("#e2e8f0")
	ColorBright     = lipgloss.Color("#ffffff")
)

// Text styles
var (
	StyleTitle = lipgloss.NewStyle().
			Foreground(ColorBright).
			Bold(true).
			Padding(0, 1)

	StyleHeader = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleLabel = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleValue = lipgloss.NewStyle().
			Foreground(ColorText)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StyleDanger = lipgloss.NewStyle().
			Foreground(ColorDanger)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorWarning)

	StyleBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2).
			Foreground(ColorText)

	StyleTableHeader = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Bold(true)

	StyleTableRow = lipgloss.NewStyle().
			Foreground(ColorText)

	StyleInput = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1).
			Foreground(ColorText)

	StyleInputFocused = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1).
				Foreground(ColorBright)

	StyleButton = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorBright).
			Padding(0, 2).
			Margin(0, 1)

	StyleButtonDanger = lipgloss.NewStyle().
				Background(ColorDanger).
				Foreground(ColorBright).
				Padding(0, 2)
)

// RenderDivider renders a horizontal divider
func RenderDivider(width int) string {
	div := ""
	for i := 0; i < width; i++ {
		div += "─"
	}
	return lipgloss.Style{}.Foreground(ColorBorder).Render(div)
}

// AppMargin is the horizontal padding around the TUI content
const AppMargin = 2

func Success(msg string) string { return StyleSuccess.Render("✓ ") + msg }
func Error(msg string) string   { return StyleDanger.Render("✗ ") + msg }
func Info(msg string) string    { return lipgloss.Style{}.Foreground(ColorPrimary).Render("ℹ ") + msg }
func Warn(msg string) string    { return StyleWarning.Render("⚠ ") + msg }
