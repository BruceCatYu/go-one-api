package utils

import (
	"github.com/goccy/go-json"
)

func BytesToObj[T any](data []byte) (T, error) {
	var obj T
	err := json.Unmarshal(data, &obj)
	return obj, err
}

var bs = []byte{32}

func MarshalToSseData(data any) ([]byte, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return append(bs, jsonBytes...), nil
}
