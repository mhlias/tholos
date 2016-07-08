package tf_helper



import (
	"log"
	"os/exec"
	"bufio"
	"strings"

)


func ExecCmd(cmdName string, args []string) bool {

	success := true

	log.Printf("[INFO] Executing command: %s %s", cmdName, strings.Join(args, " "))

	cmd := exec.Command(cmdName, args...)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		success = false
		log.Printf("Error creating StdoutPipe for Command: %s, Error: %s\n", cmdName, err.Error())
	}

	cmdErrorReader, err := cmd.StderrPipe()
	if err != nil {
		success = false
		log.Printf("Error creating StderrPipe for Command: %s, Error: %s\n", cmdName, err.Error())
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}()

	errorScanner := bufio.NewScanner(cmdErrorReader)
	go func() {
		for errorScanner.Scan() {
			log.Println(errorScanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		success = false
		log.Printf("Error starting Command: %s, Error: %s\n", cmdName, err.Error())
	}

	err = cmd.Wait()
	if err != nil {
	  success = false
		log.Printf("Error waiting for Command: %s, Error: %s\n", cmdName, err.Error())
	}

	return success


}