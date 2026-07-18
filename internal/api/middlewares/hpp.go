package middlewares

import (
	"fmt"
	"net/http"
	"strings"
)

type HPPOptions struct {
	CheckQuery               bool
	CheckBody                bool
	CheckBodyFormContentType string
	Whitelist                []string
}

func Hpp(option HPPOptions) func(http.Handler) http.Handler {
	fmt.Println("HPP Middleware...")

	return func(next http.Handler) http.Handler {
		fmt.Println("HPP Middleware being returned...")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if option.CheckBody && r.Method == http.MethodPost && isCorrectContentType(r, option.CheckBodyFormContentType) {
				//filter the body params
				filterBodyParams(r, option.Whitelist)
			}
			if option.CheckQuery && r.URL.Query() != nil {
				//filter the query params
				filterQueryParams(r, option.Whitelist)
			}

			next.ServeHTTP(w, r)
			fmt.Println("HPP Middleware ends...")

		})

	}
}

func isCorrectContentType(r *http.Request, contentType string) bool {
	return strings.Contains(r.Header.Get("Content-Type"), contentType)
}

func filterBodyParams(r *http.Request, whitelist []string) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return
	}

	for k, v := range r.Form {
		if len(v) > 1 {
			// r.Form.Set(k, v[0]) //first val
			r.Form.Set(k, v[len(v)-1]) //last value
		}

		if !isWhitelisted(k, whitelist) {
			delete(r.Form, k)
		}
	}
}

func filterQueryParams(r *http.Request, whitelist []string) {
	query := r.URL.Query()

	for k, v := range query {
		if len(v) > 1 {
			// query.Set(k, v[0]) //first val
			query.Set(k, v[len(v)-1]) //last value
		}

		if !isWhitelisted(k, whitelist) {
			query.Del(k)
		}
	}

	r.URL.RawQuery = query.Encode()
}

func isWhitelisted(param string, whitelist []string) bool {
	for _, v := range whitelist {
		if param == v {
			return true
		}
	}

	return false
}
