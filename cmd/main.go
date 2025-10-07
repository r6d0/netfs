package main

import (
	"fmt"
	_ "net/http/pprof"
	netfs "netfs/internal"
	"netfs/internal/console"
	"os"
	"strconv"
)

const NAME_INDEX = 1

func main() {
	var err error
	var config *netfs.Config

	if config, err = netfs.NewConfig(); err == nil {
		var client console.ConsoleClient
		if client, err = console.NewConsoleClient(config); err == nil {
			name := "help"
			if len(os.Args) > NAME_INDEX {
				name = os.Args[NAME_INDEX]
			}

			var command console.ConsoleCommand
			if command, err = client.GetCommand(name); err == nil {
				var res console.ConsoleCommandResult
				if res, err = command.Execute(os.Args[NAME_INDEX+1:]...); err == nil {
					print(res)
				}
			}
		}
	}

	if err != nil {
		panic(err)
	}
}

func print(result console.ConsoleCommandResult) {
	maxFieldSize := []int{0}
	for _, line := range result.Lines {
		for index, field := range line.Fields {
			if index >= len(maxFieldSize) {
				maxFieldSize = append(maxFieldSize, 0)
			}
			maxFieldSize[index] = max(maxFieldSize[index], len(field))
		}
	}

	for _, line := range result.Lines {
		for index, size := range maxFieldSize {
			value := ""
			if index < len(line.Fields) {
				value = line.Fields[index]
			}
			fmt.Printf("%-"+strconv.Itoa(size)+"s", value)
			fmt.Print(" ")
		}
		fmt.Println()
	}
}
