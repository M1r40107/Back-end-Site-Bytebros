package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"bytebros.ti/models"

	"github.com/gin-gonic/gin"
)

func CriarMensagemSuporte(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	var suporteReq models.SuporteRequest

	if err := c.ShouldBindJSON(&suporteReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	clienteEmail, exists := c.Get("email")
	clienteEmailStr := ""
	if exists && clienteEmail != nil {
		clienteEmailStr = clienteEmail.(string)
	}

	if suporteReq.TipoInteracao == "" {
		suporteReq.TipoInteracao = "suporte"
	}

	var suporte models.Suporte
	err := db.QueryRow(`
        INSERT INTO suporte (nome, email, mensagem, status, tipo_interacao, cliente_email)
        VALUES ($1, $2, $3, 'aberto', $4, $5)
        RETURNING id, criado_em`,
		suporteReq.Nome, suporteReq.Email, suporteReq.Mensagem, suporteReq.TipoInteracao, sql.NullString{String: clienteEmailStr, Valid: clienteEmailStr != ""}).
		Scan(&suporte.ID, &suporte.CriadoEm)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao registrar mensagem de suporte", "detalhes": err.Error()})
		return
	}

	suporte.Nome = suporteReq.Nome
	suporte.Email = suporteReq.Email
	suporte.Mensagem = suporteReq.Mensagem
	suporte.Status = "aberto"
	suporte.TipoInteracao = suporteReq.TipoInteracao
	suporte.ClienteEmail = clienteEmailStr

	c.JSON(http.StatusCreated, suporte)
}

func ListarMensagensSuporte(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	status := c.Query("status")
	tipoInteracao := c.Query("tipo_interacao")
	clienteEmailFilter := c.Query("cliente_email")

	var query string
	var args []interface{}
	argCounter := 1

	query = `SELECT id, nome, email, mensagem, status, tipo_interacao, cliente_email, criado_em FROM suporte `
	whereClauses := []string{}

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, status)
		argCounter++
	}
	if tipoInteracao != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("tipo_interacao = $%d", argCounter))
		args = append(args, tipoInteracao)
		argCounter++
	}
	if clienteEmailFilter != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("cliente_email = $%d", argCounter))
		args = append(args, clienteEmailFilter)
		argCounter++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY criado_em DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar mensagens de suporte", "detalhes": err.Error()})
		return
	}
	defer rows.Close()

	var mensagens []models.Suporte
	for rows.Next() {
		var s models.Suporte
		var clienteEmailSQL sql.NullString
		if err := rows.Scan(&s.ID, &s.Nome, &s.Email, &s.Mensagem, &s.Status, &s.TipoInteracao, &clienteEmailSQL, &s.CriadoEm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler mensagens de suporte", "detalhes": err.Error()})
			return
		}
		s.ClienteEmail = clienteEmailSQL.String
		mensagens = append(mensagens, s)
	}

	c.JSON(http.StatusOK, mensagens)
}

func ListarInteracoesCliente(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	clienteEmail, exists := c.Get("email")
	if !exists || clienteEmail == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Email do usuário não encontrado no token"})
		return
	}
	clienteEmailStr := clienteEmail.(string)

	var interacoes []interface{}
	interacoes = make([]interface{}, 0)

	suporteQuery := `
        SELECT id, nome, email, mensagem, status, tipo_interacao, cliente_email, criado_em
        FROM suporte
        WHERE cliente_email = $1
        ORDER BY criado_em DESC`

	suporteRows, err := db.Query(suporteQuery, clienteEmailStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar interações de suporte/contato", "detalhes": err.Error()})
		return
	}
	defer suporteRows.Close()

	for suporteRows.Next() {
		var s models.Suporte
		var clienteEmailSQL sql.NullString
		if err := suporteRows.Scan(&s.ID, &s.Nome, &s.Email, &s.Mensagem, &s.Status, &s.TipoInteracao, &clienteEmailSQL, &s.CriadoEm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler interações de suporte/contato", "detalhes": err.Error()})
			return
		}
		s.ClienteEmail = clienteEmailSQL.String
		interacoes = append(interacoes, s)
	}

	orcamentoQuery := `
        SELECT id, nome_cliente, email_cliente, telefone, descricao, servico_nome, status, criado_em, atualizado_em
        FROM orcamentos
        WHERE email_cliente = $1
        ORDER BY criado_em DESC`

	orcamentoRows, err := db.Query(orcamentoQuery, clienteEmailStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar orçamentos do cliente", "detalhes": err.Error()})
		return
	}
	defer orcamentoRows.Close()

	for orcamentoRows.Next() {
		var o models.Orcamento
		if err := orcamentoRows.Scan(&o.ID, &o.NomeCliente, &o.EmailCliente, &o.Telefone, &o.Descricao, &o.ServicoNome, &o.Status, &o.CriadoEm, &o.AtualizadoEm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler dados do orçamento do cliente", "detalhes": err.Error()})
			return
		}
		interacoes = append(interacoes, map[string]interface{}{
			"id":             o.ID,
			"nome":           o.NomeCliente,
			"email":          o.EmailCliente,
			"mensagem":       o.Descricao,
			"status":         o.Status,
			"criado_em":      o.CriadoEm.Format(time.RFC3339),
			"tipo_interacao": "orcamento",
			"servico_nome":   o.ServicoNome,
			"telefone":       o.Telefone,
			"atualizado_em":  o.AtualizadoEm.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, interacoes)
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
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar status do suporte", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Status atualizado com sucesso"})
}

func ObterMensagemSuporte(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	var suporte models.Suporte
	var clienteEmailSQL sql.NullString
	err := db.QueryRow(`
        SELECT id, nome, email, mensagem, status, tipo_interacao, cliente_email, criado_em
        FROM suporte
        WHERE id = $1`, id).
		Scan(&suporte.ID, &suporte.Nome, &suporte.Email, &suporte.Mensagem, &suporte.Status, &suporte.TipoInteracao, &clienteEmailSQL, &suporte.CriadoEm)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Mensagem de suporte não encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar mensagem de suporte", "detalhes": err.Error()})
		}
		return
	}
	suporte.ClienteEmail = clienteEmailSQL.String

	c.JSON(http.StatusOK, suporte)
}

func DeletarSuporte(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM suporte WHERE id = $1", id)
	if err != nil {
		log.Printf("ERRO BD: Falha ao deletar mensagem de suporte ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar mensagem de suporte", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Mensagem de suporte deletada com sucesso"})
}
