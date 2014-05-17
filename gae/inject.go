package gae

import (
	"fmt"
	"net/http"

	"github.com/martini-contrib/sessions"

	"github.com/go-martini/martini"

	"github.com/mjibson/appstats"

	"appengine"
)

func InjextContext(req *http.Request, c martini.Context) {
	ctx := appstats.NewContext(req)
	c.MapTo(ctx, (*appengine.Context)(nil))
	c.Next()
	ctx.Save()
}

func InjectDB(c martini.Context, ctx appengine.Context) {
	c.MapTo(DBFromContext(ctx), (*DB)(nil))
}

func InjectUser(dst User) interface{} {
	return func(c martini.Context, session sessions.Session, db DB) (interface{}, error) {
		token := session.Get("token")
		if s, ok := token.(string); ok {
			user, err := GetUserFromToken(db, s, dst)
			if err != nil {
				return nil, fmt.Errorf("you are not authorized to do that: %s", err.Error())
			}

			c.Map(user)
			return nil, nil
		}
		return nil, fmt.Errorf("you are not authorized to do that")
	}

}

func InjectPush(gcm_key, ios_cert string) interface{} {
	return func(c martini.Context, db DB) {
		c.Map(NewPush(db, gcm_key, ios_cert))
	}
}
