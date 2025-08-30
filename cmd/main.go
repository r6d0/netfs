package main

import (
	"fmt"
	_ "net/http/pprof"
	netfs "netfs/internal"
	"netfs/internal/console"
	"os"
)

const NAME_INDEX = 1

func main() {
	var name string
	var err error

	if len(os.Args) > NAME_INDEX {
		name = os.Args[NAME_INDEX]

		// Prepare configuration.
		var config netfs.Config
		if config, err = netfs.NewConfig(); err == nil {
			// Prepare console client.
			var client *console.ConsoleClient
			if client, err = console.NewConsoleClient(config); err == nil {
				// Execute command.
				var command console.ConsoleCommand
				if command, err = client.GetCommand(name); err == nil {
					var res string
					if res, err = command.Execute(os.Args[NAME_INDEX:]...); len(res) > 0 {
						fmt.Print(res)
					}
				}
			}
		}
	} else {
		err = console.NeedHelpError
	}

	if err == console.NeedHelpError || err == console.CommandNotFoundError {
		if help, _ := console.HelpCommand.Execute(name); len(help) > 0 {
			fmt.Print(help)
		}
	} else if err != nil {
		panic(err.Error())
	}
}
