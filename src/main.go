package main

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

var rootPath string = ""
var ip = ""
var fileList = list.New()
var line int = 1

type Config struct {
	Root string
	IP   string
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		if args[0] == "help" {
			help()
		} else {
			startArgs(args)
			operations()
		}
	} else {
		fmt.Println("No Args")
		if !checkConfigFile() {
			createConfigFile()
		} else {
			readConfigFile()
			operations()
		}
	}
}

func help() {
	fmt.Println(`
!If no ip address is specified, the local ip address is used.
!The Root Path has to be defined.

Use Args:
	.difip <rootpath> <ip>
	.difip <rootpath>
	.difip help

Use Config File:
	The variables inside the config.toml file must be changed.
	`)
}

func startArgs(args []string) {
	if len(args) == 1 {
		rootPath = args[0]
		ip = getLocalAddr()
		fmt.Println("RootPath: ", rootPath, "IP: ", ip)
	} else if len(args) == 2 {
		rootPath = args[0]
		ip = args[1]
		fmt.Println("Root Path: ", rootPath, "IP: ", ip)
	} else {
		fmt.Println("Undefined Arguments")
		fmt.Println("run <rootpath> <ip>")
	}
}

func operations() {
	fmt.Println("Local Addr:", ip, "Root Path:", rootPath)

	fmt.Println("Dir Tree: ")
	visit(rootPath, 0)

	fmt.Println("File Paths:")
	for e := fileList.Front(); e != nil; e = e.Next() {
		fmt.Println("\t", e.Value)
	}

	fmt.Println("Match Pattern: ")
	for e := fileList.Front(); e != nil; e = e.Next() {
		strValue := fmt.Sprintf("%v", e.Value)
		changeAddr(strValue)
	}
}

func readConfigFile() {
	var conf Config
	if _, err := toml.DecodeFile("config.toml", &conf); err != nil {
		panic(err)
	}
	fmt.Println(conf.Root, conf.IP)
	rootPath = conf.Root
	if conf.IP != "IP address" {
		ip = conf.IP
	} else {
		ip = getLocalAddr()
	}
}

func createConfigFile() {
	configFile, err := os.Create("config.toml")
	if err != nil {
		panic(err)
	}
	write := `
# difip Config File

root = "Root Path"
# default local ip
ip = "IP address"
	`
	ok, err := configFile.WriteString(write)
	if err != nil {
		panic(err)
	}
	fmt.Println("Create Config File Success.", ok)
	fmt.Println("Please check config file and restart difip.")
}

func checkConfigFile() bool {
	configFile, err := os.Open("config.toml")
	if err != nil {
		return false
	}
	defer configFile.Close()

	return true
}

func getLocalAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	addr := addrs[4].String()
	addr = addr[:len(addr)-3]

	ip = addr

	return addr
}

func visit(path string, counter int) {
	directories, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, dir := range directories {
		tree := strings.Repeat("\t", counter)
		if dir.IsDir() {
			fmt.Println(tree, "Directory:", dir.Name())
			newPath := path + dir.Name() + "/"
			visit(newPath, counter+1)
		} else {
			fmt.Println(tree, "File:", dir.Name())
			fileList.PushFront(path + dir.Name())
		}
	}
}

func changeAddr(filePath string) {
	line = 1
	re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if re.MatchString(scanner.Text()) {
			fmt.Print("Path: ", filePath, "\n", "Line: ", line, "\n", "Text: ", scanner.Text(), "\n")
			fmt.Print("Addr: ", re.FindString(scanner.Text()), "\n")
			result := strings.Replace(scanner.Text(), re.FindString(scanner.Text()), getLocalAddr(), -1)
			fmt.Println("Replace Text: ", result)

			input, err := ioutil.ReadFile(filePath)
			if err != nil {
				panic(err)
			}

			output := bytes.Replace(input, []byte(scanner.Text()), []byte(result), -1)
			if err = ioutil.WriteFile(filePath, output, 0666); err != nil {
				panic(err)
			}

			fmt.Print("Done: Replace OK!\n\n")
		}

		line++
	}
}
