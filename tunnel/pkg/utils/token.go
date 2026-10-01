package utils

import (
	"crypto/md5"
	"fmt"
)

func GenerateTokenByUserID(userID int64) string {
	base36 := ToBase36(userID)
	randomStr := GenerateRandomString(16)
	str := fmt.Sprintf("%s-%s", base36, randomStr)

	m := md5.New()
	m.Write([]byte(str))
	bs := m.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
