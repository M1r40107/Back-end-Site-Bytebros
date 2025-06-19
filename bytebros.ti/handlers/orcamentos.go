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

func CriarOrcamento(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	var req models.CriarOrcamentoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	var orcamentoID int
	err := db.QueryRow(`
		INSERT INTO orcamentos (nome_cliente, email_cliente, telefone, descricao, servico_nome, status, criado_em, atualizado_em)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		req.NomeCliente, req.EmailCliente, req.Telefone, req.Descricao, req.ServicoNome, "pendente", time.Now(), time.Now()).
		Scan(&orcamentoID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar orçamento", "detalhes": err.Error()}) //
		return
	}

	c.JSON(http.StatusCreated, gin.H{"mensagem": "Orçamento criado com sucesso!", "id": orcamentoID}) //
}

func ListarOrcamentos(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	statusFilter := c.Query("status")
	emailFilter := c.Query("email")

	query := `
		SELECT id, nome_cliente, email_cliente, telefone, descricao, servico_nome, status, criado_em, atualizado_em
		FROM orcamentos
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
		whereClauses = append(whereClauses, fmt.Sprintf("email_cliente = $%d", argCounter))
		args = append(args, emailFilter)
		argCounter++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY criado_em DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao listar orçamentos", "detalhes": err.Error()})
		return
	}
	defer rows.Close()

	var orcamentos []models.Orcamento
	for rows.Next() {
		var o models.Orcamento
		if err := rows.Scan(&o.ID, &o.NomeCliente, &o.EmailCliente, &o.Telefone, &o.Descricao, &o.ServicoNome, &o.Status, &o.CriadoEm, &o.AtualizadoEm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler dados do orçamento", "detalhes": err.Error()})
			return
		}
		orcamentos = append(orcamentos, o)
	}

	c.JSON(http.StatusOK, orcamentos)
}

func ObterOrcamento(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	var orcamento models.Orcamento
	err := db.QueryRow(`
		SELECT id, nome_cliente, email_cliente, telefone, descricao, servico_nome, status, criado_em, atualizado_em
		FROM orcamentos
		WHERE id = $1`, id).
		Scan(&orcamento.ID, &orcamento.NomeCliente, &orcamento.EmailCliente, &orcamento.Telefone, &orcamento.Descricao, &orcamento.ServicoNome, &orcamento.Status, &orcamento.CriadoEm, &orcamento.AtualizadoEm)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Orçamento não encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar orçamento", "detalhes": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, orcamento)
}

func AtualizarStatusOrcamento(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	var req models.AtualizarStatusOrcamentoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	_, err := db.Exec(`
		UPDATE orcamentos
		SET status = $1, atualizado_em = $2
		WHERE id = $3`,
		req.Status, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar status do orçamento", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Status do orçamento atualizado com sucesso"})
}

func DeletarOrcamento(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	_, err := db.Exec(`DELETE FROM orcamentos WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar orçamento", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Orçamento deletado com sucesso"})
}
