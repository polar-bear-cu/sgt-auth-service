package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/polar-bear-cu/sgt-auth-service/dtos"
	"github.com/polar-bear-cu/sgt-auth-service/usecases"
)

const stateCookie = "oauth_state"

type AuthController struct {
	uc        *usecases.AuthUsecase
	publicURL string
}

func NewAuth(uc *usecases.AuthUsecase, publicURL string) *AuthController {
	return &AuthController{uc: uc, publicURL: publicURL}
}

// GoogleLogin godoc
// @Summary  start Google OAuth flow
// @Tags     auth
// @Success  302
// @Router   /api/v1/auth/google/login [get]
func (ctl *AuthController) GoogleLogin(c *gin.Context) {
	state := randomState()
	c.SetCookie(stateCookie, state, 600, "/", "", false, true)
	c.Redirect(http.StatusFound, ctl.uc.LoginURL(state))
}

// GoogleCallback godoc
// @Summary  OAuth callback, redirects to the frontend with the token pair
// @Tags     auth
// @Param    code   query  string  true  "authorization code"
// @Param    state  query  string  true  "csrf state"
// @Success  302
// @Router   /api/v1/auth/google/callback [get]
func (ctl *AuthController) GoogleCallback(c *gin.Context) {
	want, err := c.Cookie(stateCookie)
	if err != nil || want == "" || c.Query("state") != want {
		ctl.callbackError(c, "invalid_state")
		return
	}
	c.SetCookie(stateCookie, "", -1, "/", "", false, true)

	code := c.Query("code")
	if code == "" {
		ctl.callbackError(c, "missing_code")
		return
	}

	pair, err := ctl.uc.HandleCallback(c.Request.Context(), code)
	if err != nil {
		ctl.callbackError(c, "exchange_failed")
		return
	}
	ctl.callbackSession(c, pair)
}

func (ctl *AuthController) callbackSession(c *gin.Context, pair usecases.TokenPair) {
	resp := toTokenResponse(pair)
	fragment := url.Values{
		"access_token":  {resp.AccessToken},
		"refresh_token": {resp.RefreshToken},
		"expires_in":    {strconv.Itoa(resp.ExpiresIn)},
		"token_type":    {resp.TokenType},
	}
	c.Redirect(http.StatusFound, ctl.publicURL+"/auth/callback#"+fragment.Encode())
}

func (ctl *AuthController) callbackError(c *gin.Context, reason string) {
	c.Redirect(http.StatusFound, ctl.publicURL+"/login?error="+url.QueryEscape(reason))
}

// Refresh godoc
// @Summary  rotate refresh token, issue new pair
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body  body      dtos.RefreshRequest  true  "refresh token"
// @Success  200   {object}  dtos.TokenResponse
// @Failure  400   {object}  map[string]string
// @Failure  401   {object}  map[string]string
// @Router   /api/v1/auth/refresh [post]
func (ctl *AuthController) Refresh(c *gin.Context) {
	var req dtos.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := ctl.uc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, usecases.ErrRefreshInvalid) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toTokenResponse(pair))
}

// Logout godoc
// @Summary  revoke refresh token
// @Tags     auth
// @Accept   json
// @Param    body  body  dtos.RefreshRequest  true  "refresh token"
// @Success  204
// @Failure  400  {object}  map[string]string
// @Router   /api/v1/auth/logout [post]
func (ctl *AuthController) Logout(c *gin.Context) {
	var req dtos.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctl.uc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func toTokenResponse(p usecases.TokenPair) dtos.TokenResponse {
	return dtos.TokenResponse{
		AccessToken:  p.AccessToken,
		RefreshToken: p.RefreshToken,
		ExpiresIn:    p.ExpiresIn,
		TokenType:    "Bearer",
	}
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
