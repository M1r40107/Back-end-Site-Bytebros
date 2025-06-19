package models

import "database/sql"

type Produto struct {
	ID         int            `json:"id"`
	Nome       string         `json:"name"`
	Quantidade int            `json:"quantity"`
	Preco      float64        `json:"value"`
	Oferta     bool           `json:"oferta"`
	Detalhes   sql.NullString `json:"details"`
	Imagem     sql.NullString `json:"image"`
}

type ProdutoRequest struct {
	Nome       string  `json:"name" binding:"required"`
	Quantidade int     `json:"quantity" binding:"required,min=0"`
	Preco      float64 `json:"value" binding:"required,min=0.01"`
	Oferta     bool    `json:"oferta"`
	Detalhes   string  `json:"details"`
	Imagem     string  `json:"image"`
}
