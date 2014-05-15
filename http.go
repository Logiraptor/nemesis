package nemesis

import (
	"bufio"
	"bytes"
	"net/http"
	"net/url"
	"strings"
)

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
