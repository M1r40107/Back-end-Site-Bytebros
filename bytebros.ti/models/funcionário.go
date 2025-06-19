package models

type Funcionario struct {
	ID    int    `json:"id"`
	Nome  string `json:"nome" binding:"required,min=3"`
	Cargo string `json:"cargo" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required,min=6"`
}

type FuncionarioRequest struct {
	Nome  string `json:"nome"`
	Cargo string `json:"cargo"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type FuncionarioLogin struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required,min=6"`
}

type FuncionarioResponse struct {
	ID    int    `json:"id"`
	Nome  string `json:"nome"`
	Cargo string `json:"cargo"`
	Email string `json:"email"`
	Token string `json:"token,omitempty"`
}
