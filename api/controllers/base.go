package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ulrokx/raduty-s/api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	DB     *gorm.DB
	Router *gin.Engine
}

func (s *Server) Initialize(DBDriver, DBUser, DBPass, DBPort, DBHost, DBName string) {
	if DBDriver == "postgres" {
		var err error
		DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DBHost, DBPort, DBUser, DBName, DBPass)
		s.DB, err = gorm.Open(postgres.Open(DBURL))
		if err != nil {
			log.Fatalf("could not connect to database err: %s", err)
		} else {
			fmt.Println("connected to database")
		}
	} else {
		fmt.Println("unknown driver")
	}

	s.DB.Debug().AutoMigrate(
		&models.Assistant{},
		&models.Unavailable{},
		&models.Schedule{},
		&models.Shift{},
	)

	s.Router = gin.Default()
	s.Router.Use(cors.New(cors.Config{
		AllowHeaders: []string{"*"},
		AllowOrigins: []string{"*"},
	}))
	s.initializeRoutes()

}

func (s *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, s.Router))
}
