package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ja8mpi/go-pokecache"
)

type pokeResponse struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous string      `json:"previous"`
	Results  []pokeEntry `json:"results"`
}

type pokeEntry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type config struct {
	Next     string
	Previous string
	cache    pokecache.Cache
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
	cfg         *config
}

var commands = map[string]cliCommand{
	"exit": {
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	},
	"map": {
		name:        "map",
		description: "Displays 20 location areas of the Pokemon world per call, showing the next 20 locations on each subsequent call",
		callback:    commandMap,
	},
	"mapb": {
		name:        "mapb",
		description: "Displays 20 location areas of the Pokemon world per call, showing the next 20 locations on each subsequent call",
		callback:    commandMapb,
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

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:\n")

	for key, cmd := range commands {
		fmt.Printf("%s: %s\n", key, cmd.description)
	}

	return nil
}

func makeMapRequest(url string) {}

func commandMap(cfg *config) error {
	data, exists := cfg.cache.Get(cfg.Next)
	if exists {
		var response pokeResponse

		err := json.Unmarshal(data, &response)

		if err != nil {
			return err
		} else {
			for _, loc := range response.Results {
				fmt.Println(loc.Name)
			}

			cfg.Next = response.Next
			cfg.Previous = response.Previous
		}
	}

	res, err := http.Get(cfg.Next)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
	}

	var response pokeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	for _, loc := range response.Results {
		fmt.Println(loc.Name)
	}

	cfg.Next = response.Next
	cfg.Previous = response.Previous

	return nil
}

func commandMapb(cfg *config) error {
	data, exists := cfg.cache.Get(cfg.Previous)
	if exists {
		var response pokeResponse

		err := json.Unmarshal(data, &response)

		if err != nil {
			return err
		} else {
			for _, loc := range response.Results {
				fmt.Println(loc.Name)
			}

			cfg.Next = response.Next
			cfg.Previous = response.Previous
		}
	}

	if cfg.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	res, err := http.Get(cfg.Previous)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
	}

	var response pokeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	for _, loc := range response.Results {
		fmt.Println(loc.Name)
	}

	cfg.Next = response.Next
	cfg.Previous = response.Previous

	return nil
}

func main() {
	commands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
	}

	cfg := config{
		Previous: "",
		Next:     "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20",
		cache:    *pokecache.NewCache(5 * time.Minute),
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}
		input := cleanInput(scanner.Text())

		if len(input) == 0 {
			continue
		}

		cmd, ok := commands[input[0]]

		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		if err := cmd.callback(&cfg); err != nil {
			fmt.Println("Error:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
