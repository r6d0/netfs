package console

import (
	"errors"
	"fmt"
	"net"
	netfs "netfs/internal"
	"path/filepath"
	"strconv"
	"strings"
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
			fileInfoConsoleCommand{client: client},
			copyFileConsoleCommand{client: client},
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

// The command copies the file to the remote host.
type copyFileConsoleCommand struct {
	client *consoleClient
}

// Returns the command name.
func (cmd copyFileConsoleCommand) GetName() string {
	return "copy"
}

// Returns information about command.
func (cmd copyFileConsoleCommand) GetDescription() ConsoleCommandResult {
	return ConsoleCommandResult{
		[]ConsoleCommandResultLine{
			{
				Fields: []string{
					"copy [host]/[path] [host]/[path]",
					"",
					"copies the file to the remote host",
				},
			},
			{Fields: []string{"", "netfs copy 192.51.12.1/file.txt 192.51.12.65/file.txt", "copies the file to the remote host by IP"}},
			{Fields: []string{"", "netfs copy myhostname1/file.txt myhostname2/file.txt", "copies the file to the remote host by name"}},
		},
	}
}

// Executes a command with arguments.
func (cmd copyFileConsoleCommand) Execute(args ...string) (ConsoleCommandResult, error) {
	var err error
	result := unsupportedFormat()

	if len(args) > 0 {
		var hosts []netfs.RemoteHost
		if hosts, err = cmd.client.network.GetHosts(); err == nil {
			result = noAvailableHosts()
			sourcePath := strings.SplitN(args[0], PATH_SEPARATOR, PATH_SIZE)
			targetPath := strings.SplitN(args[1], PATH_SEPARATOR, PATH_SIZE)

			var sourceHost *netfs.RemoteHost
			var targetHost *netfs.RemoteHost
			for _, host := range hosts {
				if sourcePath[0] == host.Name || sourcePath[0] == host.IP.String() {
					sourceHost = &host
				}

				if targetPath[0] == host.Name || targetPath[0] == host.IP.String() {
					targetHost = &host
				}
			}

			if sourceHost != nil && targetHost != nil {
				var sourceFile *netfs.RemoteFile
				if sourceFile, err = sourceHost.GetFileInfo(sourcePath[1]); err == nil {
					result = ConsoleCommandResult{}
					err = sourceFile.CopyTo(
						&netfs.RemoteFile{
							Host: *targetHost,
							Name: filepath.Base(targetPath[1]),
							Path: targetPath[1],
							Type: sourceFile.Type,
							Size: sourceFile.Size,
						},
					)
				}
			} else {
				hosts := []string{}
				if sourceHost == nil {
					hosts = append(hosts, sourcePath[0])
				}

				if targetHost == nil {
					hosts = append(hosts, targetPath[0])
				}
				result = noAvailableHosts(hosts...)
			}
		}
	}
	return result, err
}

// The command shows information about file on remote host.
type fileInfoConsoleCommand struct { // TODO. command for shows files in a directory
	client *consoleClient
}

// Returns the command name.
func (cmd fileInfoConsoleCommand) GetName() string {
	return "file"
}

// Returns information about command.
func (cmd fileInfoConsoleCommand) GetDescription() ConsoleCommandResult {
	return ConsoleCommandResult{
		[]ConsoleCommandResultLine{
			{
				Fields: []string{
					"file [host]/[path]",
					"",
					"shows information about file on remote host",
				},
			},
			{Fields: []string{"", "netfs file 192.51.12.1/file.txt", "shows information about file on remote host by IP"}},
			{Fields: []string{"", "netfs file myhostname/file.txt", "shows information about file on remote host by name"}},
		},
	}
}

// Executes a command with arguments.
func (cmd fileInfoConsoleCommand) Execute(args ...string) (ConsoleCommandResult, error) {
	var err error
	result := unsupportedFormat()

	if len(args) > 0 {
		path := strings.SplitN(args[0], PATH_SEPARATOR, PATH_SIZE)
		if len(path) == PATH_SIZE {
			var hosts []netfs.RemoteHost
			if hosts, err = cmd.client.network.GetHosts(); err == nil {
				result = noAvailableHosts()

				for _, host := range hosts {
					if path[0] == host.Name || path[0] == host.IP.String() {
						result = ConsoleCommandResult{}

						var file *netfs.RemoteFile
						if file, err = host.GetFileInfo(path[1]); err == nil {
							result.Lines = append(
								result.Lines,
								ConsoleCommandResultLine{Fields: []string{
									file.Host.Name,
									file.Path,
									file.Type.String(),
									strconv.Itoa(int(file.Size)), // TODO. up to maximum avaliable unit.
								}},
							)
							break
						}
					}
				}
			}
		}
	}
	return result, err
}

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
	result := noAvailableHosts()
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

// Returns result "unsupported format".
func unsupportedFormat() ConsoleCommandResult {
	return ConsoleCommandResult{
		Lines: []ConsoleCommandResultLine{{Fields: []string{"unsupported format"}}},
	}
}

// Returns result "no available hosts".
func noAvailableHosts(args ...string) ConsoleCommandResult {
	if len(args) > 0 {
		return ConsoleCommandResult{[]ConsoleCommandResultLine{{Fields: []string{fmt.Sprintf("no available hosts: %s", args)}}}}
	}
	return ConsoleCommandResult{[]ConsoleCommandResultLine{{Fields: []string{"no available hosts"}}}}
}
