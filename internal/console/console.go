package console

import (
	"errors"
	"fmt"
	netfs "netfs/internal"
	"strings"
)

// -------------------------------------------------------- PUBLIC CODE ---------------------------------------------------------

var NeedHelpError = errors.New("help")
var CommandNotFoundError = errors.New("command not found")

// The command shows instructions for the console client.
var HelpCommand = _HelpConsoleCommand{}

// Abstraction of console command.
type ConsoleCommand interface {
	// Returns the command name.
	GetName() string
	// Returns information about command.
	GetDescription() string
	// Executes a command with arguments.
	Execute(args ...string) (string, error)
}

// A client for interacting with netfs via the console.
type ConsoleClient struct {
	_Commands []ConsoleCommand
	_Config   *netfs.Config
}

// Returns the command by name and returns an error if the command is not found.
func (client *ConsoleClient) GetCommand(name string) (ConsoleCommand, error) {
	for _, command := range client._Commands {
		if name == command.GetName() {
			return command, nil
		}
	}
	return nil, CommandNotFoundError
}

// Creates a new instance of ConsoleClient, returns an error if creation failed.
func NewConsoleClient(config *netfs.Config) (*ConsoleClient, error) {
	_, err := netfs.NewNetwork(config)
	if err == nil {
		client := &ConsoleClient{}
		client._Config = config
		client._Commands = []ConsoleCommand{HelpCommand}

		return client, nil
	}
	return nil, err
}

// -------------------------------------------------------- PRIVATE CODE --------------------------------------------------------

// The command shows instructions for the console client.
type _HelpConsoleCommand struct{}

// Returns the command name.
func (cmd _HelpConsoleCommand) GetName() string {
	return "help"
}

// Returns information about command.
func (cmd _HelpConsoleCommand) GetDescription() string {
	return strings.Join(
		[]string{
			"help [command]     - shows all available commands for the console client or information about a specific command.",
			"Examples:",
			"netfs help         - shows all available commands for the console client",
			"netfs help command - shows information about a specific command",
		},
		fmt.Sprintln(),
	)
}

// Executes a command with arguments.
func (cmd _HelpConsoleCommand) Execute(args ...string) (string, error) {
	if len(args) > 0 {
		if args[0] == cmd.GetName() {
			return cmd.GetDescription(), nil
		} else {
			// TODO. Find the command and show information about it
		}
	}
	return "", nil
}
