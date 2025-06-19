package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"bytebros.ti/models"

	"github.com/gin-gonic/gin"
)

func CriarNoticia(c *gin.Context) {
	log.Println("DEBUG: Iniciando CriarNoticia handler.")
	var noticiaReq models.NoticiaRequest
	if err := c.ShouldBindJSON(&noticiaReq); err != nil {
		log.Printf("DEBUG: Erro no ShouldBindJSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}
	log.Printf("DEBUG: Request bindada. Noticia: %+v", noticiaReq)

	db := c.MustGet("db").(*sql.DB)
	log.Println("DEBUG: Obtive a conexão com o DB.")

	noticia := models.Noticia{
		Titulo:    noticiaReq.Titulo,
		Subtitulo: noticiaReq.Subtitulo,
		Conteudo:  noticiaReq.Conteudo,
		Autor:     noticiaReq.Autor,
		Data:      time.Now(),
	}

	err := db.QueryRow(`
        INSERT INTO noticias (titulo, subtitulo, conteudo, autor, data)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`,
		noticia.Titulo, noticia.Subtitulo, noticia.Conteudo, noticia.Autor, noticia.Data).
		Scan(&noticia.ID)

	if err != nil {
		log.Printf("DEBUG: Erro ao inserir notícia no DB: %v", err) // Log detalhado do erro real
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar notícia"})
		return
	}
	log.Printf("DEBUG: Notícia criada com ID: %d", noticia.ID)
	c.JSON(http.StatusCreated, noticia)
	log.Println("DEBUG: Resposta CriarNoticia enviada.")
}

func ListarNoticias(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	rows, err := db.Query(`
        SELECT id, titulo, subtitulo, conteudo, autor, data
        FROM noticias
        ORDER BY data DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar notícias"})
		return
	}
	defer rows.Close()

	var noticias []models.Noticia
	for rows.Next() {
		var n models.Noticia
		if err := rows.Scan(&n.ID, &n.Titulo, &n.Subtitulo, &n.Conteudo, &n.Autor, &n.Data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler notícias"})
			return
		}
		noticias = append(noticias, n)
	}

	c.JSON(http.StatusOK, noticias)
}

func ObterNoticia(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	var noticia models.Noticia
	err := db.QueryRow(`
        SELECT id, titulo, subtitulo, conteudo, autor, data
        FROM noticias
        WHERE id = $1`, id).
		Scan(&noticia.ID, &noticia.Titulo, &noticia.Subtitulo, &noticia.Conteudo, &noticia.Autor, &noticia.Data)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Notícia não encontrada"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar notícia"})
		}
		return
	}

	c.JSON(http.StatusOK, noticia)
}

func AtualizarNoticia(c *gin.Context) {
	id := c.Param("id")
	var noticiaReq models.NoticiaRequest

	if err := c.ShouldBindJSON(&noticiaReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	_, err := db.Exec(`
        UPDATE noticias
        SET titulo = $1, subtitulo = $2, conteudo = $3, autor = $4
        WHERE id = $5`,
		noticiaReq.Titulo, noticiaReq.Subtitulo, noticiaReq.Conteudo, noticiaReq.Autor, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar notícia"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Notícia atualizada com sucesso"})
}

func DeletarNoticia(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*sql.DB)

	_, err := db.Exec("DELETE FROM noticias WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar notícia"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Notícia deletada com sucesso"})
}
