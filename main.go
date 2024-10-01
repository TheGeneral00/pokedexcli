package main

import (
    "fmt";
    "bufio";
    "os";
    "net/http";
    "encoding/json"
)


type cliCommand struct {
    name string
    description string 
    callback func() error
}

type Response struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous any    `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
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
            name:           "map",
            description:    "Displays the name of 20 locations",
            callback:       commandMap,
        },
        "mapb": {
            name:           "mapb",
            description:    "Displays the locations from the previous map call",
            callback:       commandMapB,
        },
    }
}

func commandExit() error {
    fmt.Println("Exiting program")
    os.Exit(0)
    return nil
}

func commandHelp() error {
    commands := getCommands()
    fmt.Println("Welcome to the Pokedex!\n\nUsage:\n")
    for name, content := range commands {
        fmt.Println(name, ":", content.description)
    }
    fmt.Println("\n")
    return nil
}

func commandMap() error {
    res, err := http.Get("https://pokeapi.co/api/v2/location")
    if err != nil {
        return fmt.Errorf("Locations couldn't be displayed with error: %v", err)
    }
    defer res.Body.Close()
    if res.StatusCode > 299 {
        return fmt.Errorf("Response failed with status code: %d", res.StatusCode)
    }
    var response Response
    dec := json.NewDecoder(res.Body)
    if err := dec.Decode(&response); err != nil {
        return fmt.Errorf("Failed with error: %v", err)
    }
    for _, location := range response.Results {
        fmt.Println(location.Name)
    }
    return nil
}

func commandMapB() error {
    fmt.Println("To be implemented")
    return nil
}


func main() {
    scanner := bufio.NewScanner(os.Stdin)
    commands := getCommands()
    for {
        fmt.Printf("pokedex > ")
        if scanner.Scan() {
            _, ok := commands[scanner.Text()]
            if !ok {
                fmt.Println("Invalid command")
                continue
            } 
            commands[scanner.Text()].callback()
        }
         
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "reading standard input:", err)
    }
}
