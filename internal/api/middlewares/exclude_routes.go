package middlewares

import (
	"fmt"
	"net/http"
	"strings"
)

func MiddlewareExcludePaths(middleware func(http.Handler) http.Handler, excludedPath ...string) func(http.Handler) http.Handler {
	fmt.Println("Start MiddlewareExcludePaths initialized")
	return func(next http.Handler) http.Handler {
		fmt.Println("+++++++++++++++++++MiddlewareExcludePaths RAN")
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				for _, path := range excludedPath {
					if strings.HasPrefix(r.URL.Path, path) {
						next.ServeHTTP(w, r)
						return

					}
				}
				middleware(next).ServeHTTP(w, r)
				fmt.Println("sent reponse from MiddlewareExcludePaths RAN")
			})

	}
}
