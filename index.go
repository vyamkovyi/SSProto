package main

import (
	"path/filepath"
	"regexp"
	"os"
	"io/ioutil"
	"crypto/sha256"
)

var excludedGlob = []string{
	"/?ignored_*",
	"assets",
	"screenshots",
	"saves",
	"library",
}

func shouldExclude(path string) bool {
	for _, pattern := range excludedGlob {
		if match, _ := regexp.MatchString(pattern, filepath.ToSlash(path)); match {
			return true
		}
	}
	return false
}

func collectRecurse(root string) ([]string, error) {
	var res []string
	walkfn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if shouldExclude(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldExclude(path) {
			return nil
		}

		res = append(res, path)
		return nil
	}
	err := filepath.Walk(root, walkfn)
	return res, err
}

func collectHashList() (map[string][]byte, error) {
	res := make(map[string][]byte)

	list, err := collectRecurse(".")
	if err != nil {
		return nil, err
	}

	authlib := "libraries/com/mojang/authlib/1.5.25/authlib-1.5.25.jar"
	if fileExists(authlib) {
		list = append(list, filepath.ToSlash(authlib))
	}

	for _, path := range list {
		blob, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(blob)
		res[path] = sum[:]
	}
	return res, nil
}
