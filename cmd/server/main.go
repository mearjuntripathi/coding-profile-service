package main

import (
	"log"
	"net/http"

	"coding-profile-service/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/stats", handler.StatsHandler)

	log.Println("🚀 Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
 
// package main

// import (
// 	"coding-profile-service/internal/scraper"
// 	"fmt"
// )

// func main() {
// 	userStats, err := scraper.FetchCodeChefHTML("isthisarjun")
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}
// 	fmt.Printf("%+v\n", userStats)
// }
