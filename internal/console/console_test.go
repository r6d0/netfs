package console_test

import (
	"fmt"
	netfs "netfs/internal"
	"netfs/internal/console"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestGetCommandNotFound(t *testing.T) {
	config, _ := netfs.NewConfig()
	client, _ := console.NewConsoleClient(config)
	command, err := client.GetCommand("not found command name")
	if err != console.CommandNotFoundError {
		t.Fatalf("error should be [console.CommandNotFoundError], but error is [%s]", err)
	}

	if command != nil {
		t.Fatalf("command should be nil, but command is [%s]", reflect.TypeOf(command))
	}
}

func TestGetCommand(t *testing.T) {
	config, _ := netfs.NewConfig()
	client, _ := console.NewConsoleClient(config)
	command, err := client.GetCommand("help")
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	if command == nil {
		t.Fatalf("command should be [help], but command is nil")
	} else if command.GetName() != "help" {
		t.Fatalf("command should be [help], but command is [%s]", reflect.TypeOf(command))
	}
}

func TestHelpCommandWithoutArguments(t *testing.T) {
	config, _ := netfs.NewConfig()
	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("help")

	result, err := command.Execute()
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	if result == "" {
		t.Fatal("result should be not empty, but result is empty")
	}
	fmt.Println(result)
}

func TestHelpCommandSelfInfo(t *testing.T) {
	config, _ := netfs.NewConfig()
	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("help")

	result, err := command.Execute("help")
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	if result == "" {
		t.Fatal("result should be not empty, but result is empty")
	}
	fmt.Println(result)
}

func TestHelpCommandWithArgument(t *testing.T) {
	config, _ := netfs.NewConfig()
	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("help")

	result, err := command.Execute("hosts")
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	if result == "" {
		t.Fatal("result should be not empty, but result is empty")
	}
	fmt.Println(result)
}

func TestHostsCommandWithoutAvailableHosts(t *testing.T) {
	config, _ := netfs.NewConfig()
	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("hosts")

	_, err := command.Execute()
	if err != console.NoAvailableHosts {
		t.Fatalf("error should be [console.NoAvailableHosts], but error is [%s]", err)
	}
}

func TestHostsCommandWithAvailableHosts(t *testing.T) {
	config, _ := netfs.NewConfig()
	server, _ := netfs.NewServer(config)
	go func() {
		server.Start()
	}()
	time.Sleep(1 * time.Second) // Waiting for the server to start

	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("hosts")

	result, err := command.Execute()
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}
	fmt.Println(result)

	defer server.Stop()
	defer os.RemoveAll(config.Database.Path)
}
