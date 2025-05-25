package models

type Servico struct {
	ID       int     `json:"id"`
	Nome     string  `json:"nome" binding:"required,min=3"`
	Preco    float64 `json:"preco" binding:"required,min=0.01"`
	Oferta   bool    `json:"oferta"`
	Detalhes string  `json:"detalhes" binding:"required,min=10"`
}

type ServicoRequest struct {
	Nome     string  `json:"nome"`
	Preco    float64 `json:"preco"`
	Oferta   bool    `json:"oferta"`
	Detalhes string  `json:"detalhes"`
}
