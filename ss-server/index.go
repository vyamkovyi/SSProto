// index.go - enlisting and hashing of files that need to be present and up to date on client side.
// Copyright (c) 2018  Hexawolf
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
package main

import (
	"crypto/sha256"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// IndexedFile represents essential data shipped with the file during update.
type IndexedFile struct {
	// Where file is located on server (absolute).
	ServPath string
	// Where file should be placed on client (relative to client root directory).
	ClientPath string

	// If true - file will be not replaced at client if it's already present
	// (even if changed).
	ShouldNotReplace bool
}

var filesMap map[string]IndexedFile

// A collection of snowflakes! ❄️
// excludedPaths contains files that must not be indexed and sent to client.
var excludedPaths = []string{
	"shadowfacts",
	"FastAsyncWorldEdit",
}

func fileHash(path string) ([32]byte, error) {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(blob), nil
}

// ExcludeFunc represents any suitable function that checks if passed file path must be excluded from indexing.
// WARNING! Returning true means exclusion of file!
type ExcludeFunc func(string) bool

// allFiles is a ExcludeFunc candidate that excludes only files with ignored_ prefix
func allFiles(path string) bool {
	return strings.Contains(path, "ignored_")
}

// jarOnly is a ExcludeFunc candidate that excludes files with .jar extension or ignored_ prefix
func jarOnly(path string) bool {
	if filepath.Ext(path) != ".jar" {
		return true
	}
	return allFiles(path)
}

func index(dir string, recursive bool, excludeFunc ExcludeFunc, shouldNotReplace bool) ([]IndexedFile, error) {
	var res []IndexedFile
	var err error = nil
	if recursive {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || excludeFunc(path) {
				return nil
			}
			for _, v := range excludedPaths {
				if strings.Contains(path, v) {
					return nil
				}
			}

			res = append(res, IndexedFile{path, path, shouldNotReplace})
			return nil
		})
	} else {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if f.IsDir() || excludeFunc(f.Name()) {
				continue
			}

			fullFileName := dir
			if !(strings.HasSuffix(fullFileName, "/") || strings.HasSuffix(fullFileName, "\\")) {
				fullFileName += string(os.PathSeparator)
			}
			fullFileName += f.Name()

			res = append(res, IndexedFile{fullFileName, fullFileName, shouldNotReplace})
		}
	}
	return res, err
}

func addFile(servPath, clientPath string, shouldNotReplace bool) error {
	if _, err := os.Stat(servPath); err != nil {
		return err
	}

	filesMap[clientPath] = IndexedFile{servPath, clientPath, shouldNotReplace}
	return nil
}

func ListFiles() {
	filesMap = make(map[string]IndexedFile)

	// ==> Basic directories setup is here.
	configs, err := index("config", true, allFiles, false)
	if err != nil {
		log.Println("Failed to read config dir:", err)
	}
	mods, err := index("mods", false, jarOnly, false)
	if err != nil {
		log.Println("Failed to read mods dir:", err)
	}
	client, err := index("client", true, allFiles, false)
	if err != nil {
		log.Println("Failed to read client dir:", err)
	}
	clientCfgs, _ := index("client/config", true, allFiles, true)

	// Strip "client/" prefix from client-side paths.
	for _, part := range [][]IndexedFile{client, clientCfgs} {
		for i := range part {
			part[i].ClientPath = part[i].ClientPath[7:]
		}
	}

	// Put everything into global table.
	for _, part := range [][]IndexedFile{configs, mods, client, clientCfgs} {
		for _, entry := range part {
			filesMap[entry.ClientPath] = entry
		}
	}

	// ==> Per-file overrides go here. Call addFile for them.
	addFile("client/options.txt", "options.txt", true)
	addFile("client/optionsof.txt", "optionsof.txt", true)
}
