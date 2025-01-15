/*
  	dela - web TODO list
    Copyright (C) 2023  Kasyanov Nikolay Alexeyevich (Unbewohnte)

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package misc

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
