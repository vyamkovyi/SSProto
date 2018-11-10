package main

import (
	"encoding/json"
	"os"
)

type indexPath struct {
	Path string `json:"path"`

	// ClientPath determines where the file must be stored on a client side.
	// If this string begins with !, resulting client path will be Path with stripped ClientPath prefix.
	ClientPath string `json:"client_path"`

	// Sync defines whether file must be kept in sync with client.
	// false means that file must be present on client but is NOT required to be in sync
	Sync bool `json:"mandatory"`

	// Recursive determines whether specified path must be indexed recursively.
	// This has no effect on files.
	Recursive bool `json:"recursive"`
}

// Config is a structure with configurable data for ss-server application
type Config struct {
	// Address is a server address to bind server to.
	// Syntax: <ip>:<port>
	Address string `json:"server_address"`
	// ServerName is a server domain, used in SSL
	ServerName string `json:"server_name"`

	Certificate string `json:"ssl_cert"`
	Key         string `json:"ssl_key"`

	Index []indexPath `json:"index"`

	// A collection of snowflakes! ❄️
	// Ignored contains files that must not be indexed and sent to client.
	Ignored []string `json:"ignored"`
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
			ClientPath: "!client",
			Recursive:  true,
			Sync:       false,
		},
	}
}

// LoadConfig reads given file and constructs this Config object
func (c *Config) LoadConfig(file string) error {
	configFile, err := os.Open("ssserver.json")
	if err != nil {
		if os.IsNotExist(err) {
			configFile, err = os.Create("ssserver.json")
			if err != nil {
				return err
			}
			serverConfig.NewConfig()
			jsonstr, _ := json.MarshalIndent(serverConfig, "", "	")
			configFile.Write(jsonstr)
		} else {
			return err
		}
	}
	defer configFile.Close()
	dec := json.NewDecoder(configFile)
	err = dec.Decode(c)
	return err
}

// TODO: WriteConfig
