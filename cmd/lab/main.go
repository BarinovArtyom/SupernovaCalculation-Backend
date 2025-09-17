package main

import (
	"log"

	"lab1/internal/api"
)

func main() {
	log.Println("app start")
	api.StartServer()
	log.Println("app term")
}
