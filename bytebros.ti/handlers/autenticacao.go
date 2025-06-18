package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"bytebros.ti/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func RegistrarUsuario(c *gin.Context) {
	log.Printf("DEBUG: Iniciando handler RegistrarUsuario.")
	var user models.Usuario
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("ERRO: Falha ao fazer bind JSON para RegistrarUsuario: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}
	log.Printf("DEBUG: Dados do usuário recebidos: Email=%s, Nome=%s, Telefone=%s", user.Email, user.Nome, user.Telefone)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Senha), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERRO: Falha ao criptografar senha: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criptografar senha"})
		return
	}
	log.Printf("DEBUG: Senha criptografada com sucesso.")

	db := c.MustGet("db").(*sql.DB)
	err = db.QueryRow(`
		INSERT INTO usuarios (nome_completo, email, senha_hash, telefone) -- Incluído 'telefone'
		VALUES ($1, $2, $3, $4) -- Adicionado $4
		RETURNING id`,
		user.Nome, user.Email, string(hashedPassword), user.Telefone).
		Scan(&user.ID)

	if err != nil {
		log.Printf("ERRO BD: Falha ao inserir novo usuário: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao registrar usuário", "detalhes": err.Error()}) // Adicionado detalhes do erro
		return
	}
	log.Printf("DEBUG: Usuário registrado com ID: %d", user.ID)

	token, err := generateJWTToken(user.ID, user.Email, "")
	if err != nil {
		log.Printf("ERRO: Falha ao gerar token JWT para usuário %s: %v", user.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}
	log.Printf("DEBUG: Token JWT gerado com sucesso para usuário %s", user.Email)

	c.JSON(http.StatusCreated, models.LoginResponse{
		ID:       user.ID,
		Nome:     user.Nome,
		Email:    user.Email,
		Token:    token,
		Telefone: user.Telefone,
	})
	log.Printf("DEBUG: Resposta de registro de usuário enviada com sucesso.")
}

func LoginUsuario(c *gin.Context) {
	log.Printf("DEBUG: Iniciando handler LoginUsuario.")
	var login models.LoginRequest
	if err := c.ShouldBindJSON(&login); err != nil {
		log.Printf("ERRO: Falha ao fazer bind JSON para LoginUsuario: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}
	log.Printf("DEBUG: Tentativa de login para email: %s", login.Email)

	db := c.MustGet("db").(*sql.DB)
	var user models.Usuario
	var senhaHashDB string

	log.Printf("DEBUG: Executando query SELECT para usuario com email %s.", login.Email)
	err := db.QueryRow(`
		SELECT id, nome_completo, email, senha_hash, telefone -- Incluído 'telefone'
		FROM usuarios
		WHERE email = $1`, login.Email).
		Scan(&user.ID, &user.Nome, &user.Email, &senhaHashDB, &user.Telefone)

	if err != nil {
		log.Printf("ERRO BD: Falha ao buscar usuário: %v", err)
		handleAuthError(c, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(senhaHashDB), []byte(login.Senha)); err != nil {
		log.Printf("ERRO: Senha incorreta para usuário %s", login.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		return
	}

	token, err := generateJWTToken(user.ID, user.Email, "")
	if err != nil {
		log.Printf("ERRO: Falha ao gerar token JWT para usuário %s: %v", user.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		ID:       user.ID,
		Nome:     user.Nome,
		Email:    user.Email,
		Token:    token,
		Telefone: user.Telefone,
	})
	log.Printf("DEBUG: Resposta de login de usuário enviada com sucesso.")
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
