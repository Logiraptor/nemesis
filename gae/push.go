package gae

import (
	"time"

	"github.com/alexjlockwood/gcm"

	"appengine/urlfetch"

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
	gcm_key   string
	cert_path string
	db        DB
}

func NewPush(db DB, gcm_key, ios_cert string) *Push {
	return &Push{
		gcm_key:   gcm_key,
		cert_path: ios_cert,
		db:        db,
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

func (p *Push) RegisterIOS(token string, parent *datastore.Key) error {
	_, err := p.db.Put(&Device{
		Token:  token,
		Parent: parent,
		Type:   iosDevice,
	})
	return err
}

func (p *Push) SendAll(message string, data map[string]interface{}) error {
	q := datastore.NewQuery(p.db.Kind(&Device{}))

	var devs []Device
	_, err := p.db.GetAll(q, &devs)
	if err != nil {
		return err
	}

	return p.BatchSend(message, data, devs)
}

func (p *Push) SendMessage(message string, data map[string]interface{}, parent *datastore.Key) error {
	q := datastore.NewQuery(p.db.Kind(Device{})).Ancestor(parent)
	var tokens []Device
	_, err := p.db.GetAll(q, &tokens)
	if err != nil {
		return err
	}
	p.db.Context().Errorf("Found %d devices: %v", len(tokens), tokens)
	return p.BatchSend(message, data, tokens)
}

func (p *Push) BatchSend(message string, data map[string]interface{}, devs []Device) error {
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
		if data == nil {
			data = make(map[string]interface{})
		}
		data["Message"] = message
		client := urlfetch.Client(p.db.Context())
		msg := gcm.NewMessage(data, androidTokens...)
		sender := &gcm.Sender{ApiKey: p.gcm_key, Http: client}
		_, err := sender.Send(msg, 3)
		if err != nil {
			return err
		}
	}

	if len(iosTokens) > 0 {
		client := NewAPNSClient(p.db.Context(), p.cert_path)
		for _, token := range iosTokens {
			notif := Notification{
				Device: token,
				Payload: Payload{
					APS: APS{
						Alert: message,
						Badge: -1,
					},
				},
				Expiration: time.Now().Add(time.Hour),
				Lazy:       false,
			}
			var err = client.SendAPNS(notif)
			for i := 0; i < 3 && err != nil; i++ {
				err = client.SendAPNS(notif)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
