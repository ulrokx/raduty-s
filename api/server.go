package api

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/ulrokx/raduty-s/api/controllers"
)

var server = controllers.Server{}

func Run() {
	err := godotenv.Load()
	if err != nil {
		log.Println("found no env file")
	} else {
		fmt.Println("loaded from env file")
	}

	server.Initialize(os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

	port := fmt.Sprintf(":%s", os.Getenv("API_PORT"))
	fmt.Printf("listening on port %s", port)

	server.Run(port)
}
