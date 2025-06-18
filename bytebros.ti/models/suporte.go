package models

import "time"

type Suporte struct {
	ID            int       `json:"id"`
	Nome          string    `json:"nome" binding:"required,min=3"`
	Email         string    `json:"email" binding:"required,email"`
	Mensagem      string    `json:"mensagem" binding:"required,min=10"`
	Status        string    `json:"status"`
	TipoInteracao string    `json:"tipo_interacao"`
	ClienteEmail  string    `json:"cliente_email"`
	CriadoEm      time.Time `json:"criado_em"`
}

type SuporteRequest struct {
	Nome          string `json:"nome"`
	Email         string `json:"email"`
	Mensagem      string `json:"mensagem"`
	TipoInteracao string `json:"tipo_interacao"`
}

type SuporteUpdate struct {
	Status string `json:"status" binding:"required,oneof=aberto em_andamento resolvido"`
}
