package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ja8mpi/go-pokecache"
)

type Pokemon struct {
	BaseExperience int    `json:"base_experience"`
	Weight         int    `json:"weight"`
	Height         int    `json:"height"`
	Stats          []Stat `json:"stats"`
	Types          []Type `json:"types"`
}

type Type struct {
	Slot int      `json:"slot"`
	Type TypeInfo `json:"type"`
}

type TypeInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Stat struct {
	BaseStat int        `json:"base_stat"`
	Effort   int        `json:"effort"`
	StatInfo StatDetail `json:"stat"`
}

type StatDetail struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

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

type LocationDetails struct {
	PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"`
}

type PokemonEntry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PokemonEncounter struct {
	Pokemon PokemonEntry `json:"pokemon"`
}

type config struct {
	Pokedex       map[string]Pokemon
	Next          string
	Previous      string
	locationCache pokecache.Cache
	pokemonCache  pokecache.Cache
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, ...string) error
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
		description: "Displays 20 location areas of the Pokemon world per call, showing the previous 20 locations on each subsequent call",
		callback:    commandMapb,
	},
	"explore": {
		name:        "explore",
		description: "Displays the pokemons that you may encounter in a given area",
		callback:    commandExplore,
	},
	"catch": {
		name:        "catch",
		description: "Catches a pokemon and adds them ot the user's pokedex.",
		callback:    commandCatch,
	},
	"inspect": {
		name:        "inspect",
		description: "It takes the name of a Pokemon and prints the name, height, weight, stats and type(s) of the Pokemon if it was caught before.",
		callback:    commandInspect,
	},
	"pokedex": {
		name:        "pokedex",
		description: "Prints the names of the pokemons a user has caught",
		callback:    commandPokedex,
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

func commandExit(cfg *config, params ...string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config, params ...string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:\n")

	for key, cmd := range commands {
		fmt.Printf("%s: %s\n", key, cmd.description)
	}

	return nil
}

func makeMapRequest(url string) {}

func commandMap(cfg *config, params ...string) error {
	data, exists := cfg.locationCache.Get(cfg.Next)
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

func commandMapb(cfg *config, params ...string) error {
	data, exists := cfg.locationCache.Get(cfg.Previous)
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

func commandExplore(cfg *config, params ...string) error {
	if len(params) < 1 {
		return fmt.Errorf("missing location parameter")
	}
	location := params[0]

	data, exists := cfg.pokemonCache.Get(location)
	if exists {
		var encounters []PokemonEncounter

		err := json.Unmarshal(data, &encounters)
		if err != nil {
			return fmt.Errorf("failed to parse cached encounters: %v", err)
		}

		fmt.Printf("Exploring %s...\n", location)
		fmt.Println("Found Pokemon:")
		for _, entry := range encounters {
			fmt.Printf(" - %v\n", entry.Pokemon.Name)
		}

		return nil
	}

	res, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%v", location))
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
	}

	var response LocationDetails
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	fmt.Println("Exploring pastoria-city-area...")
	fmt.Println("Found Pokemon:")

	for _, entry := range response.PokemonEncounters {

		fmt.Printf(" - %v\n", entry.Pokemon.Name)
	}

	body, err = json.Marshal(response.PokemonEncounters)
	if err != nil {
		return fmt.Errorf("failed to serialize Pokémon encounters: %v", err)
	}

	cfg.pokemonCache.Add(params[0], body)
	return nil
}

func commandCatch(cfg *config, params ...string) error {
	if len(params) < 1 {
		return fmt.Errorf("missing pokemon parameter")
	}
	name := params[0]

	res, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%v", name))
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode > 299 {
		return fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
	}

	var response Pokemon
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	fmt.Printf("Throwing a Pokeball at %v...\n", name)
	baseExp := response.BaseExperience

	// Calculate catch chance
	catchChance := 80.0 - float64(baseExp)*0.3
	if catchChance < 10.0 {
		catchChance = 10.0
	}

	// Roll a random number from 0.0 to 100.0
	roll := rand.Float64() * 100

	if roll <= catchChance {
		fmt.Printf("%v was caught!\n", name)

		cfg.Pokedex[name] = response
	} else {
		fmt.Printf("%v escaped!\n", name)
	}

	return nil
}

func commandPokedex(cfg *config, params ...string) error {
	if len(cfg.Pokedex) <= 0 {
		fmt.Println("Your pokedex is empty. Catch some pokemons!")
		return fmt.Errorf("Empty pokedex")
	}

	for key, _ := range cfg.Pokedex {
		fmt.Printf(" - %v\n", key)
	}

	return nil
}

func commandInspect(cfg *config, params ...string) error {
	if len(params) < 1 {
		return fmt.Errorf("missing pokemon parameter")
	}
	name := params[0]
	pokemon, exists := cfg.Pokedex[name]
	if !exists {
		fmt.Println("Pokemon not found in the map!")
		return fmt.Errorf("Pokemon not found in the map!")
	} else {
		fmt.Printf("Name: %s\n", name)
		fmt.Printf("Height: %d\n", pokemon.Height)
		fmt.Printf("Weight: %d\n", pokemon.Weight)

		for _, stat := range pokemon.Stats {
			fmt.Printf("  -%v: %v\n", stat.StatInfo.Name, stat.BaseStat)
		}
		fmt.Println("Types: ")

		for _, pokemonType := range pokemon.Types {
			fmt.Printf("  -%v\n", pokemonType.Type.Name)
		}
	}

	return nil
}

func main() {
	commands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
	}

	cfg := config{
		Previous:      "",
		Next:          "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20",
		locationCache: *pokecache.NewCache(5 * time.Minute),
		pokemonCache:  *pokecache.NewCache(5 * time.Minute),
		Pokedex:       map[string]Pokemon{},
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

		params := input[1:] // everything after the command
		if err := cmd.callback(&cfg, params...); err != nil {
			fmt.Println("Error:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
