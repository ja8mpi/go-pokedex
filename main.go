package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

var commands = map[string]cliCommand{
	"exit": {
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	},
	"map": {
		name:        "map",
		description: "Exit the Pokedex",
		callback:    commandExit,
	},
}

func cleanInput(text string) []string {
	data := strings.ToLower(text)
	words := strings.Fields(data)

	for i, word := range words {
		words[i] = strings.TrimSpace(word)
	}

	return words
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:\n")

	for key, cmd := range commands {
		fmt.Printf("%s: %s\n", key, cmd.description)
	}

	return nil
}

func main() {
	commands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}
		input := cleanInput(scanner.Text())

		cmd, ok := commands[input[0]]

		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		cmd.callback()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
