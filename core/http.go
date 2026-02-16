package core

import "net/http"

func NewClient() *http.Client {
	return &http.Client{}
}

func NewRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "luffy/1.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	return req, nil
}
