package main

import (
	"fmt"
	"strings"
	"io"
	"bufio"
	"os"
	"net/http"
	"encoding/json"
	"time"
	"github.com/greencaleb3/pokedexcli/internal/pokecache"
	"math/rand"
)

type cliCommand struct {
	name string
	description string
	callback func(con *config, param string) error
}

type config struct {
	url string
	currOffset int
	cache *pokecache.Cache
	caughtPokemon map[string]Pokemon
}

type Loc struct {
    Results []struct {
        Name string `json:"name"`
        URL  string `json:"url"`
    } `json:"results"`
    Next     string `json:"next"`
    Previous string `json:"previous"`
}

type LocationArea struct {
    PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"`
}

type PokemonEncounter struct {
	Pokemon struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"pokemon"`
	VersionDetails []struct {
		EncounterDetails []struct {
			Chance          int   `json:"chance"`
			ConditionValues []any `json:"condition_values"`
			MaxLevel        int   `json:"max_level"`
			Method          struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"method"`
			MinLevel int `json:"min_level"`
		} `json:"encounter_details"`
		MaxChance int `json:"max_chance"`
		Version   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"version"`
	} `json:"version_details"`
}

type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	IsDefault      bool   `json:"is_default"`
	Order          int    `json:"order"`
	Weight         int    `json:"weight"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`

}

func main() {
	scan := bufio.NewScanner(os.Stdin)
	cliCommand := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:	"help",
			description: "Describe the Pokedex",
			callback: commandHelp,
		},
		"map": {
			name:	"map",
			description: "Display the next 20 Pokemon Locs",
			callback: commandMap,
		},
		"mapb": {
			name:	"mapb",
			description: "Display the prev 20 Pokemon Locs",
			callback: commandMapb,
		},
		"explore": {
			name:	"explore",
			description: "Display the Pokemon at Locs",
			callback: commandExplore,
		},
		"catch": {
			name:	"catch",
			description: "Catch pokemon",
			callback: commandCatch,
		},
		"inspect" : {
			name: "inspect",
			description: "Inspect caught pokemon",
			callback: commandInspect,
		},
		"pokedex" : {
			name: "pokedex",
			description: "Display all pokemon caught",
			callback: commandPokedex,
		},
	}
	
	con := config{
		url: "https://pokeapi.co/api/v2/location-area/",
		currOffset: 0,
		cache: pokecache.New(5 * time.Second),
    		caughtPokemon: make(map[string]Pokemon),
	}
	for {
		fmt.Print("Pokedex > ")
		scan.Scan() 
		text := cleanInput(scan.Text())
		command := text[0]

		param := ""
		if len(text) > 1 { 
			param = text[1]
		}
		if cmd, ok := cliCommand[command]; ok {
			cmd.callback(&con, param)
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func cleanInput(text string) []string {
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	substring := strings.Split(text, " ")

	return substring
}

func commandExit(con *config, param string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(con *config, param string) error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nexit: Exit the Pokedex\n")
	return nil
}

func commandMap(con *config, param string) error {
	fullUrl := con.url + "?limit=20&offset=" + fmt.Sprintf("%d", con.currOffset)
	
	if cachedData, found := con.cache.Get(fullUrl); found {
		loc := Loc{}
		err := json.Unmarshal(cachedData, &loc)
		if err != nil {
			return err
		}

		for _, result := range(loc.Results) {
			fmt.Println(result.Name)
		}

		return nil
	}

	res, err := http.Get(fullUrl)
	if err != nil {
		return fmt.Errorf("Failed to get map: %s", err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("Failed to read response: %s", err)
	}

	if res.StatusCode > 299 {
		return fmt.Errorf("Status Code: %d", res.StatusCode)
	}
	
	con.currOffset += 20

	con.cache.Add(fullUrl, body)	

	loc := Loc{}
	json.Unmarshal(body, &loc)

	for _, result := range(loc.Results) {
		fmt.Println(result.Name)
	}

	return nil
}

func commandMapb(con *config, param string) error {
	if (con.currOffset == 0) {
		fmt.Println("you're on the first page")
		return nil
	}

	con.currOffset -= 20

	fullUrl := con.url + "?limit=20&offset=" + fmt.Sprintf("%d", con.currOffset)

	if cachedData, found := con.cache.Get(fullUrl); found {
		loc := Loc{}
		err := json.Unmarshal(cachedData, &loc)
		if err != nil {
			return err
		}

		for _, result := range(loc.Results) {
			fmt.Println(result.Name)
		}

		return nil
	}

	res, err := http.Get(fullUrl)
	if err != nil {
		return fmt.Errorf("Failed to get map: %s", err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("Failed to read response: %s", err)
	}

	if res.StatusCode > 299 {
		return fmt.Errorf("Status Code: %d", res.StatusCode)
	}

	con.cache.Add(fullUrl, body)

	loc := Loc{}
	json.Unmarshal(body, &loc)

	for _, result := range(loc.Results) {
		fmt.Println(result.Name)
	}

	return nil
}

func commandExplore(con *config, param string) error {
	fullUrl := con.url + param 
	
	if cachedData, found := con.cache.Get(fullUrl); found {
		locArea := LocationArea{}
		err := json.Unmarshal(cachedData, &locArea)
		if err != nil {
			return err
		}

		for _, result := range(locArea.PokemonEncounters) {
			fmt.Println(result.Pokemon.Name)
		}

		return nil
	}

	res, err := http.Get(fullUrl)
	if err != nil {
		return fmt.Errorf("Failed to get map: %s", err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("Failed to read response: %s", err)
	}

	if res.StatusCode > 299 {
		return fmt.Errorf("Status Code: %d", res.StatusCode)
	}
	
	con.cache.Add(fullUrl, body)

	locArea := LocationArea{}
	err = json.Unmarshal(body, &locArea)
	if err != nil {
		return err
	}

	for _, result := range(locArea.PokemonEncounters) {
		fmt.Println(result.Pokemon.Name)
	}

	return nil
}

func commandCatch(con *config, param string) error {
	fmt.Printf("Throwing a Pokeball at %s...\n", param)

	fullUrl := "https://pokeapi.co/api/v2/pokemon/" + param 
	
	res, err := http.Get(fullUrl)
	if err != nil {
		return fmt.Errorf("Failed to get map: %s", err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("Failed to read response: %s", err)
	}

	if res.StatusCode > 299 {
		return fmt.Errorf("Status Code: %d", res.StatusCode)
	}
	
	//con.cache.Add(fullUrl, body)

	pokemon := Pokemon{}
	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		return err
	}

    	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(pokemon.BaseExperience)
	
	if randomNum > pokemon.BaseExperience / 2 {
		fmt.Println(pokemon.Name + " escaped!")
	} else {
		fmt.Println(pokemon.Name + " was caught!")
		con.caughtPokemon[pokemon.Name] = pokemon 
	}

	return nil
}

func commandInspect(con *config, param string) error {
	poke, exists := con.caughtPokemon[param]
	if !exists {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Println("Name: " + poke.Name)
	fmt.Printf("Height: %d\n", poke.Height)
	fmt.Printf("Weight: %d\n", poke.Weight)
	fmt.Println("Stats:")

	for _, stat := range(poke.Stats) {
		fmt.Printf(" -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, breed := range(poke.Types) {
		fmt.Printf(" -%s\n", breed.Type.Name)
	}

	return nil
}


func commandPokedex(con *config, param string) error {
	fmt.Println("Your Pokedex:")
	for _, pokemon := range(con.caughtPokemon) {
		fmt.Printf(" - %s\n", pokemon.Name)
	}

	return nil
}
