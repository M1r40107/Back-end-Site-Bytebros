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

func RegistrarFuncionario(c *gin.Context) {
	var funcReq models.FuncionarioRequest

	if err := c.ShouldBindJSON(&funcReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM funcionarios WHERE email = $1", funcReq.Email).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao verificar email"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email já registrado"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(funcReq.Senha), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criptografar senha"})
		return
	}

	var funcionario models.Funcionario
	err = db.QueryRow(`
        INSERT INTO funcionarios (nome, cargo, email, senha)
        VALUES ($1, $2, $3, $4)
        RETURNING id, nome, cargo, email, criado_em`,
		funcReq.Nome, funcReq.Cargo, funcReq.Email, string(hashedPassword)).
		Scan(&funcionario.ID, &funcionario.Nome, &funcionario.Cargo, &funcionario.Email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao registrar funcionário"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    funcionario.ID,
		"nome":  funcionario.Nome,
		"cargo": funcionario.Cargo,
		"email": funcionario.Email,
	})
}

func LoginFuncionario(c *gin.Context) {
	var login models.FuncionarioLogin

	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	var funcDB models.Funcionario
	err := db.QueryRow(`
        SELECT id, nome, cargo, email, senha 
        FROM funcionarios 
        WHERE email = $1`, login.Email).
		Scan(&funcDB.ID, &funcDB.Nome, &funcDB.Cargo, &funcDB.Email, &funcDB.Senha)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar funcionário"})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(funcDB.Senha), []byte(login.Senha)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"func_id": funcDB.ID,
		"email":   funcDB.Email,
		"cargo":   funcDB.Cargo,
		"exp":     time.Now().Add(time.Hour * 8).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	c.JSON(http.StatusOK, models.FuncionarioResponse{
		ID:    funcDB.ID,
		Nome:  funcDB.Nome,
		Cargo: funcDB.Cargo,
		Email: funcDB.Email,
		Token: tokenString,
	})
}

func ListarFuncionarios(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	rows, err := db.Query(`
        SELECT id, nome, cargo, email, criado_em
        FROM funcionarios
        ORDER BY nome`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar funcionários"})
		return
	}
	defer rows.Close()

	var funcionarios []models.Funcionario
	for rows.Next() {
		var f models.Funcionario
		if err := rows.Scan(&f.ID, &f.Nome, &f.Cargo, &f.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler funcionários"})
			return
		}
		funcionarios = append(funcionarios, f)
	}

	c.JSON(http.StatusOK, funcionarios)
}
