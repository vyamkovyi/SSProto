// utils.go - collection of useful hacks used in project
// This file is not copyrighted. Do whatever you want to the code below, including copying and modifying.
package main

import (
	"fmt"
	"os"
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

// askForConfirmation asks user to answer a yes/no question and interprets answer as boolean
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

// fileExists checks if file in specified path exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
