package datastore

import (
	"encoding/json"
	"io/ioutil"
)

type DataStore struct {
	FilePath string
	Contents Contents
}

type Contents struct {
	Responses map[string][]string `json:"responses"`
}

func (ds *DataStore) Save() error {
	return Write(ds.FilePath, ds.Contents)
}

func Read(filepath string) (*DataStore, error) {
	bytes, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	var contents Contents
	err = json.Unmarshal(bytes, &contents)

	if err != nil {
		return nil, err
	}

	return &DataStore{
		FilePath: filepath,
		Contents: contents,
	}, nil
}

func Write(filepath string, c Contents) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath, data, 0600)
	if err != nil {
		return err
	}

	return nil
}
