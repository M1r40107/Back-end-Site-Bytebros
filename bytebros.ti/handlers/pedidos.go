package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"bytebros.ti/models"
	"github.com/gin-gonic/gin"
)

func CriarPedido(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	clienteEmail, exists := c.Get("email")
	if !exists || clienteEmail == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Email do usuário não encontrado no token"})
		return
	}
	clienteEmailStr := clienteEmail.(string)

	var req models.CriarPedidoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao iniciar transação do pedido"})
		return
	}
	defer tx.Rollback()

	var pedidoID int
	err = tx.QueryRow(`
		INSERT INTO pedidos (cliente_email, status, endereco_entrega, tipo_frete, valor_frete, valor_total, forma_pagamento, prazo_entrega)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		clienteEmailStr, "Processando", req.EnderecoEntrega, req.TipoFrete, req.ValorFrete, req.ValorTotal, req.FormaPagamento, req.PrazoEntrega).
		Scan(&pedidoID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar pedido principal", "detalhes": err.Error()})
		return
	}

	for _, itemReq := range req.Itens {
		_, err := tx.Exec(`
			INSERT INTO pedido_itens (pedido_id, produto_id, nome_produto, quantidade, valor_unitario)
			VALUES ($1, $2, $3, $4, $5)`,
			pedidoID, itemReq.ProdutoID, itemReq.NomeProduto, itemReq.Quantidade, itemReq.ValorUnitario)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao inserir item do pedido", "detalhes": err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao comitar transação do pedido"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"mensagem": "Pedido criado com sucesso!", "pedido_id": pedidoID})
}

func ListarPedidosCliente(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	clienteEmail, exists := c.Get("email")
	if !exists || clienteEmail == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Email do usuário não encontrado no token"})
		return
	}
	clienteEmailStr := clienteEmail.(string)

	rows, err := db.Query(`
		SELECT id, cliente_email, data_pedido, status, endereco_entrega, tipo_frete, valor_frete, valor_total, forma_pagamento, prazo_entrega
		FROM pedidos
		WHERE cliente_email = $1
		ORDER BY data_pedido DESC`, clienteEmailStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar pedidos do cliente", "detalhes": err.Error()})
		return
	}
	defer rows.Close()

	var pedidos []models.Pedido
	pedidos = make([]models.Pedido, 0)

	for rows.Next() {
		var p models.Pedido
		if err := rows.Scan(&p.ID, &p.ClienteEmail, &p.DataPedido, &p.Status, &p.EnderecoEntrega, &p.TipoFrete, &p.ValorFrete, &p.ValorTotal, &p.FormaPagamento, &p.PrazoEntrega); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler pedido do cliente", "detalhes": err.Error()})
			return
		}
		pedidos = append(pedidos, p)
	}

	for i := range pedidos {
		itemRows, err := db.Query(`
            SELECT id, pedido_id, produto_id, nome_produto, quantidade, valor_unitario
            FROM pedido_itens
            WHERE pedido_id = $1`, pedidos[i].ID)
		if err != nil {
			pedidos[i].Itens = []models.PedidoItem{}
			continue
		}
		defer itemRows.Close()

		var itens []models.PedidoItem
		itens = make([]models.PedidoItem, 0)
		for itemRows.Next() {
			var pi models.PedidoItem
			if err := itemRows.Scan(&pi.ID, &pi.PedidoID, &pi.ProdutoID, &pi.NomeProduto, &pi.Quantidade, &pi.ValorUnitario); err != nil {
				continue
			}
			itens = append(itens, pi)
		}
		pedidos[i].Itens = itens
	}

	c.JSON(http.StatusOK, pedidos)
}

func ListarPedidosAdmin(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	status := c.Query("status")
	clienteEmailFilter := c.Query("cliente_email")

	query := `
        SELECT id, cliente_email, data_pedido, status, endereco_entrega, tipo_frete, valor_frete, valor_total, forma_pagamento, prazo_entrega
        FROM pedidos `

	args := []interface{}{}
	whereClauses := []string{}
	argCounter := 1

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, status)
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

	query += " ORDER BY data_pedido DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar pedidos (admin)", "detalhes": err.Error()})
		return
	}
	defer rows.Close()

	var pedidos []models.Pedido
	for rows.Next() {
		var p models.Pedido
		if err := rows.Scan(&p.ID, &p.ClienteEmail, &p.DataPedido, &p.Status, &p.EnderecoEntrega, &p.TipoFrete, &p.ValorFrete, &p.ValorTotal, &p.FormaPagamento, &p.PrazoEntrega); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler pedido (admin)", "detalhes": err.Error()})
			return
		}
		pedidos = append(pedidos, p)
	}

	for i := range pedidos {
		itemRows, err := db.Query(`
            SELECT id, pedido_id, produto_id, nome_produto, quantidade, valor_unitario
            FROM pedido_itens
            WHERE pedido_id = $1`, pedidos[i].ID)
		if err != nil {
			pedidos[i].Itens = []models.PedidoItem{}
			continue
		}
		defer itemRows.Close()

		var itens []models.PedidoItem
		for itemRows.Next() {
			var pi models.PedidoItem
			if err := itemRows.Scan(&pi.ID, &pi.PedidoID, &pi.ProdutoID, &pi.NomeProduto, &pi.Quantidade, &pi.ValorUnitario); err != nil {
				continue
			}
			itens = append(itens, pi)
		}
		pedidos[i].Itens = itens
	}

	c.JSON(http.StatusOK, pedidos)
}

func AtualizarStatusPedido(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	pedidoID := c.Param("id")

	var update struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	_, err := db.Exec(`UPDATE pedidos SET status = $1 WHERE id = $2`, update.Status, pedidoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar status do pedido", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Status do pedido atualizado com sucesso"})
}

func DeletarPedido(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	pedidoID := c.Param("id")

	_, err := db.Exec(`DELETE FROM pedidos WHERE id = $1`, pedidoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar pedido", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Pedido deletado com sucesso"})
}
