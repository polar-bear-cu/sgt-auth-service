package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/polar-bear-cu/sgt-auth-service/config"
	"github.com/polar-bear-cu/sgt-auth-service/controllers"
	"github.com/polar-bear-cu/sgt-auth-service/repositories"
	"github.com/polar-bear-cu/sgt-auth-service/routes"
	"github.com/polar-bear-cu/sgt-auth-service/usecases"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := config.ConnectPostgres(ctx, cfg.DB.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	refresh := repositories.NewRefreshTokenPostgres(pool)
	uc := usecases.NewAuth(oauthCfg, cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL, refresh)
	authCtrl := controllers.NewAuth(uc)

	r := gin.Default()
	routes.Register(r, authCtrl)

	log.Println("listening :" + cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
