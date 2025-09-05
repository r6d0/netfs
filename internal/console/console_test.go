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

const TEST_FILE_PATH = "./file.txt"

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
	os.RemoveAll(config.Database.Path)

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

func TestFileInfoCommandWithoutArguments(t *testing.T) {
	config, _ := netfs.NewConfig()
	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("file")

	_, err := command.Execute()
	if err != console.NeedHelpError {
		t.Fatalf("error should be [console.NeedHelpError], but error is [%s]", err)
	}
}

func TestFileInfoCommandWithIncorrectArguments(t *testing.T) {
	config, _ := netfs.NewConfig()
	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("file")

	_, err := command.Execute("test")
	if err != console.NeedHelpError {
		t.Fatalf("error should be [console.NeedHelpError], but error is [%s]", err)
	}
}

func TestFileInfoCommandWithoutAvailableHosts(t *testing.T) {
	config, _ := netfs.NewConfig()
	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("file")

	_, err := command.Execute("192.1.1.2/./main.exe")
	if err != console.NoAvailableHosts {
		t.Fatalf("error should be [console.NoAvailableHosts], but error is [%s]", err)
	}
}

func TestFileInfoCommandWithIncorrectHost(t *testing.T) {
	config, _ := netfs.NewConfig()
	os.RemoveAll(config.Database.Path)

	server, _ := netfs.NewServer(config)
	go func() {
		server.Start()
	}()
	time.Sleep(1 * time.Second) // Waiting for the server to start

	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("file")

	_, err := command.Execute("TEST/./main.exe")
	if err != console.NoAvailableHosts {
		t.Fatalf("error should be [console.NoAvailableHosts], but error is [%s]", err)
	}

	defer server.Stop()
	defer os.RemoveAll(config.Database.Path)
}

func TestFileInfoCommandByIp(t *testing.T) {
	os.WriteFile(TEST_FILE_PATH, []byte("TEST"), 0644)

	config, _ := netfs.NewConfig()
	os.RemoveAll(config.Database.Path)

	server, _ := netfs.NewServer(config)
	go func() {
		server.Start()
	}()
	time.Sleep(1 * time.Second) // Waiting for the server to start

	netwotk, _ := netfs.NewNetwork(config)
	hosts, _ := netwotk.GetHosts()

	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("file")

	result, err := command.Execute(hosts[0].IP.String() + "/" + TEST_FILE_PATH)
	if err != nil {
		t.Fatalf("error should be nil but error is [%s]", err)
	}
	fmt.Println(result)

	defer server.Stop()
	defer os.RemoveAll(config.Database.Path)
	defer os.RemoveAll(TEST_FILE_PATH)
}

func TestFileInfoCommandByHostName(t *testing.T) {
	os.RemoveAll(TEST_FILE_PATH)
	os.WriteFile(TEST_FILE_PATH, []byte("TEST"), 0644)

	config, _ := netfs.NewConfig()
	os.RemoveAll(config.Database.Path)

	server, _ := netfs.NewServer(config)
	go func() {
		server.Start()
	}()
	time.Sleep(1 * time.Second) // Waiting for the server to start

	netwotk, _ := netfs.NewNetwork(config)
	hosts, _ := netwotk.GetHosts()

	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("file")

	result, err := command.Execute(hosts[0].Name + "/" + TEST_FILE_PATH)
	if err != nil {
		t.Fatalf("error should be nil but error is [%s]", err)
	}
	fmt.Println(result)

	defer server.Stop()
	defer os.RemoveAll(config.Database.Path)
	defer os.RemoveAll(TEST_FILE_PATH)
}

func TestFileInfoCommandNotFoundFile(t *testing.T) {
	config, _ := netfs.NewConfig()
	os.RemoveAll(config.Database.Path)

	server, _ := netfs.NewServer(config)
	go func() {
		server.Start()
	}()
	time.Sleep(1 * time.Second) // Waiting for the server to start

	netwotk, _ := netfs.NewNetwork(config)
	hosts, _ := netwotk.GetHosts()

	client, _ := console.NewConsoleClient(config)
	command, _ := client.GetCommand("file")

	result, err := command.Execute(hosts[0].Name + "/" + TEST_FILE_PATH)
	if err == nil {
		t.Fatalf("error should be not nil but error is nil")
	}
	fmt.Println(result)

	defer server.Stop()
	defer os.RemoveAll(config.Database.Path)
}
