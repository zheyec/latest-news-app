package service

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*") 
	(*w).Header().Set("content-type", "application/json") 
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

var middleware = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[/]=>remote=>%s host=>%s   url=>%s   method=>%s\n", r.RemoteAddr, r.Host, r.URL, r.Method)
		next.ServeHTTP(w, r)
	})
}

var (
	// Cwd - current working directory
	Cwd, _ = os.Getwd()
)

//Run HTTP Server
func Run(ctx context.Context, port int) {
	rand.Seed(time.Now().Unix())

	//http router
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("i am working\n"))
	})
	http.Handle("/news_query", http.HandlerFunc(newsQueryHandle))
	http.Handle("/news_card", http.HandlerFunc(newsCardHandle))
	http.Handle("/news", http.HandlerFunc(newsPageHandle))

	ip := fmt.Sprintf(":%v", port)
	fmt.Printf("listen on:%s\n", ip)
	fmt.Println("Current working dir: ", Cwd)
	http.ListenAndServe(ip, nil)
}
