package gae

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

	"appengine"
	"appengine/socket"
)

var (
	apnsAddr = "76.72.22.33"
	// apnsAddr     = "gateway.sandbox.push.apple.com"
	port = ":8081"
	// port         = ":2195"
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
	finished   chan error
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

func NewAPNSClient(ctx appengine.Context, pem string) *APNSClient {
	return &APNSClient{
		ctx: ctx,
		pem: pem,
	}
}

func (a *APNSClient) SendAPNS(n Notification) error {
	var err error
	apnsInitSync.Do(func() {
		notifChannel = make(chan Notification)
		err = a.openConn()
	})
	if err != nil {
		return err
	}

	n.ctx = a.ctx
	n.finished = make(chan error)
	notifChannel <- n

	return <-n.finished
}

func (a *APNSClient) dial() (*socket.Conn, net.Conn, error) {
	gaeConn, err := socket.Dial(a.ctx, "tcp", apnsAddr+port)
	if err != nil {
		return nil, nil, err
	}
	certificate, err := LoadPemFile(a.pem)
	if err != nil {
		return nil, nil, err
	}

	certs := []tls.Certificate{certificate}
	conf := &tls.Config{
		Certificates: certs,
	}

	apnsConn := tls.Client(gaeConn, conf)

	return gaeConn, apnsConn, nil
}

func (a *APNSClient) openConn() error {
	gaeConn, apnsConn, err := a.dial()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case n := <-notifChannel:
				a.ctx.Infof("Sending apns: %#v", n)
				gaeConn.SetContext(n.ctx)
				_, err := n.WriteTo(apnsConn)
				n.finished <- err
				if err != nil {
					a.ctx.Infof("APNS error encountered %s, reconnecting", err.Error())
					apnsConn.Close()
					gaeConn, apnsConn, err = a.dial()
					if err != nil {
						a.ctx.Infof("apns reconnect: %s", err.Error())
						return
					}
				}
				a.ctx.Infof("Finished sending apns")
			case <-time.After(time.Minute):
				log.Println("resetting apns daemon due to inactivity")
				apnsConn.Close()
				apnsInitSync = sync.Once{}
				return
			}
		}
	}()

	return nil
}

// LoadPemFile reads a combined certificate+key pem file into memory.
func LoadPemFile(pemFile string) (cert tls.Certificate, err error) {
	pemBlock, err := ioutil.ReadFile(pemFile)
	if err != nil {
		return
	}
	return LoadPem(pemBlock)
}

// LoadPem is similar to tls.X509KeyPair found in tls.go except that this
// function reads all blocks from the same file.
func LoadPem(pemBlock []byte) (cert tls.Certificate, err error) {
	var block *pem.Block
	for {
		block, pemBlock = pem.Decode(pemBlock)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, block.Bytes)
		} else {
			break
		}
	}

	///////////////////////////////////////////////////////////////////////////
	// The rest of the code in this function is copied from the tls.X509KeyPair
	// implementation found at http://golang.org/src/pkg/crypto/tls/tls.go,
	// with the exception of minor changes (no need to decode the next block).
	///////////////////////////////////////////////////////////////////////////

	if len(cert.Certificate) == 0 {
		err = errors.New("crypto/tls: failed to parse certificate PEM data")
		return
	}

	if block == nil {
		err = errors.New("crypto/tls: failed to parse key PEM data")
		return
	}

	// OpenSSL 0.9.8 generates PKCS#1 private keys by default, while
	// OpenSSL 1.0.0 generates PKCS#8 keys. We try both.
	var key *rsa.PrivateKey
	if key, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		var privKey interface{}
		if privKey, err = x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
			err = errors.New("crypto/tls: failed to parse key: " + err.Error())
			return
		}

		var ok bool
		if key, ok = privKey.(*rsa.PrivateKey); !ok {
			err = errors.New("crypto/tls: found non-RSA private key in PKCS#8 wrapping")
			return
		}
	}

	cert.PrivateKey = key

	// We don't need to parse the public key for TLS, but we so do anyway
	// to check that it looks sane and matches the private key.
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return
	}

	if x509Cert.PublicKeyAlgorithm != x509.RSA || x509Cert.PublicKey.(*rsa.PublicKey).N.Cmp(key.PublicKey.N) != 0 {
		err = errors.New("crypto/tls: private key does not match public key")
		return
	}

	return
}
