package swole

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	ErrValueTooLong = errors.New("cookie value too long")
)

func unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, item := range values {
		uniqueValues[item] = true
	}

	return len(values) == len(uniqueValues)
}

func writeCookie(w http.ResponseWriter, cookie http.Cookie) error {
	cookie.Value = url.QueryEscape(cookie.Value)

	if len(cookie.String()) > 4096 {
		return ErrValueTooLong
	}

	http.SetCookie(w, &cookie)
	return nil
}

func readCookie(r *http.Request, cookieName string) (*http.Cookie, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}

	unescapedValue, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return nil, err
	}
	cookie.Value = unescapedValue

	return cookie, nil
}
