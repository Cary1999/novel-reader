package shared

import (
	"errors"
	"fmt"
)

type Role string
type Scope string

const (
	RoleReader     Role = "reader"
	RoleAuthor     Role = "author"
	RoleReviewer   Role = "reviewer"
	RoleSuperAdmin Role = "super_admin"
)

const (
	ScopeFront Scope = "front"
	ScopeAdmin Scope = "admin"
)

var ErrNotFound = errors.New("not found")

type Actor struct {
	ActorID  int64
	Username string
	Role     Role
	Scope    Scope
}

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

func (a Actor) IsFront() bool {
	return a.Scope == ScopeFront
}

func (a Actor) IsAdmin() bool {
	return a.Scope == ScopeAdmin
}

func (a Actor) IsAuthor() bool {
	return a.Scope == ScopeFront && a.Role == RoleAuthor
}

func (a Actor) IsReviewer() bool {
	return a.Scope == ScopeAdmin && (a.Role == RoleReviewer || a.Role == RoleSuperAdmin)
}

func (a Actor) IsSuperAdmin() bool {
	return a.Scope == ScopeAdmin && a.Role == RoleSuperAdmin
}
