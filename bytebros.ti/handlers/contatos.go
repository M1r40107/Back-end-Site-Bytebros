package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bytebros.ti/models"
	"github.com/gin-gonic/gin"
)

func CriarContato(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	var req models.CriarContatoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	var contatoID int
	err := db.QueryRow(`
		INSERT INTO contatos (nome, email, mensagem, status, criado_em, atualizado_em)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		req.Nome, req.Email, req.Mensagem, "pendente", time.Now(), time.Now()).
		Scan(&contatoID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar mensagem de contato", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"mensagem": "Mensagem de contato criada com sucesso!", "id": contatoID})
}

func ListarContatos(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	statusFilter := c.Query("status")
	emailFilter := c.Query("email")

	query := `
		SELECT id, nome, email, mensagem, status, criado_em, atualizado_em
		FROM contatos
	`
	args := []interface{}{}
	whereClauses := []string{}
	argCounter := 1

	if statusFilter != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, statusFilter)
		argCounter++
	}
	if emailFilter != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("email = $%d", argCounter))
		args = append(args, emailFilter)
		argCounter++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY criado_em DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao listar mensagens de contato", "detalhes": err.Error()})
		return
	}
	defer rows.Close()

	var contatos []models.Contato
	for rows.Next() {
		var co models.Contato
		if err := rows.Scan(&co.ID, &co.Nome, &co.Email, &co.Mensagem, &co.Status, &co.CriadoEm, &co.AtualizadoEm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler dados da mensagem de contato", "detalhes": err.Error()})
			return
		}
		contatos = append(contatos, co)
	}

	c.JSON(http.StatusOK, contatos)
}

func ObterContato(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	var contato models.Contato
	err := db.QueryRow(`
		SELECT id, nome, email, mensagem, status, criado_em, atualizado_em
		FROM contatos
		WHERE id = $1`, id).
		Scan(&contato.ID, &contato.Nome, &contato.Email, &contato.Mensagem, &contato.Status, &contato.CriadoEm, &contato.AtualizadoEm)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Mensagem de contato não encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar mensagem de contato", "detalhes": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, contato)
}

func AtualizarStatusContato(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	var req models.AtualizarStatusContatoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	_, err := db.Exec(`
		UPDATE contatos
		SET status = $1, atualizado_em = $2
		WHERE id = $3`,
		req.Status, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar status da mensagem de contato", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Status da mensagem de contato atualizado com sucesso"})
}

func DeletarContato(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	_, err := db.Exec(`DELETE FROM contatos WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar mensagem de contato", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Mensagem de contato deletada com sucesso"})
}
