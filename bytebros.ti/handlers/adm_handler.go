package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"bytebros.ti/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func CriarAdministrador(c *gin.Context) {
	var admin models.Administrador
	if err := c.ShouldBindJSON(&admin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	claims, _ := c.Get("jwt_claims")
	jwtClaims := claims.(jwt.MapClaims)
	if !jwtClaims["is_admin"].(bool) {
		c.JSON(http.StatusForbidden, gin.H{"erro": "Acesso negado"})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Senha), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criptografar senha"})
		return
	}

	err = db.QueryRow(`
        INSERT INTO administradores (nome, email, senha, is_admin)
        VALUES ($1, $2, $3, $4)
        RETURNING id, criado_em, atualizado_em`,
		admin.Nome, admin.Email, string(hashedPassword), admin.IsAdmin).
		Scan(&admin.ID, &admin.CriadoEm, &admin.Atualizado)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar administrador"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        admin.ID,
		"nome":      admin.Nome,
		"email":     admin.Email,
		"is_admin":  admin.IsAdmin,
		"criado_em": admin.CriadoEm,
	})
}

func LoginAdmin(c *gin.Context) {
	var login models.AdminLogin
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)
	var admin models.Administrador

	err := db.QueryRow(`
        SELECT id, nome, email, senha, is_admin 
        FROM administradores 
        WHERE email = $1`, login.Email).
		Scan(&admin.ID, &admin.Nome, &admin.Email, &admin.Senha, &admin.IsAdmin)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar administrador"})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Senha), []byte(login.Senha)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": admin.ID,
		"email":    admin.Email,
		"is_admin": admin.IsAdmin,
		"exp":      time.Now().Add(time.Hour * 8).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusOK, models.AdminResponse{
		ID:      admin.ID,
		Nome:    admin.Nome,
		Email:   admin.Email,
		IsAdmin: admin.IsAdmin,
		Token:   tokenString,
	})
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("jwt_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token inválido"})
			c.Abort()
			return
		}

		jwtClaims, ok := claims.(jwt.MapClaims)
		if !ok || !jwtClaims["is_admin"].(bool) {
			c.JSON(http.StatusForbidden, gin.H{"erro": "Acesso restrito a administradores"})
			c.Abort()
			return
		}

		c.Next()
	}
}
