package routes

import (
	"github.com/RG-7/go-auth/controllers"
	"github.com/RG-7/go-auth/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine){
	router.POST("/signup",controllers.Signup())
	router.POST("/login",controllers.Login())

	protected :=router.Group("/")
	protected.Use(middleware.Authenticate())
	{
		protected.GET("/users", controllers.GetUsers())
		protected.GET("/user/:id", controllers.GetUser())
	}
}