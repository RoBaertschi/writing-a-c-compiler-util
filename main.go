package main

import (
	"encoding/json"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type extraCreditType int
type site int

const (
	bitwise extraCreditType = iota
	compound
	increment
	goto_

	selectExtraCreditFeature site = iota
	saveDialog
	run
)

func (ect extraCreditType) data() extraCreditFeature {
	switch ect {
	case bitwise:
		return extraCreditFeature{t: bitwise, name: "Bitwise Operations", cmdLine: "--bitwise"}
	case compound:
		return extraCreditFeature{t: compound, name: "Compound", cmdLine: "--compound"}
	case increment:
		return extraCreditFeature{t: increment, name: "Increment and Decrement", cmdLine: "--increment"}
	case goto_:
		return extraCreditFeature{t: goto_, name: "Goto statement", cmdLine: "--goto"}
	}

	panic("Invalid extraCreditType")
}

type settings struct {
	selectedExtraCredits []extraCreditType
}

type extraCreditFeature struct {
	t       extraCreditType
	name    string
	cmdLine string
}

type model struct {
	extraCredits []extraCreditFeature
	selected     map[int]struct{}
	cursor       int
	s            site
}

func initialModel() model {
	extraCreditFeatures := make([]extraCreditFeature, 0)
	extraCreditFeatures = append(extraCreditFeatures, bitwise.data())
	extraCreditFeatures = append(extraCreditFeatures, compound.data())
	extraCreditFeatures = append(extraCreditFeatures, increment.data())
	extraCreditFeatures = append(extraCreditFeatures, goto_.data())

	return model{
		extraCredits: extraCreditFeatures,
		selected:     make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) UpdateExtraCreditFeatures(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "j", "down":
			if m.cursor < len(m.extraCredits) {
				m.cursor++
			}

		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "Enter":
			m.s = saveDialog
			m.cursor = 0
		}
	}

	return m, nil
}

// cursor = 0 => yes
// cursor = 1 => no
func (m model) UpdateSaveDialog(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "h", "left":
			if m.cursor == 1 {
				m.cursor = 0
			}
		case "l", "right":
			if m.cursor == 0 {
				m.cursor = 1
			}

		case "Enter":
			m.s = run

			if m.cursor == 0 {
				selectedExtraCredits := make([]extraCreditType, 0)

				for i, extraCreditFeature := range m.extraCredits {
					_, ok := m.selected[i]
					if ok {
						selectedExtraCredits = append(selectedExtraCredits, extraCreditFeature.t)
					}
				}

				s := settings{selectedExtraCredits: selectedExtraCredits}

				toWrite, err := json.Marshal(s)
				if err != nil {
					fmt.Printf("Error: could not marshal settings %v", err)
					return nil, tea.Quit
				}

				err = os.WriteFile(".wacc", toWrite, 0)
				if err != nil {
					fmt.Printf("Failed to write file \".wacc\", error: %v", err)
					return nil, tea.Quit
				}
			}
		}
	}

	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.s {
	case selectExtraCreditFeature:
		return m.UpdateExtraCreditFeatures(msg)
	case saveDialog:
		return m.UpdateSaveDialog(msg)
	}

	return m, nil
}

func (m model) ViewExtraCreditFeatures() string {
	s := "Select the extra credit features that you want.\n\n"

	for i, extraCextraCreditFeature := range m.extraCredits {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, extraCextraCreditFeature.name)
	}

	s += "\nPress Enter to continue.\nPress q to quit.\n"
	return s
}

func (m model) ViewSaveDialog() string {
	s := "Do you want to save these extra credit features to a json file?"

	yes := " "
	if m.cursor == 0 {
		yes = "x"
	}

	no := " "
	if m.cursor == 1 {
		no = "x"
	}

	s += fmt.Sprintf("[%s] yes [%s] no\n", yes, no)

	s += fmt.Sprintf("\nPress q to quit, press Enter to continue.\n")

	return s
}

func (m model) View() string {

	switch m.s {
	case selectExtraCreditFeature:
		return m.ViewExtraCreditFeatures()
	case saveDialog:
		return m.ViewSaveDialog()
	}

	return "invalid app site"
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
