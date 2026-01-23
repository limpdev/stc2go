package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type customTheme struct {
	fyne.Theme
}

func newCustomTheme() fyne.Theme {
	return &customTheme{Theme: theme.DefaultTheme()}
}

func (ct *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Edit these color mappings as needed
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 70, G: 130, B: 180, A: 255} // Steel Blue
	case theme.ColorNameForeground:
		return color.NRGBA{R: 220, G: 220, B: 220, A: 255} // White text in entries
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255} // Dark background for entries
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 50, G: 205, B: 50, A: 255} // Lime Green
	case theme.ColorNameError:
		return color.NRGBA{R: 220, G: 20, B: 60, A: 255} // Crimson
	default:
		return ct.Theme.Color(name, variant)
	}
}
