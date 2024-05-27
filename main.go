package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gourav-08/go-postgres-stocks-api/router"
)

func main() {
	r := router.Router()
	fmt.Println("Starting server on port 8080")

	log.Fatal(http.ListenAndServe(":8080", r))
}
