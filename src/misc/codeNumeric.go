package misc

import (
	"math/rand"
	"strconv"
)

// Generates a pseudo-random numeric code of required length
func GenerateNumericCode(length uint) string {
	code := ""
	for i := 0; uint(i) < length; i++ {
		code += strconv.Itoa(rand.Intn(10))
	}

	return code
}
