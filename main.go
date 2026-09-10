package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/polar-bear-cu/sgt-auth-service/clients"
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

	userConn, err := grpc.NewClient(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = userConn.Close() }()

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	refresh := repositories.NewRefreshTokenPostgres(pool)
	userClient := clients.NewUserClient(userConn)
	uc := usecases.NewAuth(oauthCfg, userClient, cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL, refresh)
	authCtrl := controllers.NewAuth(uc)

	r := gin.Default()
	routes.Register(r, authCtrl)

	log.Println("listening :" + cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
