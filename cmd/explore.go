package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"petezalew.ski/zip/zipfile"
)

var exploreCmd = &cobra.Command{
	Use:   "explore",
	Short: "Display the structure and metadata of zip files",
	Long:  "Print the local headers, central directory, and end record of one or more zip files.",
	Run:   runExplore,
}

var baseStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder())

type model struct {
	zipfile *zipfile.ZipFile
	table   table.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func initialModel(zf *zipfile.ZipFile) model {
	columns := []table.Column{
		{Title: "Name", Width: 78},
	}
	rows := make([]table.Row, len(zf.CentralDirectory))
	for i, file := range zf.CentralDirectory {
		rows[i] = table.Row{
			file.FileName,
		}
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	t.SetStyles(table.DefaultStyles())

	return model{
		zipfile: zf,
		table:   t,
	}
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

		p := tea.NewProgram(initialModel(zf))
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
