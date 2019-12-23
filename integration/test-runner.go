package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	defaultConfigPath = "/runner-config.json"
)

func main() {
	fmt.Println("starting up the test runner")
	conf := readConfig(defaultConfigPath)
	waitOnNodeToBeUp()

	runCommands(conf)
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
			if command.Actions[0] == "publish" && strings.Contains(command.Actions[1], "-version") {
				newPath := cutVersionPart(command.Actions[1])
				if strings.Contains(newPath, ".txt") { //File
					newContent := readContent(newPath)
					updateFileContent(newPath, newContent)
				} else { //Directory
					removeContents(newPath)
					err := copyFolder(command.Actions[1], newPath)
					if err != nil {
						log.Fatal(err)
					}
				}

				command.Actions[1] = newPath
			}
			fmt.Println("running commands: cxo-node-cli ", command.Actions)
			cmd := exec.Command("cxo-node-cli", command.Actions...)
			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}
		}

		fmt.Printf("sleeping for %v seconds\n", command.SleepInSeconds)
		time.Sleep(time.Duration(command.SleepInSeconds) * time.Second)
	}
}

func updateFileContent(path, content string) {
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

func cutVersionPart(inpath string) string {
	splitedCommands := strings.Split(inpath, "-version")
	if strings.Contains(inpath, ".txt") {
		return splitedCommands[0] + ".txt"
	}

	return splitedCommands[0]
}

func readContent(path string) string {
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

func removeContents(dir string) {
	dir = dir[1:]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModeDir)
	}

	// Open the directory and read all its files.
	dirRead, _ := os.Open(dir)
	dirFiles, _ := dirRead.Readdir(0)

	// Loop over the directory's files.
	for index := range dirFiles {
		fileHere := dirFiles[index]

		// Get name of file and its full path.
		nameHere := fileHere.Name()
		fullPath := dir + nameHere

		// Remove the file.
		os.Remove(fullPath)
	}
}

func copyFolder(source string, dest string) (err error) {
	source = source[1:]
	dest = dest[1:]
	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			err = copyFolder(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			err = copyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return
}

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}

	return
}

type config struct {
	Commands []command `json:"commands"`
}

type command struct {
	Actions        []string `json:"actions"`
	SleepInSeconds int      `json:"sleepInSeconds"`
}
