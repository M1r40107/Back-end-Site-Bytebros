package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
	log.Printf("DEBUG: Dados do usuário recebidos para registro: Nome='%s', Email='%s', Telefone='%s'", user.Nome, user.Email, user.Telefone)

	if user.Nome == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Nome é um campo obrigatório."})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Senha), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERRO: Falha ao criptografar senha: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criptografar senha"})
		return
	}
	log.Printf("DEBUG: Senha criptografada com sucesso.")

	db := c.MustGet("db").(*sql.DB)

	var newUser models.Usuario
	err = db.QueryRow(`
		INSERT INTO usuarios (nome_completo, email, senha_hash, telefone)
		VALUES ($1, $2, $3, $4)
		RETURNING id, nome_completo, email, telefone`,
		user.Nome, user.Email, string(hashedPassword), user.Telefone).
		Scan(&newUser.ID, &newUser.Nome, &newUser.Email, &newUser.Telefone)

	if err != nil {
		log.Printf("ERRO BD: Falha ao inserir novo usuário: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao registrar usuário", "detalhes": err.Error()})
		return
	}
	log.Printf("DEBUG: Usuário registrado com ID: %d. Nome após DB: '%s', Telefone após DB: '%s'", newUser.ID, newUser.Nome, newUser.Telefone)

	token, err := generateJWTToken(newUser.ID, newUser.Email, "")
	if err != nil {
		log.Printf("ERRO: Falha ao gerar token JWT para usuário %s: %v", newUser.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	nomeParaResposta := newUser.Nome
	if nomeParaResposta == "" {
		nomeParaResposta = "Usuário Padrão"
	}

	c.JSON(http.StatusCreated, models.LoginResponse{
		ID:       newUser.ID,
		Nome:     nomeParaResposta,
		Email:    newUser.Email,
		Token:    token,
		Telefone: newUser.Telefone,
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
	var telefoneDB sql.NullString // Temporário para ler o telefone do banco

	log.Printf("DEBUG: Executando query SELECT para usuario com email %s.", login.Email)
	err := db.QueryRow(`
		SELECT id, nome_completo, email, senha_hash, telefone
		FROM usuarios
		WHERE email = $1`, login.Email).
		Scan(&user.ID, &user.Nome, &user.Email, &senhaHashDB, &telefoneDB)

	if err != nil {
		log.Printf("ERRO BD: Falha ao buscar usuário: %v", err)
		handleAuthError(c, err)
		return
	}

	user.Telefone = telefoneDB.String

	nomeParaResposta := user.Nome
	if nomeParaResposta == "" {
		nomeParaResposta = "Usuário Padrão"
	}

	log.Printf("DEBUG: Usuario lido do DB: ID=%d, Nome='%s', Email='%s', Telefone='%s'", user.ID, nomeParaResposta, user.Email, user.Telefone)

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
		Nome:     nomeParaResposta,
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

func AtualizarEmailUsuario(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	claims, exists := c.Get("jwt_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token JWT ausente ou inválido."})
		return
	}
	jwtClaims := claims.(jwt.MapClaims)
	emailLogado := jwtClaims["email"].(string)

	var req models.AtualizarEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	if req.EmailAtual != emailLogado {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Email atual incorreto. Use o email com o qual você está logado."})
		return
	}

	if req.NovoEmail != req.ConfirmarEmail {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "O novo email e a confirmação não coincidem."})
		return
	}

	var senhaHashDB string
	var telefoneDBTemp sql.NullString // Temporário para ler telefone
	var userID int
	err := db.QueryRow(`SELECT id, senha_hash, telefone FROM usuarios WHERE email = $1`, emailLogado).Scan(&userID, &senhaHashDB, &telefoneDBTemp)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Usuário não encontrado."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao verificar usuário."})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(senhaHashDB), []byte(req.Senha)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Senha incorreta."})
		return
	}

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM usuarios WHERE email = $1 AND id != $2`, req.NovoEmail, userID).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao verificar novo email."})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"erro": "Este novo email já está em uso por outra conta."})
		return
	}

	_, err = db.Exec(`UPDATE usuarios SET email = $1, atualizado_em = $2 WHERE id = $3`, req.NovoEmail, time.Now(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar o email."})
		return
	}

	newToken, err := generateJWTToken(userID, req.NovoEmail, "")
	if err != nil {
		log.Printf("ERRO: Falha ao gerar novo token JWT para usuário %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Email alterado, mas falha ao gerar novo token."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensagem":   "Email atualizado com sucesso! Por favor, use o novo email para futuros logins.",
		"novo_email": req.NovoEmail,
		"token":      newToken,
	})
}

func AtualizarTelefoneUsuario(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	claims, exists := c.Get("jwt_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token JWT ausente ou inválido."})
		return
	}
	jwtClaims := claims.(jwt.MapClaims)
	emailLogado := jwtClaims["email"].(string)

	var req models.AtualizarTelefoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	if req.NovoTelefone != req.ConfirmarTelefone {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "O novo telefone e a confirmação não coincidem."})
		return
	}

	var senhaHashDB string
	var telefoneAtualDBTemp sql.NullString // Temporário para ler telefone do banco
	var userID int
	err := db.QueryRow(`SELECT id, senha_hash, telefone FROM usuarios WHERE email = $1`, emailLogado).Scan(&userID, &senhaHashDB, &telefoneAtualDBTemp)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Usuário não encontrado."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao verificar usuário."})
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(senhaHashDB), []byte(req.Senha)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Senha incorreta."})
		return
	}

	// Converte telefoneAtualDBTemp para string para a validação.
	telefoneAtual := telefoneAtualDBTemp.String

	// Validação de correspondência
	if telefoneAtual != req.TelefoneAtual && req.TelefoneAtual != "" {
		c.JSON(http.StatusConflict, gin.H{"erro": "O telefone atual fornecido não corresponde ao registrado."})
		return
	}
	// Se o telefone no DB é vazio e o usuário preencheu o campo "atual",
	// isso pode ser uma tentativa de adicionar o primeiro telefone.
	// Se a coluna telefone no DB é NOT NULL, o req.TelefoneAtual NÃO PODE SER VAZIO se telefoneAtual é vazio.
	// O seu database.go atual para usuarios.telefone é VARCHAR(20) e NÃO NOT NULL.
	// Então, telefoneAtual pode ser "".

	_, err = db.Exec(`UPDATE usuarios SET telefone = $1, atualizado_em = $2 WHERE id = $3`, req.NovoTelefone, time.Now(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar o telefone."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensagem":      "Telefone atualizado com sucesso!",
		"novo_telefone": req.NovoTelefone,
	})
}

func ListarUsuarios(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	busca := c.Query("busca")

	query := `SELECT id, nome_completo, email, telefone FROM usuarios`
	args := []interface{}{}
	whereClauses := []string{}
	argCounter := 1

	if busca != "" {
		buscaPattern := "%" + strings.ToLower(busca) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(email) LIKE $%d OR LOWER(telefone) LIKE $%d)", argCounter, argCounter+1))
		args = append(args, buscaPattern, buscaPattern)
		argCounter += 2
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY nome_completo ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar usuários", "detalhes": err.Error()})
		return
	}
	defer rows.Close()

	var usuarios []models.Usuario
	// Inicializa como slice vazio (não nil) para garantir que sempre serializa para []
	usuarios = make([]models.Usuario, 0)

	for rows.Next() {
		var u models.Usuario
		var telefoneSQL sql.NullString // Variável temporária para ler telefone
		// AQUI: Leia nome_completo para u.Nome e telefone para telefoneSQL
		if err := rows.Scan(&u.ID, &u.Nome, &u.Email, &telefoneSQL); err != nil {
			log.Printf("ERRO: Erro ao ler dados do usuário durante Scan: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler dados do usuário", "detalhes": err.Error()})
			return
		}
		u.Telefone = telefoneSQL.String // Converte para string do modelo

		usuarios = append(usuarios, u)
	}

	c.JSON(http.StatusOK, usuarios)
}
