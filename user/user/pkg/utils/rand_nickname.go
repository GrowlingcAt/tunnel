package utils

import (
	"fmt"
	"math/rand"
)

const chars1 = "1ghrs2inopq3abcjklmdef045vwxy67tuz89"
const nicknamechars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func toBase36(num int64) string {
	result := ""
	for num > 0 {
		result = string(chars1[num%36]) + result
		num /= 36
	}
	return result
}
func RandNickName(userID int64) string {
	nickname := toBase36(userID)
	if len(nickname) < 8 {
		nickname = fmt.Sprintf("%s%s", generateRandomString(nicknamechars, 8-len(nickname)), nickname)
	}
	return nickname
}

func generateRandomString(charsIn string, length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charsIn[rand.Intn(len(charsIn))]
	}
	return string(b)
}
