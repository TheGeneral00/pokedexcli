package main

import (
    "fmt";
    "bufio";
    "os"
)


type cliCommand struct {
    name string
    description string 
    callback func() error
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
