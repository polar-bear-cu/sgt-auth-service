package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/polar-bear-cu/sgt-auth-service/clients"
	"github.com/polar-bear-cu/sgt-auth-service/config"
	"github.com/polar-bear-cu/sgt-auth-service/controllers"
	_ "github.com/polar-bear-cu/sgt-auth-service/docs"
	"github.com/polar-bear-cu/sgt-auth-service/repositories"
	"github.com/polar-bear-cu/sgt-auth-service/routes"
	"github.com/polar-bear-cu/sgt-auth-service/usecases"
)

// @title        Auth Service API
// @version      1.0
// @description  Google OAuth2 + JWT for the Subglutee project
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
	routes.Register(r, authCtrl, cfg.SwaggerEnabled)
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go func() {
		log.Println("listening :" + cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
}
