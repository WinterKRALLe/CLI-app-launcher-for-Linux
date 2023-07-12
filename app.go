package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type Application struct {
	Name        string
	DesktopFile string
	Commands    []string
}

func main() {
	// Get a list of all desktop entry files
	desktopFiles, err := filepath.Glob("/usr/share/applications/*.desktop")
	if err != nil {
		fmt.Println("Error finding desktop entry files:", err)
		os.Exit(1)
	}

	// Check if any desktop entry files were found
	if len(desktopFiles) == 0 {
		fmt.Println("No executable applications found.")
		os.Exit(1)
	}

	// Display the list of applications with numbers for selection
	fmt.Println("Select an application to launch:")
	fmt.Println("-------------------------------")

	apps := make(map[int]Application)
	for i, desktopFile := range desktopFiles {
		appName, err := getAppName(desktopFile)
		if err != nil {
			fmt.Printf("Error reading desktop file '%s': %v\n", desktopFile, err)
			continue
		}

		commands, err := getAppCommands(desktopFile)
		if err != nil {
			fmt.Printf("Error getting commands for '%s': %v\n", desktopFile, err)
			continue
		}

		app := Application{
			Name:        appName,
			DesktopFile: desktopFile,
			Commands:    commands,
		}

		apps[i+1] = app
		fmt.Printf("%d. %s\n", i+1, appName)
	}

	// Prompt the user to choose an application
	fmt.Print("Enter the number of the application to launch: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()

	// Validate the user's choice
	appChoice := 0
	_, err = fmt.Sscanf(choice, "%d", &appChoice)
	if err != nil || appChoice < 1 || appChoice > len(apps) {
		fmt.Println("Invalid choice. Exiting.")
		os.Exit(1)
	}

	// Launch the selected application
	selectedApp := apps[appChoice]
	if len(selectedApp.Commands) == 1 {
		// Only one command, launch it directly
		command := selectedApp.Commands[0]
		fmt.Println("Launching", command)
		cmd := exec.Command("bash", "-c", command)

		// Detach the process from the script's process group
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}

		err = cmd.Start()
		if err != nil {
			fmt.Println("Error launching application:", err)
			os.Exit(1)
		}
	} else {
		// Multiple commands, present options to the user
		fmt.Println("Select an option to launch:")
		fmt.Println("-------------------------------")
		for i, command := range selectedApp.Commands {
			fmt.Printf("%d. %s\n", i+1, command)
		}

		// Prompt the user to choose an option
		fmt.Print("Enter the number of the option to launch: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		optionChoice := scanner.Text()

		// Validate the user's option choice
		option := 0
		_, err := fmt.Sscanf(optionChoice, "%d", &option)
		if err != nil || option < 1 || option > len(selectedApp.Commands) {
			fmt.Println("Invalid option choice. Exiting.")
			os.Exit(1)
		}

		// Launch the selected option
		command := selectedApp.Commands[option-1]
		fmt.Println("Launching", command)
		cmd := exec.Command("bash", "-c", command)

		// Detach the process from the script's process group
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}

		err = cmd.Start()
		if err != nil {
			fmt.Println("Error launching option:", err)
			os.Exit(1)
		}
	}
}

func getAppName(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name=") {
			return strings.TrimPrefix(line, "Name="), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("application name not found in desktop file")
}

func getAppCommands(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Exec=") {
			command := strings.TrimPrefix(line, "Exec=")
			command = strings.ReplaceAll(command, "%u", "")
			command = strings.ReplaceAll(command, "%F", "")
			command = strings.TrimSpace(command)
			commands = append(commands, command)
		}
	}

	return commands, nil
}
