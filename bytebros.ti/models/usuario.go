package models

type Usuario struct {
	ID       int    `json:"id"`
	Nome     string `json:"nome" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Senha    string `json:"senha" binding:"required,min=6"`
	Telefone string `json:"telefone"`
}

type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required,min=6"`
}

type LoginResponse struct {
	ID       int    `json:"id"`
	Nome     string `json:"nome"`
	Email    string `json:"email"`
	Token    string `json:"token,omitempty"`
	Telefone string `json:"telefone,omitempty"`
}
