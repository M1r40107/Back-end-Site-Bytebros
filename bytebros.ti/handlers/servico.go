package handlers

import (
	"database/sql"
	"net/http"

	"bytebros.ti/models"

	"github.com/gin-gonic/gin"
)

func CriarServico(c *gin.Context) {
	var servicoReq models.ServicoRequest

	if err := c.ShouldBindJSON(&servicoReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	var servico models.Servico
	err := db.QueryRow(`
        INSERT INTO servicos (nome, preco, oferta, detalhes)
        VALUES ($1, $2, $3, $4)
        RETURNING id`,
		servicoReq.Nome, servicoReq.Preco, servicoReq.Oferta, servicoReq.Detalhes).
		Scan(&servico.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar serviço"})
		return
	}

	servico.Nome = servicoReq.Nome
	servico.Preco = servicoReq.Preco
	servico.Oferta = servicoReq.Oferta
	servico.Detalhes = servicoReq.Detalhes

	c.JSON(http.StatusCreated, servico)
}

func ListarServicos(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	somenteOfertas := c.Query("ofertas") == "true"

	var query string
	if somenteOfertas {
		query = `SELECT id, nome, preco, oferta, detalhes FROM servicos WHERE oferta = true ORDER BY nome`
	} else {
		query = `SELECT id, nome, preco, oferta, detalhes FROM servicos ORDER BY nome`
	}

	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar serviços"})
		return
	}
	defer rows.Close()

	var servicos []models.Servico
	for rows.Next() {
		var s models.Servico
		if err := rows.Scan(&s.ID, &s.Nome, &s.Preco, &s.Oferta, &s.Detalhes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler serviços"})
			return
		}
		servicos = append(servicos, s)
	}

	c.JSON(http.StatusOK, servicos)
}

func ObterServico(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	var servico models.Servico
	err := db.QueryRow(`
        SELECT id, nome, preco, oferta, detalhes
        FROM servicos
        WHERE id = $1`, id).
		Scan(&servico.ID, &servico.Nome, &servico.Preco, &servico.Oferta, &servico.Detalhes)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Serviço não encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar serviço"})
		}
		return
	}

	c.JSON(http.StatusOK, servico)
}

func AtualizarServico(c *gin.Context) {
	id := c.Param("id")
	var servicoReq models.ServicoRequest

	if err := c.ShouldBindJSON(&servicoReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	_, err := db.Exec(`
        UPDATE servicos
        SET nome = $1, preco = $2, oferta = $3, detalhes = $4
        WHERE id = $5`,
		servicoReq.Nome, servicoReq.Preco, servicoReq.Oferta, servicoReq.Detalhes, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar serviço"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Serviço atualizado com sucesso"})
}

func DeletarServico(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*sql.DB)

	_, err := db.Exec("DELETE FROM servicos WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar serviço"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Serviço deletado com sucesso"})
}
