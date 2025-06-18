package models

import "time"

type Orcamento struct {
	ID           int       `json:"id"`
	NomeCliente  string    `json:"nome_cliente"`
	EmailCliente string    `json:"email_cliente"`
	Telefone     string    `json:"telefone"`
	Descricao    string    `json:"descricao"`
	ServicoNome  string    `json:"servico_nome"`
	Status       string    `json:"status"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

type CriarOrcamentoRequest struct {
	NomeCliente  string `json:"nome_cliente" binding:"required"`
	EmailCliente string `json:"email_cliente" binding:"required,email"`
	Telefone     string `json:"telefone" binding:"required"`
	Descricao    string `json:"descricao" binding:"required"`
	ServicoNome  string `json:"servico_nome"`
}

type AtualizarStatusOrcamentoRequest struct {
	Status string `json:"status" binding:"required,oneof=pendente em_analise aprovado rejeitado"`
}
