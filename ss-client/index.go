// index.go - client files hashing ("indexing")
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
	"os"
	"path/filepath"
	"regexp"
)

// excludedGlob is a collection of snowflakes ❄️ that must not be hashes
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
