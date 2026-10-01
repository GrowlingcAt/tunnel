package zjwt

import "github.com/golang-jwt/jwt/v5"

type JwtToken interface {
	Sign(data jwt.Claims) (string, error)
	Verify(sign string, data jwt.Claims) error
}
