package inmemory

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
	result, ok := h.data[key]
	if !ok {
		return "", ErrNotFound
	}
	return result, nil
}

func (h *HashTable) Del(key string) error {
	delete(h.data, key)
	return nil
}
