package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
	"strings"
)

const (
	defaultConfigPath = "/runner-config.json"
)

func main() {
	fmt.Println("starting up the test runner")
	go func() {
		fmt.Println("starting up the cxo node")
		cmd := exec.Command("cxo-node") //TODO consider sending log to output
		if err := cmd.Run(); err != nil {
			fmt.Println("cxo node failed")
			log.Fatal(err)
		}
	}()
	conf := readConfig(defaultConfigPath)
	waitOnNodeToBeUp()

	runCommands(conf)
	keepRunningOnEmpty()
}

func readConfig(path string) config {
	var confFile config

	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return confFile
	}
	fmt.Println("Successfully opened json file")
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &confFile)

	return confFile
}

func waitOnNodeToBeUp() {
	//TODO ideally this should be checked in better way
	time.Sleep(10 * time.Second)
}

func runCommands(conf config) {
	for _, command := range conf.Commands {
		if len(command.Actions) > 0 {
			if command.Actions[0] == "publish" {
				newContent := readContent(command.Actions[1])
				newPath := cutVersionPart(command.Actions[1])
				command.Actions[1] = newPath
				updateFileContent(newPath, newContent)
			}
			fmt.Println("running commands: cxo-node-cli ", command.Actions)
			cmd := exec.Command("cxo-node-cli", command.Actions...)
			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}
		}

		fmt.Printf("sleeping for %v seconds", command.SleepInSeconds)
		time.Sleep(time.Duration(command.SleepInSeconds) * time.Second)
	}
}

func keepRunningOnEmpty() {
	for {
		time.Sleep(time.Second)
	}
}

func updateFileContent(path, content string){
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)

	if err != nil {
		log.Fatalf("Failed opening file: %s", err)
	}
	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		log.Fatalf("Failed truncete of file: %s", err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		log.Fatalf("Failed writing to file: %s", err)
	}
}

func cutVersionPart(inpath string) string{
	if strings.Contains(inpath, "-version") {
		splitedCommands := strings.Split(inpath, "-version")
		return splitedCommands[0] + ".txt"
	}

	return inpath
}

func readContent(path string) string{
	file, err := os.Open(path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
        log.Fatal(err)
		return ""
	}

	return string(b)
}

type config struct {
	Commands []command `json:"commands"`
}

type command struct {
	Actions        []string `json:"actions"`
	SleepInSeconds int      `json:"sleepInSeconds"`
}
