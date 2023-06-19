package structs

import (
	"errors"
	"fmt"
)

type Storage struct {
	KV map[string]*[]byte
}

func NewStorage() *Storage {
	s := &Storage{}
	s.KV = make(map[string]*[]byte)
	return s
}

func (s *Storage) Create(key []byte, data *[]byte) error {
	//fmt.Println("INTO Create Method:", key)
	id := string(key)
	//fmt.Println("The id is:", id)
	_, exists := s.KV[id]
	if exists {
		return errors.New("the key already exists")
	}

	s.KV[id] = data
	//fmt.Println("The stored value in KV is: ", s.KV[id])
	return nil
}

func (s *Storage) Read(key []byte, start int64, end int64) (*[]byte, error) {
	v, exists := s.KV[string(key)]
	if !exists {
		return nil, errors.New("the key is not found")
	}
	if end == 0 {
		end = int64(len(*v))
	}
	result := (*v)[start:end]
	//fmt.Println("Result is: ", result)
	return &result, nil
}

func (s *Storage) Delete(key []byte) error {
	id := string(key)

	_, exists := s.KV[id]
	if !exists {
		return errors.New("the key is not found")
	}

	delete(s.KV, id)
	return nil
}

func (s *Storage) GetKeys() [][]byte {
	keys := [][]byte{}
	for k, v := range s.KV {
		fmt.Println("Inside GetKeys", k, v)
		keyStr := []byte(k)
		fmt.Println(keyStr)
		keys = append(keys, keyStr)
	}
	fmt.Println(keys)
	return keys
}
