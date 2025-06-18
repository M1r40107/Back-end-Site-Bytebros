package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func AdminDashboard(c *gin.Context) {
	claims, _ := c.Get("jwt_claims")
	jwtClaims := claims.(jwt.MapClaims)

	c.JSON(http.StatusOK, gin.H{
		"mensagem": "Bem-vindo ao painel administrativo",
		"usuario":  jwtClaims["email"],
		"is_admin": jwtClaims["is_admin"],
	})
}
