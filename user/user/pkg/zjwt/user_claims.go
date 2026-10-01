package zjwt

import "github.com/golang-jwt/jwt/v5"

const EXPIRES_IN int = 15 * 86400

type UserClaims struct {
	*jwt.RegisteredClaims
	UserID int64  `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}
