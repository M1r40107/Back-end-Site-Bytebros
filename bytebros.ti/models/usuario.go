package models

type Usuario struct {
	ID       int    `json:"id"`
	Nome     string `json:"nome_completo" binding:"required"`
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

type AtualizarEmailRequest struct {
	EmailAtual     string `json:"email_atual" binding:"required,email"`
	NovoEmail      string `json:"novo_email" binding:"required,email"`
	ConfirmarEmail string `json:"confirmar_email" binding:"required,email"`
	Senha          string `json:"senha" binding:"required"`
}

type AtualizarTelefoneRequest struct {
	TelefoneAtual     string `json:"telefone_atual" binding:"required"`
	NovoTelefone      string `json:"novo_telefone" binding:"required"`
	ConfirmarTelefone string `json:"confirmar_telefone" binding:"required"`
	Senha             string `json:"senha" binding:"required"`
}
