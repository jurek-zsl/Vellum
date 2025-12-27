package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jurekzsl/vellum/internal/tui"
)

func main() {
	if os.Geteuid() != 0 {
		exe, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting executable path: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("Vellum requires root privileges. Requesting sudo...")
		
		cmd := exec.Command("sudo", append([]string{exe}, os.Args[1:]...)...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			// If sudo fails or user cancels, exit
			os.Exit(1)
		}
		// Exit the parent process successfully after child (sudo) finishes
		return
	}

	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
