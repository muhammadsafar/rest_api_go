package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

func ResponseTimeMiddleware(next http.Handler) http.Handler {

	fmt.Println("Response time Middleware...")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Response time Middleware being returned...")
		start := time.Now()

		//create a custom response writer add status code
		wrappedWriter := &responseWriterCustom{ResponseWriter: w, status: http.StatusOK}

		//Calculate the duration
		duration := time.Since(start)
		w.Header().Set("X-Response-Time", duration.String())
		next.ServeHTTP(wrappedWriter, r)

		//Log the request details
		duration = time.Since(start)
		fmt.Printf("Methode : %s, Url: %s, Status %d, Duration %v\n", r.Method, r.URL, wrappedWriter.status, duration)
		fmt.Println("Sent Response from Response Time Middleware")

	})
}

// responseWriterCustom
type responseWriterCustom struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriterCustom) WriterHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
