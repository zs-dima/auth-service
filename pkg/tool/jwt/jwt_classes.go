package jwt_tool

import (
	"github.com/google/uuid"
)

type key int

const (
	Issuer            = "auth-service"
	Audience          = "auth-service"
	UserClaimsKey key = iota
	TokenLength   int = 54 // To ensure the base64-encoded string is ≤ 72 bytes supported by bcrypt
)

type JwtUserRole string

type JwtUserInfo struct {
	Id    *uuid.UUID
	Role  JwtUserRole
	Email string
}

type JwtAuthInfo struct {
	UserInfo       *JwtUserInfo
	DeviceId       *uuid.UUID
	InstallationId *uuid.UUID
}
