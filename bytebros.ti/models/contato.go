package models

import "time"

type Contato struct {
	ID           int       `json:"id"`
	Nome         string    `json:"nome" binding:"required"`
	Email        string    `json:"email" binding:"required,email"`
	Mensagem     string    `json:"mensagem" binding:"required"`
	Status       string    `json:"status"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

type CriarContatoRequest struct {
	Nome     string `json:"nome" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Mensagem string `json:"mensagem" binding:"required"`
}

type AtualizarStatusContatoRequest struct {
	Status string `json:"status" binding:"required,oneof=pendente respondido fechado"`
}
