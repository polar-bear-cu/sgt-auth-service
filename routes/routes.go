package routes

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"github.com/polar-bear-cu/sgt-auth-service/controllers"
)

func Register(r *gin.Engine, auth *controllers.AuthController, swaggerEnabled bool) {
	r.GET("/health", controllers.GetHealth)

	if swaggerEnabled {
		r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	}

	v1 := r.Group("/api/v1/auth")
	v1.GET("/google/login", auth.GoogleLogin)
	v1.GET("/google/callback", auth.GoogleCallback)
	v1.POST("/refresh", auth.Refresh)
	v1.POST("/logout", auth.Logout)
}
