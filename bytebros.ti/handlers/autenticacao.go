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

func RegistrarUsuario(c *gin.Context) {
	var user models.Usuario
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	if err := checkEmailExists(c, user.Email, "usuarios"); err != nil {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Senha), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criptografar senha"})
		return
	}

	db := c.MustGet("db").(*sql.DB)
	err = db.QueryRow(`
        INSERT INTO usuarios (nome, email, senha)
        VALUES ($1, $2, $3)
        RETURNING id, criado_em`,
		user.Nome, user.Email, string(hashedPassword)).
		Scan(&user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao registrar usuário"})
		return
	}

	token, err := generateJWTToken(user.ID, user.Email, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusCreated, models.FuncionarioResponse{
		ID:    user.ID,
		Nome:  user.Nome,
		Email: user.Email,
		Token: token,
	})
}

func LoginUsuario(c *gin.Context) {
	var login models.LoginRequest
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)
	var user models.Usuario

	err := db.QueryRow(`
        SELECT id, nome, email, senha 
        FROM usuarios 
        WHERE email = $1`, login.Email).
		Scan(&user.ID, &user.Nome, &user.Email, &user.Senha)

	if err != nil {
		handleAuthError(c, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Senha), []byte(login.Senha)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		return
	}

	token, err := generateJWTToken(user.ID, user.Email, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusOK, models.FuncionarioResponse{
		ID:    user.ID,
		Nome:  user.Nome,
		Email: user.Email,
		Token: token,
	})
}

func checkEmailExists(c *gin.Context, email, table string) error {
	db := c.MustGet("db").(*sql.DB)
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM "+table+" WHERE email = $1", email).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao verificar email"})
		return err
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email já registrado"})
		return err
	}
	return nil
}

func generateJWTToken(id int, email, cargo string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": id,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 8).Unix(),
	}

	if cargo != "" {
		claims["cargo"] = cargo
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func handleAuthError(c *gin.Context, err error) {
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao autenticar"})
	}
}

func ObterPerfil(c *gin.Context) {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Usuário não autenticado"})
		return
	}

	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao obter informações do token"})
		return
	}

	if cargo, exists := jwtClaims["cargo"]; exists {
		c.JSON(http.StatusOK, gin.H{
			"id":    jwtClaims["user_id"],
			"email": jwtClaims["email"],
			"cargo": cargo,
			"tipo":  "funcionario",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    jwtClaims["user_id"],
		"email": jwtClaims["email"],
		"tipo":  "usuario",
	})
}
