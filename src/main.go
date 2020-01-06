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
)

var rootPath string = "<RootPath>"
var fileList = list.New()
var line int = 1

func main() {
	fmt.Println("Local Addr:", getLocalAddr())

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

func getLocalAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	addr := addrs[4].String()
	addr = addr[:len(addr)-3]

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
