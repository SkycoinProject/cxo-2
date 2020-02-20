package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultConfigPath = "/runner-config.json"
	filesPath         = "files"
)

func main() {
	fmt.Println("starting up the test runner")
	conf := readConfig(defaultConfigPath)
	waitOnNodeToBeUp()

	runCommands(conf)

	if conf.FilesToAssert != nil && len(conf.FilesToAssert) > 0 {
		checkFiles(conf.FilesToAssert)
	}
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

			if command.Actions[0] == "subscribe" {
				fmt.Println("running commands: cxo-node-cli ", command.Actions)
				cmd := exec.Command("cxo-node-cli", command.Actions...)
				if err := cmd.Run(); err != nil {
					log.Fatal(err)
				}
			} else {
				fmt.Println("running commands: cxo-file-sharing-cli ", command.Actions)
				cmd := exec.Command("cxo-file-sharing-cli", command.Actions...)
				// var outb, errb bytes.Buffer
				// cmd.Stdout = &outb
				// cmd.Stderr = &errb
				if err := cmd.Run(); err != nil {
					// fmt.Println("runner out:", outb.String())
					// fmt.Println("runner err:", errb.String())
					log.Fatal(err)
				}
			}
			// fmt.Println("runner out:", outb.String())
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

func checkFiles(files []string) {
	fmt.Println("Checking test output files and directories...")
	for _, file := range files {
		splited := strings.Split(file, ".")
		if len(splited) == 1 {
			if checkDirExistance(splited[0]) == false {
				fmt.Printf("WARNING : Directory with name %s does not exists\n", splited[0])
			}
		} else if len(splited) == 2 {
			if checkFileExistance(splited[0], splited[1], "", false) == false {
				fmt.Printf("WARNING : File with name %s does not exists in directory %s\n", splited[1], splited[0])
			}
		} else if len(splited) == 3 {
			if checkFileExistance(splited[0], splited[1], splited[2], true) == false {
				fmt.Printf("WARNING : File with name %s in directory %s does not have correct content\n", splited[1], splited[0])
			}
		} else {
			fmt.Printf("WARNING : Format of input configuration is not correct\n")
		}
	}
}

func checkFileExistance(directoryName string, fileName string, content string, checkContent bool) bool {
	// check if the source dir exist
	inputPath := filepath.Join(filesPath, directoryName, string(fileName+".txt"))
	src, err := os.Stat(inputPath)
	if err != nil {
		return false
	}

	// check if the source is directory
	if src.IsDir() {
		return false
	}

	if checkContent {
		fileContent := readContent(inputPath)
		if runtime.GOOS == "windows" {
			fileContent = strings.TrimRight(fileContent, "\r\n")
		} else {
			fileContent = strings.TrimRight(fileContent, "\n")
		}
		if strings.Compare(content, fileContent) != 0 {
			return false
		}
	}

	return true
}

func checkDirExistance(name string) bool {
	// check if the source dir exist
	src, err := os.Stat(filepath.Join(filesPath, name))
	if err != nil {
		return false
	}

	// check if the source is indeed a directory or not
	if !src.IsDir() {
		return false
	}

	return true
}

type config struct {
	Commands      []command `json:"commands"`
	FilesToAssert []string  `json:"assert"`
}

type command struct {
	Actions        []string `json:"actions"`
	SleepInSeconds int      `json:"sleepInSeconds"`
}
