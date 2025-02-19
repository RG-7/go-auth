package main

import (
	"log"

	"github.com/RG-7/go-auth/config"
	"github.com/RG-7/go-auth/helpers"
	"github.com/RG-7/go-auth/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	Key := config.GenerateRandomKey()
	helpers.SetJWTKey(Key)

	r := gin.Default()

	routes.SetupRoutes(r)

	// start the server
	r.Run(":8080")
	log.Println(" 🚀🚀 Server is running on port: 8080")
}
