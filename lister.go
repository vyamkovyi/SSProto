package main

import (
	"crypto/sha256"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type IndexedFile struct {
	// Where file is located on server (absolute).
	ServPath string
	// Where file should be placed on client (relative to client root directory).
	ClientPath string
	Hash       [32]byte

	// If true - file will be not replaced at client if it's already present
	// (even if changed).
	ShouldNotReplace bool
}

var filesMap map[[32]byte]IndexedFile

func fileHash(path string) ([32]byte, error) {
	blob, err := ioutil.ReadFile(path)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(blob), nil
}

func allFiles(path string) bool {
	return strings.Contains(path, "ignored_")
}
func jarOnly(path string) bool {
	if filepath.Ext(path) != ".jar" {
		return true
	}
	return strings.Contains(path, "ignored_")
}

var excludedPaths = []string{
	"shadowfacts",
}

func index(dir string, recursive bool, excludeFunc func(string) bool, shouldNotReplace bool) ([]IndexedFile, error) {
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

			hash, err := fileHash(path)
			if err != nil {
				return err
			}

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

			fullFileName := dir
			if !(strings.HasSuffix(fullFileName, "/") || strings.HasSuffix(fullFileName, "\\")) {
				fullFileName += string(os.PathSeparator)
			}
			fullFileName += f.Name()

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
	return nil
}

func ListFiles() {
	filesMap = make(map[[32]byte]IndexedFile)

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
		}
	}

	// ==> Per-file overrides go here. Call addFile for them.
	addFile("client/options.txt", "options.txt", true)
	addFile("client/optionsof.txt", "optionsof.txt", true)
}
