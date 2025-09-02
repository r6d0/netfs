package console_test

import (
	netfs "netfs/internal"
	"netfs/internal/console"
	"reflect"
	"testing"
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
	command, err := client.GetCommand(console.HelpCommand.GetName())
	if err != nil {
		t.Fatalf("error should be nil, but error is [%s]", err)
	}

	if command == nil {
		t.Fatalf("command should be [console.HelpCommand], but command is nil")
	} else if command != console.HelpCommand {
		t.Fatalf("command should be [console.HelpCommand], but command is [%s]", reflect.TypeOf(command))
	}
}
