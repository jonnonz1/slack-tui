package app

import "github.com/charmbracelet/lipgloss"

// Theme defines the MONOSPACE_CMD color palette.
// Derived from the design beads — Slack purple + pink + green on deep dark.
type Theme struct {
	Primary          lipgloss.Color
	PrimaryContainer lipgloss.Color
	Secondary        lipgloss.Color
	Surface          lipgloss.Color
	SurfaceContainer lipgloss.Color
	SurfaceLow       lipgloss.Color
	SurfaceLowest    lipgloss.Color
	SurfaceHigh      lipgloss.Color
	OnSurface        lipgloss.Color
	OnSurfaceVariant lipgloss.Color
	Outline          lipgloss.Color
	Error            lipgloss.Color
	Username         lipgloss.Color
	Timestamp        lipgloss.Color
	AIHook           lipgloss.Color
	AIBorder         lipgloss.Color
}

var DefaultTheme = Theme{
	Primary:          lipgloss.Color("#f6afef"),
	PrimaryContainer: lipgloss.Color("#4a154b"),
	Secondary:        lipgloss.Color("#5edda0"),
	Surface:          lipgloss.Color("#10141a"),
	SurfaceContainer: lipgloss.Color("#1c2026"),
	SurfaceLow:       lipgloss.Color("#181c22"),
	SurfaceLowest:    lipgloss.Color("#0a0e14"),
	SurfaceHigh:      lipgloss.Color("#262a31"),
	OnSurface:        lipgloss.Color("#dfe2eb"),
	OnSurfaceVariant: lipgloss.Color("#d2c2cd"),
	Outline:          lipgloss.Color("#4f434c"),
	Error:            lipgloss.Color("#ffb4ab"),
	Username:         lipgloss.Color("#5edda0"),
	Timestamp:        lipgloss.Color("#666666"),
	AIHook:           lipgloss.Color("#f6afef"),
	AIBorder:         lipgloss.Color("#4a154b"),
}
