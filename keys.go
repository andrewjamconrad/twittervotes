package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type keys struct {
	TwitterKey          string `json:"TwitterKey"`
	TwitterSecret       string `json:"TwitterSecret"`
	TwitterAccessToken  string `json:"TwitterAccessToken"`
	TwitterAccessSecret string `json:"TwitterAccessSecret"`
}

func loadKeys(path string) keys {
	file, err := os.Open(path)
	var settings keys
	if err != nil {
		fmt.Println("Error opening config file: ", err.Error())
	}

	jsonParser := json.NewDecoder(file)
	if err = jsonParser.Decode(&settings); err != nil {
		fmt.Println("Error parsing config file: ", err.Error())
	}
	return settings
}
