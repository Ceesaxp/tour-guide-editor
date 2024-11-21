// internal/models/user.go
package models

type User struct {
	Username string
	Password string // In production, use proper password hashing
}
