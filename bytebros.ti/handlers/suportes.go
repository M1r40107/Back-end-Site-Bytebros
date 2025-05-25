package handlers

import (
	"database/sql"
	"net/http"

	"bytebros.ti/models"

	"github.com/gin-gonic/gin"
)

func CriarMensagemSuporte(c *gin.Context) {
	var suporteReq models.SuporteRequest

	if err := c.ShouldBindJSON(&suporteReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	var suporte models.Suporte
	err := db.QueryRow(`
        INSERT INTO suporte (nome, email, mensagem, status)
        VALUES ($1, $2, $3, 'aberto')
        RETURNING id, criado_em`,
		suporteReq.Nome, suporteReq.Email, suporteReq.Mensagem).
		Scan(&suporte.ID, &suporte.CriadoEm)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao registrar mensagem de suporte"})
		return
	}

	suporte.Nome = suporteReq.Nome
	suporte.Email = suporteReq.Email
	suporte.Mensagem = suporteReq.Mensagem
	suporte.Status = "aberto"

	c.JSON(http.StatusCreated, suporte)
}

func ListarMensagensSuporte(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	status := c.Query("status")

	var query string
	var rows *sql.Rows
	var err error

	if status != "" {
		query = `SELECT id, nome, email, mensagem, status, criado_em FROM suporte WHERE status = $1 ORDER BY criado_em DESC`
		rows, err = db.Query(query, status)
	} else {
		query = `SELECT id, nome, email, mensagem, status, criado_em FROM suporte ORDER BY criado_em DESC`
		rows, err = db.Query(query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar mensagens de suporte"})
		return
	}
	defer rows.Close()

	var mensagens []models.Suporte
	for rows.Next() {
		var s models.Suporte
		if err := rows.Scan(&s.ID, &s.Nome, &s.Email, &s.Mensagem, &s.Status, &s.CriadoEm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler mensagens de suporte"})
			return
		}
		mensagens = append(mensagens, s)
	}

	c.JSON(http.StatusOK, mensagens)
}

func AtualizarStatusSuporte(c *gin.Context) {
	id := c.Param("id")
	var update models.SuporteUpdate

	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	_, err := db.Exec(`
        UPDATE suporte
        SET status = $1
        WHERE id = $2`,
		update.Status, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar status do suporte"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Status atualizado com sucesso"})
}

func ObterMensagemSuporte(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	var suporte models.Suporte
	err := db.QueryRow(`
        SELECT id, nome, email, mensagem, status, criado_em
        FROM suporte
        WHERE id = $1`, id).
		Scan(&suporte.ID, &suporte.Nome, &suporte.Email, &suporte.Mensagem, &suporte.Status, &suporte.CriadoEm)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Mensagem de suporte não encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar mensagem de suporte"})
		}
		return
	}

	c.JSON(http.StatusOK, suporte)
}
