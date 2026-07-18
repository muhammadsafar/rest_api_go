package middlewares

import (
	"fmt"
	"net/http"
)

func SecurityHeaders(next http.Handler) http.Handler {

	fmt.Println("Security Headers Middleware...")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Security Headers Middleware being returned...")

		w.Header().Set("X-DNS-Prefetch-Control", "off") //for security, prevent DNS prefetching

		w.Header().Set("X-Frame-Options", "DENY")                                                  //for security, prevent clickjacking
		w.Header().Set("X-XSS-Protection", "1; mode=block")                                        //for security, prevent cross-site scripting (XSS) attacks
		w.Header().Set("X-Content-Type-Options", "nosniff")                                        //for security, prevent MIME type sniffing
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains;preload") //for security, enforce HTTPS
		w.Header().Set("Content-Security-Policy", "default-src 'self'")                            //for security, prevent cross-site scripting (XSS) attacks and other code injection attacks
		w.Header().Set("Referrer-Policy", "no-referrer")                                           //for security, control the amount of information sent in the Referer header
		w.Header().Set("X-Powered-By", "go")                                                       //examlpe

		next.ServeHTTP(w, r)
		fmt.Println("Security Headers Middleware ends...")
	})
}

//BASIC MIDDLEWARE SKELETON
// func securityHeaders(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	next.ServeHTTP(w, r)
//	})
//}
