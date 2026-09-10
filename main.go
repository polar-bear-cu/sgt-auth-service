package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/polar-bear-cu/sgt-auth-service/controllers"
	"github.com/polar-bear-cu/sgt-auth-service/repositories"
	"github.com/polar-bear-cu/sgt-auth-service/routes"
	"github.com/polar-bear-cu/sgt-auth-service/usecases"
)

func main() {
	oauthCfg := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	accessTTL := envDuration("ACCESS_TTL", 15*time.Minute)
	refreshTTL := envDuration("REFRESH_TTL", 720*time.Hour)

	refresh := repositories.NewInMemoryRefreshToken()
	uc := usecases.NewAuth(oauthCfg, os.Getenv("JWT_SECRET"), accessTTL, refreshTTL, refresh)
	authCtrl := controllers.NewAuth(uc)

	r := gin.Default()
	routes.Register(r, authCtrl)

	port := envDefault("PORT", "8080")
	log.Println("listening :" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func envDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
