package selector

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
)

type Item struct {
	ID    string
	Label string
}

var ErrSelectionCanceled = fmt.Errorf("selection canceled")

func SelectOne(title string, items []Item) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items available")
	}
	if !isTerminalStdin() {
		return "", fmt.Errorf("interactive selection requires a terminal")
	}

	options := make([]huh.Option[string], 0, len(items))
	for _, item := range items {
		if item.ID == "" {
			continue
		}
		label := item.Label
		if label == "" {
			label = item.ID
		}
		options = append(options, huh.NewOption(label, item.ID))
	}
	if len(options) == 0 {
		return "", fmt.Errorf("no items available")
	}

	height := len(options) + 2
	if height > 20 {
		height = 20
	}
	if height < 6 {
		height = 6
	}

	var selected string
	err := huh.NewSelect[string]().
		Title(title).
		Height(height).
		Options(options...).
		Value(&selected).
		Run()
	if err != nil {
		if err == huh.ErrUserAborted {
			return "", ErrSelectionCanceled
		}
		return "", fmt.Errorf("interactive selection failed: %w", err)
	}

	if selected == "" {
		return "", fmt.Errorf("no item selected")
	}

	return selected, nil
}

func isTerminalStdin() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
