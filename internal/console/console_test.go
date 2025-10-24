package console_test

// import (
// 	"fmt"
// 	netfs "netfs/internal"
// 	"netfs/internal/console"
// 	"netfs/internal/server"
// 	"os"
// 	"reflect"
// 	"testing"
// 	"time"
// )

// const TEST_FILE_PATH = "./file.txt"
// const TARGET_TEST_FILE_PATH = "./file_copy.txt"

// func TestGetCommandNotFound(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	client, _ := console.NewConsoleClient(config)
// 	command, err := client.GetCommand("not found command name")
// 	if err == nil {
// 		t.Fatalf("error should be not nil, but error is nil")
// 	}

// 	if command != nil {
// 		t.Fatalf("command should be nil, but command is [%s]", reflect.TypeOf(command))
// 	}
// }

// func TestGetCommand(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	client, _ := console.NewConsoleClient(config)
// 	command, err := client.GetCommand("help")
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}

// 	if command == nil {
// 		t.Fatalf("command should be [help], but command is nil")
// 	} else if command.GetName() != "help" {
// 		t.Fatalf("command should be [help], but command is [%s]", reflect.TypeOf(command))
// 	}
// }

// // Help command

// func TestHelpCommandWithoutArguments(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("help")

// 	result, err := command.Execute()
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}

// 	if len(result.Lines) == 0 {
// 		t.Fatal("result should be not empty, but result is empty")
// 	}
// 	fmt.Println(result)
// }

// func TestHelpCommandSelfInfo(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("help")

// 	result, err := command.Execute("help")
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}

// 	if len(result.Lines) == 0 {
// 		t.Fatal("result should be not empty, but result is empty")
// 	}
// 	fmt.Println(result)
// }

// func TestHelpCommandWithArgument(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("help")

// 	result, err := command.Execute("hosts")
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}

// 	if len(result.Lines) == 0 {
// 		t.Fatal("result should be not empty, but result is empty")
// 	}
// 	fmt.Println(result)
// }

// // Hosts command

// func TestHostsCommandWithoutAvailableHosts(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("hosts")

// 	result, err := command.Execute()
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}

// 	if result.Lines[0].Fields[0] != "no available hosts" {
// 		t.Fatalf("message should be [no available hosts], but message is [%s]", result.Lines[0].Fields[0])
// 	}
// }

// func TestHostsCommandWithAvailableHosts(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("hosts")

// 	result, err := command.Execute()
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}

// 	if len(result.Lines) == 0 {
// 		t.Fatalf("result should be not empty, but result is empty")
// 	}
// 	fmt.Println(result)

// 	srv.Stop()
// 	os.RemoveAll(config.Database.Path)
// }

// // File command

// func TestFileInfoCommandWithoutArguments(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("file")

// 	result, err := command.Execute()
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}

// 	if result.Lines[0].Fields[0] != "unsupported format" {
// 		t.Fatalf("result should be equals [unsupported format]")
// 	}
// 	fmt.Println(result)
// }

// func TestFileInfoCommandWithIncorrectArguments(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("file")

// 	result, err := command.Execute("test")
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}

// 	if result.Lines[0].Fields[0] != "unsupported format" {
// 		t.Fatalf("result should be equals [unsupported format]")
// 	}
// 	fmt.Println(result)
// }

// func TestFileInfoCommandWithoutAvailableHosts(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("file")

// 	result, err := command.Execute("192.1.1.2/./main.exe")
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}

// 	if result.Lines[0].Fields[0] != "no available hosts" {
// 		t.Fatalf("result should be equals [no available hosts]")
// 	}
// 	fmt.Println(result)
// }

// func TestFileInfoCommandWithIncorrectHost(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("file")

// 	result, err := command.Execute("TEST/./file.txt")
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}

// 	if result.Lines[0].Fields[0] != "no available hosts" {
// 		t.Fatalf("result should be equals [no available hosts], but result is [%s]", result.Lines[0].Fields[0])
// 	}

// 	srv.Stop()
// 	os.RemoveAll(config.Database.Path)
// }

// func TestFileInfoCommandByIp(t *testing.T) {
// 	os.WriteFile(TEST_FILE_PATH, []byte("TEST"), 0644)

// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	netwotk, _ := netfs.NewNetwork(config)
// 	hosts, _ := netwotk.GetHosts()

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("file")

// 	result, err := command.Execute(hosts[0].IP.String() + "/" + TEST_FILE_PATH)
// 	if err != nil {
// 		t.Fatalf("error should be nil but error is [%s]", err)
// 	}
// 	fmt.Println(result)

// 	srv.Stop()
// 	os.RemoveAll(config.Database.Path)
// 	os.RemoveAll(TEST_FILE_PATH)
// }

// func TestFileInfoCommandByHostName(t *testing.T) {
// 	os.RemoveAll(TEST_FILE_PATH)
// 	os.WriteFile(TEST_FILE_PATH, []byte("TEST"), 0644)

// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	netwotk, _ := netfs.NewNetwork(config)
// 	hosts, _ := netwotk.GetHosts()

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("file")

// 	result, err := command.Execute(hosts[0].Name + "/" + TEST_FILE_PATH)
// 	if err != nil {
// 		t.Fatalf("error should be nil but error is [%s]", err)
// 	}
// 	fmt.Println(result)

// 	srv.Stop()
// 	os.RemoveAll(config.Database.Path)
// 	os.RemoveAll(TEST_FILE_PATH)
// }

// func TestFileInfoCommandNotFoundFile(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	netwotk, _ := netfs.NewNetwork(config)
// 	hosts, _ := netwotk.GetHosts()

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("file")

// 	result, err := command.Execute(hosts[0].Name + "/" + TEST_FILE_PATH)
// 	if err == nil {
// 		t.Fatalf("error should be not nil but error is nil")
// 	}
// 	fmt.Println(result)

// 	srv.Stop()
// 	os.RemoveAll(config.Database.Path)
// }

// // Copy command

// func TestCopyFileConsoleCommandByHostName(t *testing.T) {
// 	os.RemoveAll(TARGET_TEST_FILE_PATH)
// 	os.RemoveAll(TEST_FILE_PATH)
// 	os.WriteFile(TEST_FILE_PATH, []byte("TEST"), 0644)

// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	netwotk, _ := netfs.NewNetwork(config)
// 	hosts, _ := netwotk.GetHosts()

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("copy")

// 	result, err := command.Execute(hosts[0].Name+"/"+TEST_FILE_PATH, hosts[0].Name+"/"+TARGET_TEST_FILE_PATH)
// 	if err != nil {
// 		t.Fatalf("error should be nil but error is [%s]", err)
// 	}
// 	fmt.Println(result)

// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	_, err = os.Open(TARGET_TEST_FILE_PATH)
// 	if err != nil {
// 		t.Fatalf("error should be nil but error is [%s]", err)
// 	}

// 	srv.Stop()
// 	os.RemoveAll(config.Database.Path)
// 	os.RemoveAll(TARGET_TEST_FILE_PATH)
// 	os.RemoveAll(TEST_FILE_PATH)
// }

// func TestCopyFileConsoleCommandByIp(t *testing.T) {
// 	os.RemoveAll(TARGET_TEST_FILE_PATH)
// 	os.RemoveAll(TEST_FILE_PATH)
// 	os.WriteFile(TEST_FILE_PATH, []byte("TEST"), 0644)

// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	netwotk, _ := netfs.NewNetwork(config)
// 	hosts, _ := netwotk.GetHosts()

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("copy")

// 	result, err := command.Execute(hosts[0].IP.String()+"/"+TEST_FILE_PATH, hosts[0].IP.String()+"/"+TARGET_TEST_FILE_PATH)
// 	if err != nil {
// 		t.Fatalf("error should be nil but error is [%s]", err)
// 	}
// 	fmt.Println(result)

// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	_, err = os.Open(TARGET_TEST_FILE_PATH)
// 	if err != nil {
// 		t.Fatalf("error should be nil but error is [%s]", err)
// 	}

// 	srv.Stop()
// 	os.RemoveAll(config.Database.Path)
// 	os.RemoveAll(TARGET_TEST_FILE_PATH)
// 	os.RemoveAll(TEST_FILE_PATH)
// }

// func TestCopyFileConsoleCommandWithNotFoundFile(t *testing.T) {
// 	os.RemoveAll(TARGET_TEST_FILE_PATH)

// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	netwotk, _ := netfs.NewNetwork(config)
// 	hosts, _ := netwotk.GetHosts()

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("copy")

// 	result, err := command.Execute(hosts[0].Name+"/"+"./incorrect_file.txt", hosts[0].Name+"/"+TARGET_TEST_FILE_PATH)
// 	if err == nil {
// 		t.Fatalf("error should be not nil but error is nil")
// 	}
// 	fmt.Println(result)

// 	srv.Stop()
// 	os.RemoveAll(config.Database.Path)
// 	os.RemoveAll(TARGET_TEST_FILE_PATH)
// }

// func TestCopyFileConsoleCommandWithIncorrectHost(t *testing.T) {
// 	os.RemoveAll(TARGET_TEST_FILE_PATH)

// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("copy")

// 	result, err := command.Execute("123.1.1.1/"+TEST_FILE_PATH, "123.1.1.1/"+TARGET_TEST_FILE_PATH)
// 	if err != nil {
// 		t.Fatalf("error should be nil but error is [%s]", err)
// 	}

// 	if result.Lines[0].Fields[0] != "no available hosts: [123.1.1.1 123.1.1.1]" {
// 		t.Fatalf("result should be equals [no available hosts: [123.1.1.1 123.1.1.1]], but result is: [%s]", result.Lines[0].Fields[0])
// 	}

// 	srv.Stop()
// 	os.RemoveAll(config.Database.Path)
// 	os.RemoveAll(TARGET_TEST_FILE_PATH)
// }

// // Start command

// func TestStartCommand(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	go func() {
// 		client, _ := console.NewConsoleClient(config)
// 		command, _ := client.GetCommand("start")
// 		if _, err := command.Execute(); err != nil {
// 			panic(err)
// 		}
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("hosts")

// 	_, err := command.Execute()
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}
// }

// func TestStartCommandWhenServerStartedAlready(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("start")

// 	_, err := command.Execute()
// 	if err == nil {
// 		t.Fatal("error should be not nil, but error is nil")
// 	}
// }

// // Stop command

// func TestStopCommand(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	srv, _ := server.NewServer(config)
// 	go func() {
// 		srv.Start()
// 	}()
// 	time.Sleep(1 * time.Second) // Waiting for the server to start

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("stop")

// 	_, err := command.Execute()
// 	if err != nil {
// 		t.Fatalf("error should be nil, but error is [%s]", err)
// 	}
// }

// func TestStopCommandWhenServerNotStarted(t *testing.T) {
// 	config, _ := netfs.NewConfig()
// 	os.RemoveAll(config.Database.Path)

// 	client, _ := console.NewConsoleClient(config)
// 	command, _ := client.GetCommand("stop")

// 	_, err := command.Execute()
// 	if err == nil {
// 		t.Fatal("error should be not nil, but error is nil")
// 	}
// }
