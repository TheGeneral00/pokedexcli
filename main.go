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
    cache *pokeCache.Cache
    prev string
    next string
    current string
    exploringArea string
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
    fmt.Printf("Exploring %v\n", config.exploringArea)
    res, err := http.Get("https://pokeapi.co/api/v2/location-area" + config.exploringArea)
    if err != nil {
        return fmt.Errorf("Request failed with error: %v", err)
    }
    defer res.Body.Close()
    var response exploreResponse
    rawByteBody, err := io.ReadAll(res.Body)
    if err != nil {
        return fmt.Errorf("Failed to read response body with error: %v", err)
    }
    err = json.Unmarshal(rawByteBody, &response)
    if err != nil {
        return fmt.Errorf("Failed to unmarshal the response body with err: %v", err)
    }
    for _, encounter := range response.PokemonEncounters {
        fmt.Printf(" - %v\n", encounter.Pokemon.Name)
    } 
    config.cache.Add(res.Request.URL.String(), rawByteBody)
    return nil 
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    config := config{
        prev: "",
        next: "",
    }
    // .NewCache returns pointer to the created cache!
    config.cache = pokeCache.NewCache(60)
    commands := getCommands()
    for {
        fmt.Printf("pokedex > ")
        if scanner.Scan() {
            input := strings.Split(scanner.Text(), " ")
            if len(input) > 1{
                config.exploringArea = input[1]
            }
            switch input[0] {
            case "help":
                commands["help"].callback(&config)
            case "exit":
                commands["exit"].callback(&config)
            case "map":
                commands["map"].callback(&config)
            case "mapb":
                commands["mapb"].callback(&config)
            case "printResponse":
                commands["printResponse"].callback(&config)
            case "explore":
                commands["explore"].callback(&config)
            default:
                fmt.Printf("%v is not a valid command", scanner.Text())
            }    
        }
    }
}
