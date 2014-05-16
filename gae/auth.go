package gae

import (
	"fmt"
	"net/http"

	"github.com/martini-contrib/sessions"

	"appengine/datastore"

	"appengine"

	"code.google.com/p/go.crypto/bcrypt"
)

const (
	bcryptCost = 10
)

type User interface {
	SetID(id int64)
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

	db := DBFromContext(ctx)
	query := datastore.NewQuery(db.Kind(u)).Filter("Name=", username)
	_, err = db.Run(query).Next(u)
	if err != datastore.Done {
		return nil, fmt.Errorf("%s is taken", username)
	}

	u.SetUsername(username)
	u.SetPassword(buf)

	_, err = db.Put(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func LoginUser(ctx appengine.Context, req *http.Request, dst User, session sessions.Session) (User, error) {
	username := req.FormValue("username")
	password := req.FormValue("password")

	db := DBFromContext(ctx)
	query := datastore.NewQuery(db.Kind(dst)).Filter("Name=", username)
	_, err := db.Run(query).Next(dst)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %s", err.Error())
	}

	actual := dst.Password()
	err = bcrypt.CompareHashAndPassword(actual, []byte(password))
	if err != nil {
		return nil, fmt.Errorf("Invalid credentials")
	}

	sess := &Session{
		User: db.Key(dst),
	}
	k, err := db.Put(sess)
	if err != nil {
		return nil, err
	}
	session.Set("token", k.Encode())

	return dst, nil
}

type Session struct {
	ID   int64          `datastore:"-" goon:"id"`
	User *datastore.Key `goon:"parent"`
}

func GetUserFromToken(db DB, token string, dst User) (User, error) {
	k, err := datastore.DecodeKey(token)
	if err != nil {
		return nil, err
	}

	sess := &Session{
		ID:   k.IntID(),
		User: k.Parent(),
	}

	err = db.Get(sess)
	if err != nil {
		return nil, err
	}
	if sess.User != k.Parent() {
		return nil, fmt.Errorf("invalid token")
	}

	dst.SetID(k.Parent().IntID())
	err = db.Get(dst)
	if err != nil {
		return nil, err
	}
	return dst, err
}
