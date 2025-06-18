package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"bytebros.ti/models"

	"github.com/gin-gonic/gin"
)

func CriarProduto(c *gin.Context) {
	var produtoReq models.ProdutoRequest

	if err := c.ShouldBindJSON(&produtoReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	var produto models.Produto
	detalhesNull := sql.NullString{String: produtoReq.Detalhes, Valid: produtoReq.Detalhes != ""}
	imagemNull := sql.NullString{String: produtoReq.Imagem, Valid: produtoReq.Imagem != ""}

	err := db.QueryRow(`
        INSERT INTO produtos (nome, quantidade, preco, oferta, detalhes, imagem)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id`,
		produtoReq.Nome, produtoReq.Quantidade, produtoReq.Preco, produtoReq.Oferta, detalhesNull, imagemNull).
		Scan(&produto.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar produto", "detalhes": err.Error()})
		return
	}

	produto.Nome = produtoReq.Nome
	produto.Quantidade = produtoReq.Quantidade
	produto.Preco = produtoReq.Preco
	produto.Oferta = produtoReq.Oferta
	produto.Detalhes.String = produtoReq.Detalhes
	produto.Detalhes.Valid = (produtoReq.Detalhes != "")
	produto.Imagem.String = produtoReq.Imagem
	produto.Imagem.Valid = (produtoReq.Imagem != "")

	c.JSON(http.StatusCreated, produto)
}

func ListarProdutos(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	somenteOfertas := c.Query("ofertas") == "true"

	var rows *sql.Rows
	var err error

	query := `SELECT id, nome, quantidade, preco, oferta, detalhes, imagem FROM produtos `
	if somenteOfertas {
		query += `WHERE oferta = true ORDER BY nome`
	} else {
		query += `ORDER BY nome`
	}

	rows, err = db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar produtos", "detalhes": err.Error()})
		return
	}
	defer rows.Close()

	var produtos []models.Produto
	for rows.Next() {
		var p models.Produto
		if err := rows.Scan(&p.ID, &p.Nome, &p.Quantidade, &p.Preco, &p.Oferta, &p.Detalhes, &p.Imagem); err != nil {
			log.Printf("ERRO BD: Erro ao ler produto durante Scan: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler produtos", "detalhes": err.Error()})
			return
		}
		produtos = append(produtos, p)
	}
	c.JSON(http.StatusOK, produtos)
}

func ObterProduto(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	var produto models.Produto
	err := db.QueryRow(`
        SELECT id, nome, quantidade, preco, oferta, detalhes, imagem
        FROM produtos
        WHERE id = $1`, id).
		Scan(&produto.ID, &produto.Nome, &produto.Quantidade, &produto.Preco, &produto.Oferta, &produto.Detalhes, &produto.Imagem)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Produto não encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar produto", "detalhes": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, produto)
}

func AtualizarProduto(c *gin.Context) {
	id := c.Param("id")
	var produtoReq models.ProdutoRequest

	if err := c.ShouldBindJSON(&produtoReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	detalhesNull := sql.NullString{String: produtoReq.Detalhes, Valid: produtoReq.Detalhes != ""}
	imagemNull := sql.NullString{String: produtoReq.Imagem, Valid: produtoReq.Imagem != ""}

	_, err := db.Exec(`
        UPDATE produtos
        SET nome = $1, quantidade = $2, preco = $3, oferta = $4, detalhes = $5, imagem = $6
        WHERE id = $7`,
		produtoReq.Nome, produtoReq.Quantidade, produtoReq.Preco, produtoReq.Oferta, detalhesNull, imagemNull, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar produto", "detalhes": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mensagem": "Produto atualizado com sucesso"})
}

func DeletarProduto(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*sql.DB)

	_, err := db.Exec("DELETE FROM produtos WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar produto", "detalhes": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mensagem": "Produto deletado com sucesso"})
}
