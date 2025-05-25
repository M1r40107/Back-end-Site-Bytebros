package models

import "time"

type Noticia struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Titulo    string    `json:"titulo"`
	Subtitulo string    `json:"subtitulo"`
	Conteudo  string    `json:"conteudo"`
	Autor     string    `json:"autor"`
	Data      time.Time `json:"data"`
}

type NoticiaRequest struct {
	Titulo    string `json:"titulo"`
	Subtitulo string `json:"subtitulo"`
	Conteudo  string `json:"conteudo"`
	Autor     string `json:"autor"`
}
