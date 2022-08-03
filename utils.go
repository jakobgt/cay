package cay

import "github.com/gofrs/uuid"

func RandomKeys(size int) []string {
	return RandomKeysWithSuffix(size, "")
}

func RandomKeysWithSuffix(size int, suffix string) []string {
	res := make([]string, size)

	for i := 0; i < size; i++ {
		res[i] = uuid.Must(uuid.NewV4()).String() + suffix
	}

	return res
}

func Simdmap(keys []string) *Map {
	val := NewMap(len(keys))
	for n := 0; n < len(keys); n++ {
		val.Insert(keys[n], []byte("data"))
	}
	return val
}
