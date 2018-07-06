package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"os"
	"bufio"
	"io/ioutil"
	"strings"
)

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true if slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		Crash(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

func launchClient() {
	var com *exec.Cmd = nil
	if runtime.GOOS == "windows" {
		com = exec.Command("Launch.bat")
	} else {
		os.Chmod("Launch.sh", 0770)
		com = exec.Command("./Launch.sh")
	}
	err := com.Run()
	if err != nil {
		fmt.Println()
		fmt.Println("==================================")
		fmt.Println("Client was installed successfully!")
		fmt.Println("==================================")
		fmt.Println("However, we were unable to start TLauncher.")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("!MAKE SURE JAVA IS INSTALLED AND RUN UPDATER AGAIN!")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("Press any key to exit.")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func checkDir() bool {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		Crash("Unable to read current directory:", err.Error())
	}

	if len(files) > 1 {
		checkFirst := false
		checkSecond := false
		checkThird := false
		for _, v := range files {
			if strings.Contains(v.Name(), "versions") {
				checkFirst = true
			} else if strings.Contains(v.Name(), "mods") {
				checkSecond = true
			} else if strings.Contains(v.Name(), "config") {
				checkThird = true
			}
		}

		if !(checkSecond && checkFirst && checkThird) {
			return true
		}
	}
	return false
}
