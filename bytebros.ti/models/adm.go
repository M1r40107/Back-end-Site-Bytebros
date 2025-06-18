package models

import "time"

type Administrador struct {
	ID         int       `json:"id"`
	Nome       string    `json:"nome" binding:"required,min=3"`
	Email      string    `json:"email" binding:"required,email"`
	Senha      string    `json:"senha" binding:"required,min=8"`
	IsAdmin    bool      `json:"is_admin"`
	CriadoEm   time.Time `json:"criado_em"`
	Atualizado time.Time `json:"atualizado_em"`
}

type AdminLogin struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required,min=8"`
}

type AdminResponse struct {
	ID      int    `json:"id"`
	Nome    string `json:"nome"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	Token   string `json:"token"`
}
