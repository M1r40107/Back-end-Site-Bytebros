package models

import "time"

type Pedido struct {
	ID              int          `json:"id"`
	ClienteEmail    string       `json:"cliente_email"`
	DataPedido      time.Time    `json:"data_pedido"`
	Status          string       `json:"status"`
	EnderecoEntrega string       `json:"endereco_entrega"`
	TipoFrete       string       `json:"tipo_frete"`
	ValorFrete      float64      `json:"valor_frete"`
	ValorTotal      float64      `json:"valor_total"`
	FormaPagamento  string       `json:"forma_pagamento"`
	PrazoEntrega    string       `json:"prazo_entrega"`
	Itens           []PedidoItem `json:"itens"`
	CriadoEm        time.Time    `json:"criado_em"`
}

type PedidoItem struct {
	ID            int     `json:"id"`
	PedidoID      int     `json:"pedido_id"`
	ProdutoID     int     `json:"produto_id"`
	NomeProduto   string  `json:"nome_produto"`
	Quantidade    int     `json:"quantidade"`
	ValorUnitario float64 `json:"valor_unitario"`
}

type CriarPedidoRequest struct {
	Itens           []PedidoItemRequest `json:"itens" binding:"required"`
	EnderecoEntrega string              `json:"endereco_entrega" binding:"required"`
	TipoFrete       string              `json:"tipo_frete" binding:"required"`
	ValorFrete      float64             `json:"valor_frete" binding:"required,min=0"`
	ValorTotal      float64             `json:"valor_total" binding:"required,min=0"`
	FormaPagamento  string              `json:"forma_pagamento" binding:"required"`
	PrazoEntrega    string              `json:"prazo_entrega"`
}

type PedidoItemRequest struct {
	ProdutoID     int     `json:"produto_id" binding:"required"`
	NomeProduto   string  `json:"nome_produto" binding:"required"`
	Quantidade    int     `json:"quantidade" binding:"required,min=1"`
	ValorUnitario float64 `json:"valor_unitario" binding:"required,min=0"`
}
