package scraper

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func ReadDocument(client *http.Client, req *http.Request) (*goquery.Document, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
