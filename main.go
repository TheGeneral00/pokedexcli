package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/TheGeneral00/pokedexcli/internal"
)


type cliCommand struct {
    name string
    description string 
    callback func(*config) error
}

type config struct {
    cache *internal.Cache
    pokedex *internal.Pokedex
    prev string
    next string
    current string
    currentLocation string
    additionalInput string
}

type Response struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type exploreResponse struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	GameIndex            int    `json:"game_index"`
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	Location struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Names []struct {
		Name     string `json:"name"`
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
			MaxChance        int `json:"max_chance"`
			EncounterDetails []struct {
				MinLevel        int   `json:"min_level"`
				MaxLevel        int   `json:"max_level"`
				ConditionValues []any `json:"condition_values"`
				Chance          int   `json:"chance"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
			} `json:"encounter_details"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

func getCommands() map[string]cliCommand {
    return map[string]cliCommand{
        "help": {
            name:   "help",
            description:    "Displayes a help message",
            callback:       commandHelp,
        },
        "exit": {
            name:           "exit",
            description:    "Exit the Pokedex",
            callback:       commandExit,
        },
        "map": {
            name:           "map ",
            description:    "Displays the name of 20 locations",
            callback:       commandMap,
        },
        "mapb": {
            name:           "mapb",
            description:    "Displays the locations from the previous map call",
            callback:       commandMapB,
        },
        "printResponse": {
            name:           "printResponse",
            description:    "Debugging function that prints the response of a http request",
            callback:       printResponse,
        },
        "explore": {
            name:           "explore",
            description:    "Shows the pokemon located in the area",
            callback:       commandExplore,
        },
        "catch":    {
            name:           "catch",
            description:    "Allows you to try to catch the pokemon discovered by exploring the area",
            callback:       commandCatch,
        },
        "inspect":  {
            name:           "inspect",
            description:    "Gives detailed information about the pokemon in your pokedex",
            callback:       commandInspect,
        },
        "pokedex":  {
            name:           "pokedex",
            description:    "Lists all your caught pokemon",
            callback:       commandPokedex,
        },
    }
}

func commandExit(*config) error {
    fmt.Println("Exiting program")
    os.Exit(0)
    return nil
}

func commandHelp(*config) error {
    commands := getCommands()
    fmt.Println("Welcome to the Pokedex!\n\nUsage:\n")
    for name, content := range commands {
        fmt.Println(name, ":", content.description)
    }
    fmt.Println("\n")
    return nil
}

func commandMap(config *config) error {
    res, err := http.Get("https://pokeapi.co/api/v2/location-area?offset=0&limit=20")
    if config.next != ""{
        res, err = http.Get(config.next)
    }
    if err != nil {
        return fmt.Errorf("Locations couldn't be displayed with error: %v", err)
    }
    defer res.Body.Close()
    if res.StatusCode > 299 {
        return fmt.Errorf("Response failed with status code: %d", res.StatusCode)
    }
    var response Response
    rawByteSlice, err := io.ReadAll(res.Body)
    if err != nil {
        return fmt.Errorf("Failed reading response Body to byte slice with error: %v", err)
    }
    err = json.Unmarshal(rawByteSlice, &response)
    if err != nil {
        return fmt.Errorf("Failed unmarshaling response with error: %v", err)
    }
    for _, location := range response.Results {
        fmt.Println(location.Name)
    }
    
    config.next = response.Next
    config.prev = response.Previous
    config.current = res.Request.URL.String()
    config.cache.Add(config.current, rawByteSlice)
    fmt.Println("Storing URL:", res.Request.URL.String()) 
    return nil
}

func commandMapB(config *config) error {
    if config.prev == ""{
        return fmt.Errorf("There are no locations to go back to")
    }
    fmt.Printf("Requested URL: %v\n", config.prev)
    var response Response
    if val, ok := config.cache.Get(config.prev); ok {
        err := json.Unmarshal(val, &response)    
        if err != nil {
            return fmt.Errorf("Failed to unmarshal the raw byte val from cache with error: %v", err)
        }
    } else {
        res, err := http.Get(config.prev)
        defer res.Body.Close()
        if err != nil {
            return fmt.Errorf("Response failed with error: %v", err)
        }
        if res.StatusCode > 299 {
            return fmt.Errorf("Response failed with status code: %d", res.StatusCode)
        }
        rawByteBody, err := io.ReadAll(res.Body)
        if err != nil {
            return fmt.Errorf("Failed to readd responde body with error: %v", err)
        }
        err = json.Unmarshal(rawByteBody, &response)
        if err != nil {
            return fmt.Errorf("Failed to unmarshal rawByteBody with error: %v", err)
        }
        config.cache.Add(res.Request.URL.String(), rawByteBody)
    }
    
    for _, location := range response.Results {
        fmt.Println(location.Name)
    }
    config.next = response.Next
    config.current = config.prev
    config.prev = response.Previous 
    return nil 
}

func printResponse(config *config) error {
    res, err := http.Get("https://pokeapi.co/api/v2/location-area")
    if err != nil {
        return fmt.Errorf("Response failed with error: %v", err)
    }
    if res.StatusCode > 299 {
        return fmt.Errorf("Response failed with status code: %v\nand body: \n", res.StatusCode, res.Body)
    }
    defer res.Body.Close()
    bodyBytes, err := io.ReadAll(res.Body)
    if err != nil {
        return fmt.Errorf("Reading of response body failed")
    }
    fmt.Printf("%s", res.Request.URL.String())
    fmt.Printf("%s\n", string(bodyBytes))
    return nil
}

func commandExplore(config *config) error {
    fmt.Printf("Exploring %v\n", config.additionalInput)
    locationURL := "https://pokeapi.co/api/v2/location-area/" + config.additionalInput 
    var response exploreResponse
    if entry, ok := config.cache.Get(locationURL); ok {
        err := json.Unmarshal(entry, &response)    
        if err != nil {
            return fmt.Errorf("Failed to unmarshal the raw byte val from cache with error: %v", err)
        } 
    } else {
        res, err := http.Get(locationURL)
        if err != nil {
            return fmt.Errorf("Request failed with error: %v", err)
        }
        defer res.Body.Close()
        rawByteBody, err := io.ReadAll(res.Body)
        if err != nil {
            return fmt.Errorf("Failed to read response body with error: %v", err)
            }
        err = json.Unmarshal(rawByteBody, &response)
        if err != nil {
            return fmt.Errorf("Failed to unmarshal the response body with err: %v", err)
        }
        config.cache.Add(res.Request.URL.String(), rawByteBody)
    }
    for _, encounter := range response.PokemonEncounters {
        fmt.Printf(" - %v\n", encounter.Pokemon.Name)
    }
    config.currentLocation = locationURL
    return nil 
}

func commandCatch(config *config) error {
    fmt.Printf("Throwing a Pokeball at %v ...\n", config.additionalInput)
    pokemonURL := "https://pokeapi.co/api/v2/pokemon/" + config.additionalInput 
    rawLocation, _ := config.cache.Get(config.currentLocation)
    var location exploreResponse
    err := json.Unmarshal(rawLocation, &location)
    if err != nil {
        return fmt.Errorf("Failed to unmarshal cached data with error: %v", err)
    }
    localPokemon := false
    for _, encounter := range location.PokemonEncounters {
        if encounter.Pokemon.Name == config.additionalInput {
            localPokemon = true
        }
    }

    if !localPokemon {
        return fmt.Errorf("%v is not present in the area", config.additionalInput)
    }  
    if _, ok := config.pokedex.Entries[config.additionalInput]; ok{
        return fmt.Errorf("%v has allready been caught", config.additionalInput)
    } else {
        res, err := http.Get(pokemonURL)
        if err != nil {
            return fmt.Errorf("Request failed with error: %v", err)
        }
        defer res.Body.Close()
        rawBytePokemon, err := io.ReadAll(res.Body)
        if err != nil {
            return fmt.Errorf("Failed to read response body with error: %v", err)
        }
        var pokemon internal.Pokemon
        err = json.Unmarshal(rawBytePokemon, &pokemon)
        if err != nil {
            return fmt.Errorf("Failed to unmarshal with error: %v", err)
        }
        config.pokedex.Add(pokemon)
    }
    return nil 
}

func commandInspect(config *config) error {
    pokemon, err := config.pokedex.Get(config.additionalInput)
    if err != nil {
        return err
    }
    fmt.Printf("%v: %v\n", "Name", pokemon.Name)
    fmt.Printf("%v: %v\n", "Height", pokemon.Height)
    fmt.Printf("%v: %v\n", "Weight", pokemon.Weight)
    fmt.Printf("%v:\n", "Stats")
    for _, stat := range pokemon.Stats {
        fmt.Printf("    - %v: %v\n", stat.Stat.Name, stat.BaseStat)
    }
    fmt.Printf("%v:\n", "Types")
    for _, pokeType := range pokemon.Types {
        fmt.Printf("    - %v\n", pokeType.Type.Name)
    }
    return nil
}

func commandPokedex(config *config) error {
    fmt.Println("Your Pokedex:")
    return config.pokedex.Show() 
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    config := config{
        prev: "",
        next: "",
    }
    // .NewCache returns pointer to the created cache!
    config.cache = internal.NewCache(60)
    config.pokedex = internal.NewPokedex()
    commands := getCommands()
    for {
        fmt.Printf("pokedex > ")
        if scanner.Scan() {
            input := strings.Split(scanner.Text(), " ")
            if len(input) > 1{
                config.additionalInput = strings.TrimSpace(input[1])
            }
            if _, ok := commands[input[0]]; ok {
                if err := commands[input[0]].callback(&config); err != nil {
                    fmt.Printf("%v\n", err)
                }
            } else {
                fmt.Printf("%v is not a valid command\n", scanner.Text())
            }    
        }
    }
}
