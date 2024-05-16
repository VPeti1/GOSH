package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var fileName string = tempPath + "output.txt"
var tempfile string = tempPath + "temp.go"
var tempPath string = "/tmp/"
var fastmode bool = false
var doNewFileCreate bool = true
var magicExitCode bool = false

func main() {
	startup()
	if doNewFileCreate {
		file := createFile()
		defer file.Close()
		writeInitialMainFunction(file)
	}

	var file, _ = os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := scanner.Text()

		if input == "LIST" {
			if !fastmode {
				printFileContents(fileName)
			} else {
				fmt.Println("You cant use this command in fastmode")
				break
			}
		}

		if input == "RETURN" {
			if !fastmode {
				err := removeLastLineFromFile(fileName)
				if err != nil {
					fmt.Println("Error:", err)
					return
				} else {
					fmt.Println("Last line removed")
					fmt.Println("<")
				}
			} else {
				fmt.Println("You cant use this command in fastmode")
				break
			}
		}

		if input == "exit" {
			os.Exit(0)
		}

		if input == "RUN" {
			if !fastmode {
				RUNcmd()
				fmt.Println("<")
			} else {
				fmt.Println("You cant use this command in fastmode")
				break
			}
		}
		if input == "RESET" {
			if !fastmode {
				err := file.Truncate(0)
				if err != nil {
					fmt.Println("Error resetting file:", err)
					return
				}
				_, err = file.Seek(0, 0) // move the file pointer to the beginning (I found pointers in go)
				if err != nil {
					fmt.Println("Error resetting file:", err)
					return
				}

				continue
			} else {
				fmt.Println("You cant use this command in fastmode")
				break
			}
		}

		_, err := file.WriteString(input + "\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		prefixes := []string{"RETURN", "RUN", "LIST", "EXIT", "RESET"}
		err = removeLinesStartingWith(fileName, prefixes)
		if err != nil {
			fmt.Println("Error:", err)
		}

		if fastmode {
			RUNcmd()
			os.Remove(fileName)
			os.Remove(tempfile)
		}
	}
	magicExitCode = true
	systemPause()
	clear()
	fmt.Println("<")
	main()
}

func createFile() *os.File {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
	return file
}

func writeInitialMainFunction(file *os.File) {
	_, err := file.WriteString("\nfunc main() {\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func RUNcmd() {
	if isFileEmpty(fileName) {
		fmt.Println("You hadnt written anything yet!")
		systemPause()
		clear()
		main()
	}
	existingCode, err := readFile(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	_, err = file.WriteString("\n}\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	cmd := exec.Command("goimports", "output.txt")

	var outBuf, errBuf bytes.Buffer

	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running goimports:", err)
		fmt.Println("goimports stderr:", errBuf.String())
		return
	}
	err = ioutil.WriteFile(tempfile, []byte(outBuf.String()), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	runCommand("go run " + tempfile)
	os.Remove(tempfile)
	os.Remove(fileName)
	file, err = os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	_, err = file.WriteString(existingCode)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

}

func runCommand(input string) {
	cmd := exec.Command("sh", "-c", input)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func readFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func startup() {
	if !magicExitCode {
		defVarsForWin()
		if !goexists() {
			fmt.Println("Go not found!")
			os.Exit(1)
		}
		fmt.Println("Welcome to GOSH!")
		if !isFileEmpty(fileName) {
			if lastModifiedTime(fileName) != "Unknown time" {
				fmt.Println("Found previous GOSH code that was modified on: " + lastModifiedTime(fileName))
				ans := ask("Do you want to start over or not?")
				if strings.EqualFold(ans, "y") || strings.EqualFold(ans, "yes") {
					err := os.Remove(fileName)
					if err != nil {
						fmt.Println("Error deleting file:", err)
						return
					}
					doNewFileCreate = true
				} else {
					doNewFileCreate = false
				}
				clear()
			}
		}
		ans := ask("Do you want to use fastmode? (Recommended)")
		if ans == "y" || ans == "yes" {
			fastmode = true
			clear()
			fmt.Println("<")
			return
		} else {
			clear()
			fmt.Println("<")
		}
	}
}
func ask(question string) (answer string) {
	fmt.Println(question)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	ans := scanner.Text()
	return ans
}

func isFileEmpty(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return false
	}

	return fileInfo.Size() == 0
}

func lastModifiedTime(filePath string) string {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "Unknown time"
	}

	// Format the last modified time as a string
	modTime := fileInfo.ModTime().Format("2006-01-02 15:04:05")

	return modTime
}

func systemPause() {
	fmt.Println("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func printFileContents(fileName string) {
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(contents)))
	for scanner.Scan() {
		line := scanner.Text()
		// Skip  "func main() {" and empty lines
		if !strings.Contains(line, "func main() {") && len(strings.TrimSpace(line)) > 0 {
			fmt.Println(line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning file:", err)
	}
	fmt.Println("<")
}

func removeLastLineFromFile(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if err := file.Truncate(0); err != nil {
		return err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	for i, line := range lines {
		if i < len(lines)-1 {
			if _, err := fmt.Fprintln(file, line); err != nil {
				return err
			}
		}
	}

	return nil
}

func removeLinesStartingWith(fileName string, prefixes []string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	// Read all lines and excluding lines starting with prefixes
	for scanner.Scan() {
		line := scanner.Text()
		shouldRemove := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(strings.TrimSpace(line), prefix) {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if err := file.Truncate(0); err != nil {
		return err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(file, line); err != nil {
			return err
		}
	}

	return nil
}

func detectOS() string {
	if runtime.GOOS == "linux" {
		return "linux"
	}
	return "windows"
}

func defVarsForWin() {
	if detectOS() == "windows" {
		tempPath = os.Getenv("TEMP") + "\\"
		fileName = tempPath + "output.txt"
		tempfile = tempPath + "temp.go"
	}
}

func goexists() bool {
	cmd := exec.Command("go", "version")
	err := cmd.Run()
	return err == nil
}
