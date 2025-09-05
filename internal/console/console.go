package console

import (
	"errors"
	"fmt"
	netfs "netfs/internal"
	"strconv"
	"strings"
)

// -------------------------------------------------------- PUBLIC CODE ---------------------------------------------------------

var NeedHelpError = errors.New("help")
var CommandNotFoundError = errors.New("command not found")
var NoAvailableHosts = errors.New("No available hosts")

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
	_Config   *netfs.Config
	_Network  *netfs.Network
	_Commands []ConsoleCommand
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
	network, err := netfs.NewNetwork(config)
	if err == nil {
		client := &ConsoleClient{}
		client._Config = config
		client._Network = network
		client._Commands = []ConsoleCommand{
			_HelpConsoleCommand{_Client: client},
			_HostsConsoleCommand{_Client: client},
			_FileInfoConsoleCommand{_Client: client},
		}

		return client, nil
	}
	return nil, err
}

// -------------------------------------------------------- PRIVATE CODE --------------------------------------------------------

const PATH_SEPARATOR = "/"
const PATH_SIZE = 2

// The command shows information about file on remote host.
type _FileInfoConsoleCommand struct {
	_Client *ConsoleClient
}

// Returns the command name.
func (cmd _FileInfoConsoleCommand) GetName() string {
	return "file"
}

// Returns information about command.
func (cmd _FileInfoConsoleCommand) GetDescription() string {
	return strings.Join(
		[]string{
			"file host/file.txt              - shows information about file on remote host.",
			"Examples:",
			"netfs file 192.51.12.1/file.txt - shows information about file on remote host.",
			"netfs file myhostname/file.txt  - shows information about file on remote host.",
		},
		fmt.Sprintln(),
	)
}

// Executes a command with arguments.
func (cmd _FileInfoConsoleCommand) Execute(args ...string) (string, error) {
	err := NeedHelpError

	if len(args) > 0 {
		path := strings.SplitN(args[0], PATH_SEPARATOR, PATH_SIZE)
		if len(path) == PATH_SIZE {
			var hosts []netfs.RemoteHost
			if hosts, err = cmd._Client._Network.GetHosts(); err == nil {
				for _, host := range hosts {
					if path[0] == host.Name || path[0] == host.IP.String() {
						var file *netfs.RemoteFile
						if file, err = host.GetFileInfo(path[1]); err == nil {
							buffer := strings.Builder{}
							buffer.WriteString(file.Host.Name)
							buffer.WriteString(" ")
							buffer.WriteString(file.Path)
							buffer.WriteString(" ")
							buffer.WriteString(strconv.Itoa(int(file.Size)))
							buffer.WriteString(" ")
							buffer.WriteString(strconv.Itoa(int(file.Type)))

							return buffer.String(), nil
						} else {
							return "", err
						}
					}
				}
				err = NoAvailableHosts // TODO. Change error
			}
		}
	}
	return "", err
}

// The command shows all available hosts.
type _HostsConsoleCommand struct {
	_Client *ConsoleClient
}

// Returns the command name.
func (cmd _HostsConsoleCommand) GetName() string {
	return "hosts"
}

// Returns information about command.
func (cmd _HostsConsoleCommand) GetDescription() string {
	return strings.Join(
		[]string{
			"hosts       - The command shows all available hosts.",
			"Examples:",
			"netfs hosts - The command shows all available hosts.",
		},
		fmt.Sprintln(),
	)
}

// Executes a command with arguments.
func (cmd _HostsConsoleCommand) Execute(args ...string) (string, error) {
	hosts, err := cmd._Client._Network.GetHosts()
	if err == nil {
		if len(hosts) > 0 {
			buffer := strings.Builder{}
			for _, host := range hosts {
				buffer.WriteString(host.Name)
				buffer.WriteString(" ")
				buffer.WriteString(fmt.Sprintln(host.IP.String()))
			}
			return buffer.String(), nil
		}
		return "", NoAvailableHosts
	}
	return "", err
}

// The command shows instructions for the console client.
type _HelpConsoleCommand struct {
	_Client *ConsoleClient
}

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
			for _, command := range cmd._Client._Commands {
				if args[0] == command.GetName() {
					return command.GetDescription(), nil
				}
			}
		}
	}

	buffer := strings.Builder{}
	for _, command := range cmd._Client._Commands {
		buffer.WriteString(command.GetDescription())
	}
	return buffer.String(), nil
}
