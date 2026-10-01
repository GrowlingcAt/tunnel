package zjwt

import (
	"github.com/golang-jwt/jwt/v5"
	"user/pkg/log"
)

type HSSignMethod string

const (
	HS256 HSSignMethod = "HS256"
	HS384 HSSignMethod = "HS384"
	HS512 HSSignMethod = "HS512"
)

type hs struct {
	key        string
	signMethod HSSignMethod
}

func NewHs(key string, signMethod HSSignMethod) JwtToken {
	return &hs{
		key:        key,
		signMethod: signMethod,
	}
}

func (hs *hs) getSignMethod() *jwt.SigningMethodHMAC {
	switch hs.signMethod {
	case HS256:
		return jwt.SigningMethodHS256
	case HS384:
		return jwt.SigningMethodHS384
	case HS512:
		return jwt.SigningMethodHS512
	}
	return jwt.SigningMethodHS256
}

func (hs *hs) Sign(data jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(hs.getSignMethod(), data)
	sign, err := token.SignedString([]byte(hs.key))
	if err != nil {
		log.Error(err)
		return "", err
	}
	return sign, nil
}

func (hs *hs) Verify(sign string, data jwt.Claims) error {
	_, err := jwt.ParseWithClaims(sign, data, func(token *jwt.Token) (interface{}, error) {
		return []byte(hs.key), nil
	})
	return err
}
