package gae

import (
	"fmt"

	"appengine/urlfetch"

	"github.com/alexjlockwood/gcm"

	"appengine/datastore"
)

const (
	androidDevice = "ANDROID"
	iosDevice     = "IOS"
)

type Device struct {
	Token  string         `datastore:"-" goon:"id"`
	Parent *datastore.Key `datastore:"-" goon:"parent"`
	Type   string
}

type Push struct {
	gcm_key string
	db      DB
}

func NewPush(db DB, gcm_key string) *Push {
	return &Push{
		gcm_key: gcm_key,
		db:      db,
	}
}

func (p *Push) RegisterAndroid(token string, parent *datastore.Key) error {
	_, err := p.db.Put(&Device{
		Token:  token,
		Parent: parent,
		Type:   androidDevice,
	})
	return err
}

func (p *Push) SendAll(data map[string]interface{}) error {
	q := datastore.NewQuery(p.db.Kind(&Device{}))

	var devs []Device
	_, err := p.db.GetAll(q, &devs)
	if err != nil {
		return err
	}

	var androidTokens []string
	var iosTokens []string
	for _, d := range devs {
		switch d.Type {
		case androidDevice:
			androidTokens = append(androidTokens, d.Token)
		case iosDevice:
			iosTokens = append(iosTokens, d.Token)
		}
	}

	if len(androidTokens) > 0 {
		client := urlfetch.Client(p.db.Context())
		msg := gcm.NewMessage(data, androidTokens...)
		sender := &gcm.Sender{ApiKey: p.gcm_key, Http: client}
		_, err = sender.Send(msg, 2)
		if err != nil {
			return err
		}
	}

	if len(iosTokens) > 0 {
		// ......
	}
	return nil
}

func (p *Push) SendMessage(data map[string]interface{}, device *datastore.Key) error {
	d := &Device{
		Token:  device.StringID(),
		Parent: device.Parent(),
	}
	err := p.db.Get(d)
	if err != nil {
		return fmt.Errorf("invalid device key: %s", device)
	}

	switch d.Type {
	case androidDevice:
		client := urlfetch.Client(p.db.Context())
		msg := gcm.NewMessage(data, d.Token)
		sender := &gcm.Sender{ApiKey: p.gcm_key, Http: client}
		_, err = sender.Send(msg, 2)
		if err != nil {
			return err
		}
	case iosDevice:
		// TODO: enqueue a task to a crazy backend thing that
		// nobody understands
		return fmt.Errorf("unimplemented")
	}

	return nil
}
