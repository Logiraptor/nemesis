package gae

import (
	"net/http"

	"appengine"

	"github.com/mjibson/goon"

	"code.google.com/p/go.crypto/bcrypt"
)

const (
	bcryptCost = 10
)

type User interface {
	SetUsername(s string)
	SetPassword(p []byte)
	Username() string
	Password() []byte
}

type BaseUser struct {
	Name string
	Pass []byte `json:"-"`
}

func (b *BaseUser) SetUsername(s string) { b.Name = s }
func (b *BaseUser) SetPassword(p []byte) { b.Pass = p }
func (b *BaseUser) Username() string     { return b.Name }
func (b *BaseUser) Password() []byte     { return b.Pass }

// InsertUser inserts a user into the datastore.
// It expects a request with parameters:
// - username
// - password
func CreateUser(ctx appengine.Context, req *http.Request, u User) (User, error) {
	username := req.FormValue("username")
	password := req.FormValue("password")

	buf, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}

	u.SetUsername(username)
	u.SetPassword(buf)

	g := goon.FromContext(ctx)
	_, err = g.Put(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}
