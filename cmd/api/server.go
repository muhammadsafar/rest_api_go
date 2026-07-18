package main

import (
	"crypto/tls"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"restapi/internal/api/middlewares"
	"restapi/internal/api/router"
	"time"

	"github.com/joho/godotenv"
)

//go:embed .env
var envFile embed.FS

func loadEnvFromEmbeddedFile() {
	//Read the embedded .env file
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading .env file:%v", err)
	}

	//create a temp file to load the env vars

	tempfile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error creating temp .env file:%v", err)
	}

	defer os.Remove(tempfile.Name())

	//Write content of the embedded .en vfile to the time file
	_, err = tempfile.Write(content)
	if err != nil {
		log.Fatalf("Error writing to temp .env file:%v", err)
	}

	err = tempfile.Close()
	if err != nil {
		log.Fatalf("Error close temp .env file:%v", err)
	}

	//Load enc vars from the temp file
	err = godotenv.Load(tempfile.Name())
	if err != nil {
		log.Fatalf("Error loading temp .env file:%v", err)
	}

}

func main() {

	//Only in production, for running source code
	// err := godotenv.Load()
	// if err != nil {
	// 	return
	// }

	// _, err = sqlconnect.ConnectDB()
	// if err != nil {
	// 	utils.ErrorHandler(err, "Error connecting to database")
	// 	return
	// }

	//load env vars from the embedded .env file
	loadEnvFromEmbeddedFile()

	// fmt.Println("Environment variable CERT_FILE:", os.Getenv("CERT_FILE"))

	port := os.Getenv("SERVER_PORT")

	// cert := "cert.pem"
	// key := "key.pem"

	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	fmt.Println("Starting server on port : ", port)
	//using default server
	// err := http.ListenAndServe(port, nil)

	//USING TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		// MinVersion: tls.VersionTLS10,
	}

	rl := middlewares.NewRateLimiter(5, time.Minute)

	hppOptions := middlewares.HPPOptions{
		CheckQuery:               true,
		CheckBody:                true,
		CheckBodyFormContentType: "application/x-www-form-urlencoded",
		Whitelist:                []string{"sortBy", "sortOrder", "name", "age", "class", "country"},
	}

	router := router.MainRouter()

	jwtMiddleware := middlewares.MiddlewareExcludePaths(middlewares.JWTMiddleware,
		"/execs/login",
		"/execs/forgotpassword",
		"/execs/resetpassword/reset",
		// "/execs"
	)

	secureMux := applyMiddlewares(
		router,
		middlewares.SecurityHeaders,
		middlewares.Compression,
		middlewares.Hpp(hppOptions),
		middlewares.XSSMiddleware,
		jwtMiddleware,
		middlewares.ResponseTimeMiddleware,
		rl.RateLimiterMiddleware,
		middlewares.Cors)

	// secureMux := middlewares.SecurityHeaders(router)
	// secureMux := middlewares.JWTMiddleware(middlewares.SecurityHeaders(router))

	// secureMux := jwtMiddleware(middlewares.SecurityHeaders(router))
	// secureMux := middlewares.XSSMiddleware(router)

	// secureMux := middlewares.Cors(
	// 	middlewares.XSSMiddleware(
	// 		jwtMiddleware(
	// 			rl.RateLimiterMiddleware(
	// 				middlewares.ResponseTimeMiddleware(
	// 					middlewares.SecurityHeaders(
	// 						middlewares.Compression(
	// 							middlewares.Hpp(hppOptions)(router))))))))

	//Create a custom server with TLS configuration
	server := &http.Server{
		Addr: port,
		// Handler: mux, //if without mux, set nil
		Handler: secureMux,
		// Handler: router,
		// Handler: logProtocol(router),
		TLSConfig: tlsConfig,
	}

	err := server.ListenAndServeTLS(cert, key)
	// err := server.ListenAndServe()

	if err != nil {
		log.Fatal("Error starting the server:", err)
	}
}

type Middleware func(http.Handler) http.Handler

func applyMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func logProtocol(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Protocol:", r.Proto)
		next.ServeHTTP(w, r)
	})
}
