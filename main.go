package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type extraCreditType int
type site int

const (
	bitwise extraCreditType = iota
	compound
	increment
	goto_
)
const (
	selectExtraCreditFeature site = iota
	saveDialog
	run
	ran
)

func (ect *extraCreditType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	default:
		return fmt.Errorf("unknown extraCreditType %v", s)
	case "bitwise":
		*ect = bitwise
	case "compound":
		*ect = compound
	case "increment":
		*ect = increment
	case "goto":
		*ect = goto_
	}
	return nil
}

func (ect extraCreditType) MarshalJSON() ([]byte, error) {
	var s string
	switch ect {
	default:
		return []byte{}, fmt.Errorf("unkown extraCreditType %v", ect)
	case bitwise:
		s = "Bitwise"
	case compound:
		s = "Compound"
	case increment:
		s = "Increment"
	case goto_:
		s = "Goto"
	}

	return json.Marshal(s)
}

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
	SelectedExtraCredits []extraCreditType
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
		case "enter":
			m.s = saveDialog
			m.cursor = 0
		default:
			fmt.Println(msg.String())
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

		case "enter":
			m.s = run

			if m.cursor == 0 {
				selectedExtraCredits := make([]extraCreditType, 0)

				for i, extraCreditFeature := range m.extraCredits {
					_, ok := m.selected[i]
					if ok {
						selectedExtraCredits = append(selectedExtraCredits, extraCreditFeature.t)
					}
				}

				s := settings{SelectedExtraCredits: selectedExtraCredits}

				toWrite, err := json.Marshal(s)
				if err != nil {
					fmt.Printf("Error: could not marshal settings %v", err)
					return nil, tea.Quit
				}

				err = os.WriteFile(".wacc", toWrite, 0o775)
				if err != nil {
					fmt.Printf("Failed to write file \".wacc\", error: %v", err)
					return nil, tea.Quit
				}
			}
		}
	}

	return m, nil
}

func (m model) UpdateRun(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			cmd := exec.Command("../../rust/writing-a-c-compiler-tests/test_compiler")

			for i, choice := range m.extraCredits {
				_, ok := m.selected[i]
				if ok {
					cmd.Args = append(cmd.Args, choice.cmdLine)
				}
			}

			m.s = ran
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				fmt.Printf("Failed to execute test_compiler %v", err)
				return tea.Quit
			})
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
	case run:
		return m.UpdateRun(msg)
	case ran:
		return m, tea.Quit
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

	s += fmt.Sprintf("\n[%s] yes [%s] no\n", yes, no)

	s += fmt.Sprintf("\nPress q to quit, press Enter to continue.\n")

	return s
}

func (m model) ViewRun() string {
	s := "Are you sure, if you want to run the test compiler? [press enter to continue, q to exit]\n"

	return s
}

func (m model) View() string {

	switch m.s {
	case selectExtraCreditFeature:
		return m.ViewExtraCreditFeatures()
	case saveDialog:
		return m.ViewSaveDialog()
	case run:
		return m.ViewRun()
	case ran:
		return ""
	}

	panic("invalid app site")
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
