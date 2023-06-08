package utils

import "crypto/sha1"

func GetSha1Hash(data string) []byte {
	dataBytes := []byte(data)
	sha := sha1.Sum(dataBytes)
	keyHash := sha[:]
	return keyHash
}
