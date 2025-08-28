package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"petezalew.ski/zip/zipfile"
)

var exploreCmd = &cobra.Command{
	Use:   "explore",
	Short: "Display the structure and metadata of zip files",
	Long:  "Print the local headers, central directory, and end record of one or more zip files.",
	Run:   runExplore,
}

type model struct {
	zipfile *zipfile.ZipFile
	cursor  int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.cursor = max(0, m.cursor-1)
			return m, nil
		case "down", "j":
			m.cursor = min(len(m.zipfile.CentralDirectory)-1, m.cursor+1)
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	s := ""

	sliceEnd := min(m.cursor+10, len(m.zipfile.CentralDirectory))
	for _, header := range m.zipfile.LocalHeaders[m.cursor:sliceEnd] {
		s += fmt.Sprintf("%+q\n", header.Name)
	}

	return s
}

func runExplore(cmd *cobra.Command, args []string) {
	for _, filename := range args {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer f.Close()
		zf, err := zipfile.Parse(f)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		p := tea.NewProgram(model{cursor: 0, zipfile: zf})
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
