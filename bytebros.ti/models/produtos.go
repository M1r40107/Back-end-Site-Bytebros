package models

type Produto struct {
	ID         int     `json:"id"`
	Nome       string  `json:"nome"`
	Quantidade int     `json:"quantidade"`
	Preco      float64 `json:"preco"`
	Oferta     bool    `json:"oferta"`
}

type ProdutoRequest struct {
	Nome       string  `json:"nome"`
	Quantidade int     `json:"quantidade"`
	Preco      float64 `json:"preco"`
	Oferta     bool    `json:"oferta"`
}
