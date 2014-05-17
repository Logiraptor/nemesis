package gae

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"appengine"
	"appengine/socket"
)

var (
	apnsAddr     = "gateway.sandbox.push.apple.com"
	port         = ":2195"
	apnsInitSync sync.Once
	notifChannel chan Notification
)

type APS struct {
	Alert string `json:"alert,omitempty"`
	Badge int    `json:"badge,omitempty"`
	Sound string `json:"string,omitempty"`
}

type Payload struct {
	APS APS `json:"aps"`
}

type Notification struct {
	Device     string
	Payload    Payload
	Identifier int32
	Expiration time.Time
	Lazy       bool
	ctx        appengine.Context
}

// Implemented based on
// https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/CommunicatingWIthAPS.html
func (n *Notification) WriteTo(wr io.Writer) (int64, error) {
	var buf = new(bytes.Buffer)
	buf.WriteByte(2)

	var frame = new(bytes.Buffer)

	// Token
	frame.WriteByte(1)
	token, err := hex.DecodeString(n.Device)
	if err != nil {
		return 0, fmt.Errorf("invalid device token: %s", err.Error())
	}
	binary.Write(frame, binary.BigEndian, int16(len(token)))
	frame.Write(token)

	// Payload
	frame.WriteByte(2)
	payload, err := json.Marshal(n.Payload)
	if err != nil {
		return 0, fmt.Errorf("payload json error: %s", err.Error())
	}
	binary.Write(frame, binary.BigEndian, int16(len(payload)))
	frame.Write(payload)

	// Identifier (arbitrary)
	frame.WriteByte(3)
	binary.Write(frame, binary.BigEndian, int16(4))
	binary.Write(frame, binary.BigEndian, n.Identifier) // 3 is arbitrary unique identifier

	// Expiration
	frame.WriteByte(4)
	binary.Write(frame, binary.BigEndian, int16(4))
	binary.Write(frame, binary.BigEndian, int32(n.Expiration.Unix())) // 0 means do not keep on apns

	// Priority (10 for instant, 5 for lazy)
	frame.WriteByte(5)
	binary.Write(frame, binary.BigEndian, int16(1))
	if n.Lazy {
		binary.Write(frame, binary.BigEndian, byte(5)) // 5 means send now
	} else {
		binary.Write(frame, binary.BigEndian, byte(10)) // 10 means send now
	}

	// Write the frame
	frameLen := frame.Len()
	binary.Write(buf, binary.BigEndian, int32(frameLen))
	frame.WriteTo(buf)

	return buf.WriteTo(wr)
}

type APNSClient struct {
	ctx appengine.Context
	pem string
}

func (a *APNSClient) SendAPNS(n Notification) error {
	apnsInitSync.Do(func() {
		a.openConn()
	})

	n.ctx = a.ctx
	notifChannel <- n

	return nil
}

func (a *APNSClient) openConn() error {
	notifChannel = make(chan Notification)
	gaeConn, err := socket.Dial(a.ctx, "tcp", apnsAddr+port)
	if err != nil {
		return err
	}
	certificate, err := LoadPemFile(cert)
	if err != nil {
		return err
	}

	certs := []tls.Certificate{certificate}
	conf := &tls.Config{
		Certificates: certs,
	}

	apnsConn := tls.Client(gaeConn, conf)
	defer apnsConn.Close()

	for {
		select {
		case n := <-notifChannel:
			gaeConn.SetContext(n.ctx)
			_, err := n.WriteTo(apnsConn)
		}
	}
}
