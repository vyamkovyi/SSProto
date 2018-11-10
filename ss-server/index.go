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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/crypto/blake2b"
)

// IndexedFile represents essential data shipped with the file during update.
type IndexedFile struct {
	// Where file is located on server (absolute).
	ServPath string
	// Where file should be placed on client (relative to client root directory).
	ClientPath string

	Hash [32]byte

	// If true - file will be not replaced at client if it's already present
	// (even if changed).
	ShouldNotReplace bool
}

var filesMap map[[32]byte]IndexedFile
var filesMapLock sync.RWMutex
var filepathMap map[string][32]byte
var reindexTimer *time.Timer
var reindexRequired = false
var watcher *fsnotify.Watcher

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
	return blake2b.Sum256(blob), nil
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

	watch(dir)
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

			hash, err := fileHash(path)
			if err != nil {
				return err
			}

			watch(filepath.Dir(path))
			res = append(res, IndexedFile{path, path, hash, shouldNotReplace})
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

			fullFileName := filepath.Join(dir, f.Name())
			hash, err := fileHash(fullFileName)
			if err != nil {
				return nil, err
			}

			res = append(res, IndexedFile{fullFileName, fullFileName, hash, shouldNotReplace})
		}
	}
	return res, err
}

func addFile(servPath, clientPath string, shouldNotReplace bool) error {
	hash, err := fileHash(servPath)
	if err != nil {
		return err
	}

	filesMap[hash] = IndexedFile{servPath, clientPath, hash, shouldNotReplace}
	filepathMap[servPath] = hash
	watch(filepath.Dir(servPath))
	return nil
}

func ListFiles() {
	filesMap = make(map[[32]byte]IndexedFile)
	filepathMap = make(map[string][32]byte)

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
			filesMap[entry.Hash] = entry
			filepathMap[entry.ServPath] = entry.Hash
		}
	}

	// ==> Per-file overrides go here. Call addFile for them.
	addFile("client/options.txt", "options.txt", true)
	addFile("client/optionsof.txt", "optionsof.txt", true)
}

func watch(path string) {
	// We will catch changes in all files in directory we watch.
	abs, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		log.Println("Failed to convert to abs path:", err)
		return
	}
	if err := watcher.Add(abs); err != nil {
		log.Println("Failed to add watcher for", abs+":", err)
	}
}

func processFsnotifyEvent(ev fsnotify.Event) {
	if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
		// We don't care.
		return
	}

	if ev.Op&fsnotify.Create == fsnotify.Create {
		stat, err := os.Stat(ev.Name)
		if err != nil {
			log.Println("Failed to stat file/dir received in event:", err)
			return
		}
		if stat.IsDir() {
			log.Println("New directory:", ev.Name+"; watching it too...")
			// fsnotify (inotify actually) doesn't supports recursive watching of
			// subdirectories so we should add each manually.
			watcher.Add(ev.Name)
			return
		}
		// We don't need to anything other than adding watcher when
		// directory is created. We will receive another CREATE event
		// event for each file in created directory.
	}

	if ev.Op&fsnotify.Remove == fsnotify.Remove {
		log.Println("File/directory removed:", ev.Name)
		// We don't know if this was a directory or not.
		// However try to remove it from watcher just in case.
		watcher.Remove(ev.Name)

		// We don't have to rebuild entire index, just remove entry.
		filesMapLock.Lock()
		hash, prs := filepathMap[ev.Name] // ev.Name is already absolute path
		if !prs {
			return
		}

		delete(filesMap, hash)
		delete(filepathMap, ev.Name)
		filesMapLock.Unlock()
		return
	}

	log.Println("fsnotify event", ev)

	// Basically, most of settings are isolated in ListFiles so we don't know what
	// to do here. Our only rescue is to rebuild index using ListFiles itself.
	//
	// However we can't even do it here. If something creates files x, y, z we will get
	// separate event for each thus rebuilding index 3 times what is expensive.
	// Instead we mark existing index as "out-of-date" and rebuild it later (either
	// when client connects or after 5 seconds).
	filesMapLock.Lock()
	if reindexTimer == nil {
		reindexTimer = time.NewTimer(5 * time.Second)
		go deferredIndexRebuild()
	}
	reindexTimer.Reset(5 * time.Second)
	reindexRequired = true
	filesMapLock.Unlock()
}

func deferredIndexRebuild() {
	<-reindexTimer.C
	filesMapLock.Lock()
	if reindexRequired {
		log.Println("Reindexing files...")
		ListFiles()
		seenIDsMtx.Lock()
		seenIDs = make(map[string]struct{}) // reset seen IDs
		seenIDsMtx.Unlock()
		reindexTimer.Stop()
		reindexRequired = false
		log.Println("Reindexing done")
	}
	filesMapLock.Unlock()
}

func handleFSEvents() {
	for {
		select {
		case ev, ok := <-watcher.Events:
			if !ok {
				return
			}
			processFsnotifyEvent(ev)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("fsnotify error:", err)
		}
	}
}
