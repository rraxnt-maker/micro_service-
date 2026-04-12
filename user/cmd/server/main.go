package main

import (
	"log"
	"net/http"
	"user/internal/handler"
)

func main() {
	http.HandleFunc("/profile", handler.GetProfile)
	http.HandleFunc("/profile/update", handler.UpdateProfile)
	http.HandleFunc("/user/", handler.GetUserByID)
	http.HandleFunc("/profile/delete", handler.DeleteProfile)

	http.HandleFunc("/sync", handler.SyncUser)

	log.Println("Server on :8069")
	log.Fatal(http.ListenAndServe(":8069", nil))
}
