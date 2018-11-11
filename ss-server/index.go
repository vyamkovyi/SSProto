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

func fileHash(path string) ([32]byte, error) {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return [32]byte{}, err
	}
	return blake2b.Sum256(blob), nil
}

func index(record indexPath) error {
	var err error

	fi, err := os.Stat(record.Path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		hash, err := fileHash(record.Path)
		if err != nil {
			return err
		}

		res := IndexedFile{record.Path, record.ClientPath, hash, !record.Sync}
		filesMap[hash] = res
		filepathMap[record.Path] = hash
		watch(filepath.Dir(record.Path))
		return nil
	}

	watch(record.Path)
	if record.Recursive {
		err = filepath.Walk(record.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || strings.Contains(path, "ignored_") {
				return nil
			}
			for _, v := range serverConfig.Ignored {
				if strings.Contains(path, v) {
					return nil
				}
			}

			hash, err := fileHash(path)
			if err != nil {
				return err
			}

			watch(filepath.Dir(path))
			rel, err := filepath.Rel(record.Path, path)
			if err != nil {
				return err
			}
			res := IndexedFile{path, filepath.Join(record.ClientPath, rel), hash, !record.Sync}
			filesMap[res.Hash] = res
			filepathMap[res.ServPath] = res.Hash
			return nil
		})
	} else {
		files, err := ioutil.ReadDir(record.Path)
		if err != nil {
			return err
		}
		for _, f := range files {
			if f.IsDir() || strings.Contains(f.Name(), "ignored_") {
				continue
			}

			fullFileName := filepath.Join(record.Path, f.Name())
			hash, err := fileHash(fullFileName)
			if err != nil {
				return err
			}

			res := IndexedFile{fullFileName, filepath.Join(record.ClientPath, f.Name()), hash, !record.Sync}
			filesMap[res.Hash] = res
			filepathMap[res.ServPath] = res.Hash
		}
	}
	return err
}

func ListFiles() {
	filesMap = make(map[[32]byte]IndexedFile)
	filepathMap = make(map[string][32]byte)

	// ==> Basic directories setup is here.
	for _, v := range serverConfig.Index {
		err := index(v)
		if err != nil {
			log.Println("Something went wrong during indexing:", err)
		}
	}
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

	log.Println("fsnotify event", ev)

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

	// Basically, most of settings are isolated in ListFiles so we don't know what
	// to do here. Our only rescue is to rebuild index using ListFiles itself.
	//
	// However we can't even do it here. If something creates files x, y, z we will get
	// separate event for each thus rebuilding index 3 times what is expensive.
	// Instead we mark existing index as "out-of-date" and rebuild it later (either
	// when client connects or after 5 seconds).
	log.Println("Reindexing scheduled.")

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
	for {
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
