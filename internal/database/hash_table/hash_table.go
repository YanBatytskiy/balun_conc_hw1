package hashtable

import (
	"fmt"

	"lesson1/internal/database/dberrors"
)

type HashTable struct {
	data map[string]string
}

func NewHashTable() *HashTable {
	return &HashTable{
		data: make(map[string]string),
	}
}

func (h *HashTable) Set(key, value string) error {
	h.data[key] = value
	return nil
}

func (h *HashTable) Get(key string) (string, error) {
	const op = "HashTable.Get"

	result, ok := h.data[key]
	if !ok {
		return "", fmt.Errorf("%s: %w", op, dberrors.ErrNotFound)
	}
	return result, nil
}

func (h *HashTable) Del(key string) error {
	const op = "HashTable.Del"

	_, ok := h.data[key]
	if !ok {
		return fmt.Errorf("%s: %w", op, dberrors.ErrNotFound)
	}

	delete(h.data, key)
	return nil
}
