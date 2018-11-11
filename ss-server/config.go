// config.go - loading magical snowflakes and numbers from config file!
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
	"github.com/BurntSushi/toml"
	"os"
)

type indexPath struct {
	Path string `toml:"path"`

	// ClientPath determines where the file must be stored on a client side.
	// If this string begins with !, resulting client path will be Path with stripped ClientPath prefix.
	ClientPath string `toml:"client_path"`

	// Sync defines whether file must be kept in sync with client.
	// false means that file must be present on client but is NOT required to be in sync
	Sync bool `toml:"mandatory"`

	// Recursive determines whether specified path must be indexed recursively.
	// This has no effect on files.
	Recursive bool `toml:"recursive"`
}

// Config is a structure with configurable data for ss-server application
type Config struct {
	// Address is a server address to bind server to.
	// Syntax: <ip>:<port>
	Address string `toml:"server_address"`
	// ServerName is a server domain, used in SSL
	ServerName string `toml:"server_name"`

	Certificate string `toml:"ssl_cert"`
	Key         string `toml:"ssl_key"`

	Index []indexPath `toml:"index"`

	// A collection of snowflakes! ❄️
	// Ignored contains files that must not be indexed and sent to client.
	Ignored []string `toml:"ignored"`
}

// NewConfig initializes a Config instance with some default values
func (c *Config) NewConfig() {
	c.Address = "0.0.0.0:48879"
	c.ServerName = "hexawolf.me"
	c.Certificate = "cert.pem"
	c.Key = "key.pem"
	c.Ignored = []string{
		"shadowfacts",
		"FastAsyncWorldEdit",
	}
	c.Index = []indexPath{
		{
			Path:       "config",
			ClientPath: "config",
			Recursive:  true,
			Sync:       false,
		},
		{
			Path:       "mods",
			ClientPath: "mods",
			Recursive:  false,
			Sync:       true,
		},
		{
			Path:       "client",
			ClientPath: ".",
			Recursive:  true,
			Sync:       false,
		},
	}
}

// LoadConfig reads given file and constructs this Config object
func (c *Config) LoadConfig(file string) error {
	configFile, err := os.Open("ssserver.toml")
	if err != nil {
		if os.IsNotExist(err) {
			configFile, err = os.Create("ssserver.toml")
			if err != nil {
				return err
			}
			serverConfig.NewConfig()
			enc := toml.NewEncoder(configFile)
			enc.Encode(serverConfig)
		} else {
			return err
		}
	}
	defer configFile.Close()
	_, err = toml.DecodeReader(configFile, c)
	return err
}
