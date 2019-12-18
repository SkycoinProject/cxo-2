package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
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

type config struct {
	Commands []command `json:"commands"`
}

type command struct {
	Actions        []string `json:"actions"`
	SleepInSeconds int      `json:"sleepInSeconds"`
}
