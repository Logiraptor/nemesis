package nemesis

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RW struct {
	io.Writer
}

func (r RW) Header() http.Header {
	return http.Header{}
}

func (r RW) WriteHeader(int) {

}

func ForgeFormPost(url string, values url.Values) (*http.Request, error) {
	req, err := http.NewRequest("POST", url,
		strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	buf := new(bytes.Buffer)
	err = req.Write(buf)
	if err != nil {
		return nil, err
	}
	req, err = http.ReadRequest(bufio.NewReader(buf))
	if err != nil {
		return nil, err
	}
	return req, nil
}
