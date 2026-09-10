package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/polar-bear-cu/sgt-auth-service/controllers"
)

func Register(r *gin.Engine, auth *controllers.AuthController) {
	r.GET("/health", controllers.GetHealth)

	v1 := r.Group("/api/v1/auth")
	v1.GET("/google/login", auth.GoogleLogin)
	v1.GET("/google/callback", auth.GoogleCallback)
	v1.POST("/refresh", auth.Refresh)
	v1.POST("/logout", auth.Logout)
}
