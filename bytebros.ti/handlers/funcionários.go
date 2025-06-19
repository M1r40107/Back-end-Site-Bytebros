package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"bytebros.ti/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func RegistrarFuncionario(c *gin.Context) {
	log.Printf("DEBUG: Iniciando handler RegistrarFuncionario.")
	var funcionario models.Funcionario
	if err := c.ShouldBindJSON(&funcionario); err != nil {
		log.Printf("ERRO: Falha ao fazer bind JSON para RegistrarFuncionario: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}
	log.Printf("DEBUG: Dados do funcionário recebidos: Email=%s, Nome=%s, Cargo=%s", funcionario.Email, funcionario.Nome, funcionario.Cargo)

	db := c.MustGet("db").(*sql.DB)
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM funcionarios WHERE email = $1", funcionario.Email).Scan(&count)
	if err != nil {
		log.Printf("ERRO BD: Falha ao verificar existência de email em 'funcionarios': %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno ao verificar email de funcionário"})
		return
	}
	if count > 0 {
		log.Printf("AVISO: Tentativa de registro de funcionário com email já existente: %s", funcionario.Email)
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email já registrado para funcionário"})
		return
	}
	log.Printf("DEBUG: Email %s não encontrado em funcionários. Prosseguindo com o registro.", funcionario.Email)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(funcionario.Senha), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERRO: Falha ao criptografar senha de funcionário: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criptografar senha de funcionário"})
		return
	}
	log.Printf("DEBUG: Senha de funcionário criptografada com sucesso.")

	err = db.QueryRow(`
        INSERT INTO funcionarios (nome, cargo, email, senha_hash)
        VALUES ($1, $2, $3, $4)
        RETURNING id`,
		funcionario.Nome, funcionario.Cargo, funcionario.Email, string(hashedPassword)).
		Scan(&funcionario.ID)

	if err != nil {
		log.Printf("ERRO BD: Falha ao inserir novo funcionário: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao registrar funcionário"})
		return
	}
	log.Printf("DEBUG: Funcionário registrado com ID: %d", funcionario.ID)

	token, err := generateJWTToken(funcionario.ID, funcionario.Email, funcionario.Cargo)
	if err != nil {
		log.Printf("ERRO: Falha ao gerar token JWT para funcionário %s: %v", funcionario.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token para funcionário"})
		return
	}
	log.Printf("DEBUG: Token JWT gerado com sucesso para funcionário %s", funcionario.Email)

	c.JSON(http.StatusCreated, models.FuncionarioResponse{
		ID:    funcionario.ID,
		Nome:  funcionario.Nome,
		Cargo: funcionario.Cargo,
		Email: funcionario.Email,
		Token: token,
	})
	log.Printf("DEBUG: Resposta de registro de funcionário enviada com sucesso.")
}

func LoginFuncionario(c *gin.Context) {
	log.Printf("DEBUG: Iniciando handler LoginFuncionario.")
	var login models.FuncionarioLogin
	if err := c.ShouldBindJSON(&login); err != nil {
		log.Printf("ERRO: Falha ao fazer bind JSON para LoginFuncionario: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}
	log.Printf("DEBUG: Tentativa de login para funcionário: %s", login.Email)

	db := c.MustGet("db").(*sql.DB)
	var funcionario models.Funcionario

	var senhaHashDB string

	log.Printf("DEBUG: Executando query SELECT para funcionário com email %s.", login.Email)
	err := db.QueryRow(`
        SELECT id, nome, cargo, email, senha_hash
        FROM funcionarios
        WHERE email = $1`, login.Email).
		Scan(&funcionario.ID, &funcionario.Nome, &funcionario.Cargo, &funcionario.Email, &senhaHashDB)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("AVISO: Tentativa de login de funcionário falhou: Email %s não encontrado.", login.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		} else {
			log.Printf("ERRO BD: Erro ao buscar funcionário %s: %v", login.Email, err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao autenticar funcionário"})
		}
		return
	}
	log.Printf("DEBUG: Funcionário %s encontrado. Comparando senhas.", funcionario.Email)

	if err := bcrypt.CompareHashAndPassword([]byte(senhaHashDB), []byte(login.Senha)); err != nil {
		log.Printf("AVISO: Senha inválida para funcionário %s: %v", login.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		return
	}
	log.Printf("DEBUG: Senha correta para funcionário %s. Gerando token.", funcionario.Email)

	token, err := generateJWTToken(funcionario.ID, funcionario.Email, funcionario.Cargo)
	if err != nil {
		log.Printf("ERRO: Falha ao gerar token JWT para funcionário %s: %v", funcionario.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token para funcionário"})
		return
	}
	log.Printf("DEBUG: Token JWT gerado para funcionário %s.", funcionario.Email)

	c.JSON(http.StatusOK, models.FuncionarioResponse{
		ID:    funcionario.ID,
		Nome:  funcionario.Nome,
		Cargo: funcionario.Cargo,
		Email: funcionario.Email,
		Token: token,
	})
	log.Printf("DEBUG: Resposta de login de funcionário enviada com sucesso.")
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
