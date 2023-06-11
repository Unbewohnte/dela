package encryption

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// Encodes given string via Base64
func EncodeString(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// Decodes given string via Base64
func DecodeString(encodedStr string) string {
	decodedBytes, _ := base64.StdEncoding.DecodeString(encodedStr)
	return string(decodedBytes)
}

// Returns HEX string of SHA256'd data
func SHA256Hex(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
