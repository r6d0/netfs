package console

import (
	"errors"
	"fmt"
	"net"
	netfs "netfs/internal"
)

// -------------------------------------------------------- PUBLIC CODE ---------------------------------------------------------

var NeedHelpError = errors.New("help")
var NoAvailableHosts = errors.New("No available hosts")

type ConsoleCommandResultLine struct {
	Fields []string
}

// The result of the command execution.
type ConsoleCommandResult struct {
	Lines []ConsoleCommandResultLine
}

// Abstraction of console command.
type ConsoleCommand interface {
	// Returns the command name.
	GetName() string
	// Returns information about command.
	GetDescription() ConsoleCommandResult
	// Executes a command with arguments.
	Execute(args ...string) (ConsoleCommandResult, error)
}

// A client for interacting with netfs via the console.
type ConsoleClient interface {
	// Returns the command by name and returns an error if the command is not found.
	GetCommand(name string) (ConsoleCommand, error)
}

// Creates a new instance of ConsoleClient, returns an error if creation failed.
func NewConsoleClient(config *netfs.Config) (ConsoleClient, error) {
	network, err := netfs.NewNetwork(config)
	if err == nil {
		client := &consoleClient{}
		client.config = config
		client.network = network
		client.commands = []ConsoleCommand{
			helpConsoleCommand{client: client},
			hostsConsoleCommand{client: client},
		}

		return client, nil
	}
	return nil, err
}

// -------------------------------------------------------- PRIVATE CODE --------------------------------------------------------

const PATH_SEPARATOR = "/"
const PATH_SIZE = 2

type consoleClient struct {
	config   *netfs.Config
	network  *netfs.Network
	commands []ConsoleCommand
}

// Returns the command by name and returns an error if the command is not found.
func (client *consoleClient) GetCommand(name string) (ConsoleCommand, error) {
	for _, command := range client.commands {
		if name == command.GetName() {
			return command, nil
		}
	}
	return nil, fmt.Errorf("Command [%s] not found. Use [help] for details", name)
}

// type _CopyFileConsoleCommand struct {
// 	_Client *ConsoleClient
// }

// // Returns the command name.
// func (cmd _CopyFileConsoleCommand) GetName() string {
// 	return "copy"
// }

// // Returns information about command.
// func (cmd _CopyFileConsoleCommand) GetDescription() string {
// 	return strings.Join(
// 		[]string{},
// 		fmt.Sprintln(),
// 	)
// }

// // Executes a command with arguments.
// func (cmd _CopyFileConsoleCommand) Execute(args ...string) (string, error) {
// 	var err error

// 	if len(args) > 0 {
// 		var hosts []netfs.RemoteHost
// 		if hosts, err = cmd._Client._Network.GetHosts(); err == nil {
// 			sourcePath := strings.SplitN(args[0], PATH_SEPARATOR, PATH_SIZE)
// 			targetPath := strings.SplitN(args[1], PATH_SEPARATOR, PATH_SIZE)

// 			var sourceHost *netfs.RemoteHost
// 			var targetHost *netfs.RemoteHost
// 			for _, host := range hosts {
// 				if sourcePath[0] == host.Name || sourcePath[0] == host.IP.String() {
// 					sourceHost = &host
// 				}

// 				if targetPath[0] == host.Name || targetPath[0] == host.IP.String() {
// 					targetHost = &host
// 				}
// 			}

// 			if sourceHost != nil && targetHost != nil {
// 				var sourceFile *netfs.RemoteFile
// 				if sourceFile, err = sourceHost.GetFileInfo(sourcePath[1]); err == nil {
// 					err = sourceFile.CopyTo(
// 						&netfs.RemoteFile{
// 							Host: *targetHost,
// 							Name: filepath.Base(targetPath[1]),
// 							Path: targetPath[1],
// 							Type: sourceFile.Type,
// 							Size: sourceFile.Size,
// 						},
// 					)
// 				}
// 			} else {
// 				err = NoAvailableHosts
// 			}
// 		}
// 	}
// 	return "", err
// }

// // The command shows information about file on remote host.
// type _FileInfoConsoleCommand struct {
// 	_Client *ConsoleClient
// }

// // Returns the command name.
// func (cmd _FileInfoConsoleCommand) GetName() string {
// 	return "file"
// }

// // Returns information about command.
// func (cmd _FileInfoConsoleCommand) GetDescription() string {
// 	return strings.Join(
// 		[]string{
// 			"file host/file.txt              - shows information about file on remote host.",
// 			"Examples:",
// 			"netfs file 192.51.12.1/file.txt - shows information about file on remote host.",
// 			"netfs file myhostname/file.txt  - shows information about file on remote host.",
// 		},
// 		fmt.Sprintln(),
// 	)
// }

// // Executes a command with arguments.
// func (cmd _FileInfoConsoleCommand) Execute(args ...string) (string, error) {
// 	err := NeedHelpError

// 	if len(args) > 0 {
// 		path := strings.SplitN(args[0], PATH_SEPARATOR, PATH_SIZE)
// 		if len(path) == PATH_SIZE {
// 			var hosts []netfs.RemoteHost
// 			if hosts, err = cmd._Client._Network.GetHosts(); err == nil {
// 				for _, host := range hosts {
// 					if path[0] == host.Name || path[0] == host.IP.String() {
// 						var file *netfs.RemoteFile
// 						if file, err = host.GetFileInfo(path[1]); err == nil {
// 							buffer := strings.Builder{}
// 							buffer.WriteString(file.Host.Name)
// 							buffer.WriteString(" ")
// 							buffer.WriteString(file.Path)
// 							buffer.WriteString(" ")
// 							buffer.WriteString(strconv.Itoa(int(file.Size)))
// 							buffer.WriteString(" ")
// 							buffer.WriteString(strconv.Itoa(int(file.Type)))

// 							return buffer.String(), nil
// 						} else {
// 							return "", err
// 						}
// 					}
// 				}
// 				err = NoAvailableHosts // TODO. Change error
// 			}
// 		}
// 	}
// 	return "", err
// }

// The command shows all available hosts.
type hostsConsoleCommand struct {
	client *consoleClient
}

// Returns the command name.
func (cmd hostsConsoleCommand) GetName() string {
	return "hosts"
}

// Returns information about command.
func (cmd hostsConsoleCommand) GetDescription() ConsoleCommandResult {
	return ConsoleCommandResult{
		[]ConsoleCommandResultLine{
			{
				Fields: []string{
					"hosts",
					"",
					"shows all available hosts",
				},
			},
			{Fields: []string{"", "netfs hosts", "shows all available hosts"}},
		},
	}
}

// Executes a command with arguments.
func (cmd hostsConsoleCommand) Execute(args ...string) (ConsoleCommandResult, error) {
	result := ConsoleCommandResult{[]ConsoleCommandResultLine{{Fields: []string{"no available hosts"}}}}
	hosts, err := cmd.client.network.GetHosts()
	if err == nil {
		if len(hosts) > 0 {
			var ip net.IP
			if ip, err = cmd.client.network.GetLocalIP(); err == nil {
				result = ConsoleCommandResult{Lines: []ConsoleCommandResultLine{}}
				for _, host := range hosts {
					current := ""
					if ip.Equal(host.IP) {
						current = "(you)"
					}

					result.Lines = append(
						result.Lines,
						ConsoleCommandResultLine{Fields: []string{current, host.Name, fmt.Sprint(host.IP.String())}},
					)
				}
			}
		}
	}
	return result, err
}

// The command shows instructions for the console client.
type helpConsoleCommand struct {
	client *consoleClient
}

// Returns the command name.
func (cmd helpConsoleCommand) GetName() string {
	return "help"
}

// Returns information about command.
func (cmd helpConsoleCommand) GetDescription() ConsoleCommandResult {
	return ConsoleCommandResult{
		[]ConsoleCommandResultLine{
			{
				Fields: []string{
					"help [command]",
					"",
					"shows all available commands for the console client or information about a specific command",
				},
			},
			{Fields: []string{"", "netfs help", "shows all available commands for the console client"}},
			{Fields: []string{"", "netfs help command", "shows information about a specific command"}},
		},
	}
}

// Executes a command with arguments.
func (cmd helpConsoleCommand) Execute(args ...string) (ConsoleCommandResult, error) {
	if len(args) > 0 {
		if args[0] == cmd.GetName() {
			return cmd.GetDescription(), nil
		} else {
			for _, command := range cmd.client.commands {
				if args[0] == command.GetName() {
					return command.GetDescription(), nil
				}
			}
		}
	}

	result := ConsoleCommandResult{}
	for _, command := range cmd.client.commands {
		result.Lines = append(result.Lines, command.GetDescription().Lines...)
	}
	return result, nil
}
