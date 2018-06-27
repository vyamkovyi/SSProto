package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"crypto/sha256"
	"os"
)

var filesMap map[[32]byte]string

func listMods() {
	files, err := ioutil.ReadDir("mods/")
	if err != nil {
		log.Fatalln("Failed to read mods directory:", err.Error())
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fullFileName := "mods/" + f.Name()
		if filepath.Ext(fullFileName) != ".jar" {
			continue
		}

		s, err := ioutil.ReadFile(fullFileName)
		if err != nil {
			log.Fatalln("Failed to read file", fullFileName, ":", err.Error())
		}
		hash := sha256.Sum256(s)

		filesMap[hash] = fullFileName
	}
}

func walkConfigs(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Fatalln("Failed to read config directory!", err.Error())
	}

	// TODO: skip unwanted directories and files
	if info.IsDir() {
		return nil
	}

	s, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln("Failed to read file", path, ":", err.Error())
	}
	hash := sha256.Sum256(s)

	filesMap[hash] = path
	return nil
}

func walkClientFiles(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Fatalln("Failed to read client files directory!", err.Error())
	}

	if info.IsDir() {
		return nil
	}

	s, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln("Failed to read file", path, ":", err.Error())
	}
	hash := sha256.Sum256(s)

	filesMap[hash] = path
	return nil
}

func ListFiles() map[[32]byte]string {
	filesMap = make(map[[32]byte]string)
	listMods()
	err := filepath.Walk("config/", walkConfigs)
	if err != nil {
		log.Fatalln("Failed to list configs:", err.Error())
	}
	err = filepath.Walk("client/", walkClientFiles)
	if err != nil {
		log.Fatalln("Failed to list client files:", err.Error())
	}
	return filesMap
}
