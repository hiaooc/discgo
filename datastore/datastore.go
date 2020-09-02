package datastore

import (
	"encoding/json"
	"io/ioutil"
)

type DataStore struct {
	Responses map[string][]string
}

func Read(filepath string) (*DataStore, error) {
	bytes, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	var dataStore *DataStore

	err = json.Unmarshal(bytes, &dataStore)

	if err != nil {
		return nil, err
	}

	return dataStore, nil
}

func Write(filepath string, dataStore DataStore) error {
	return nil
}
