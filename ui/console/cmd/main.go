package main

// import (
// 	"fmt"
// 	"netfs/api"
// 	"netfs/api/transport"
// 	"time"
// )

// func main() {
// 	network, err := api.NewNetwork(api.NetworkConfig{Port: 8989, Protocol: transport.HTTP, Timeout: time.Second * 5})
// 	if err != nil {
// 		panic(err)
// 	}

// 	hosts, err := network.Hosts()
// 	fmt.Println(hosts)
// }

import (
	"netfs/api"
	"netfs/api/transport"
	"netfs/ui/console"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	network, err := api.NewNetwork(api.NetworkConfig{Port: 8989, Protocol: transport.HTTP, Timeout: time.Second * 1})
	if err == nil {
		program := tea.NewProgram(console.NewConsoleViewModel(network), tea.WithAltScreen())

		go func(program *tea.Program) {
			time.Sleep(1 * time.Second) // TODO. from settings

		}(program)
		_, err = program.Run()
	}

	if err != nil {
		panic(err)
	}
}
