package handlers

import (
	"database/sql"
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

	produto := models.Produto{
		Nome:       produtoReq.Nome,
		Quantidade: produtoReq.Quantidade,
		Preco:      produtoReq.Preco,
		Oferta:     produtoReq.Oferta,
	}

	err := db.QueryRow(`
        INSERT INTO produtos (nome, quantidade, preco, oferta)
        VALUES ($1, $2, $3, $4)
        RETURNING id`,
		produto.Nome, produto.Quantidade, produto.Preco, produto.Oferta).
		Scan(&produto.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar produto"})
		return
	}

	c.JSON(http.StatusCreated, produto)
}

func ListarProdutos(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	somenteOfertas := c.Query("ofertas") == "true"

	var rows *sql.Rows
	var err error

	if somenteOfertas {
		rows, err = db.Query(`
            SELECT id, nome, quantidade, preco, oferta
            FROM produtos
            WHERE oferta = true
            ORDER BY nome`)
	} else {
		rows, err = db.Query(`
            SELECT id, nome, quantidade, preco, oferta
            FROM produtos
            ORDER BY nome`)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar produtos"})
		return
	}
	defer rows.Close()

	var produtos []models.Produto
	for rows.Next() {
		var p models.Produto
		if err := rows.Scan(&p.ID, &p.Nome, &p.Quantidade, &p.Preco, &p.Oferta); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao ler produtos"})
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
        SELECT id, nome, quantidade, preco, oferta
        FROM produtos
        WHERE id = $1`, id).
		Scan(&produto.ID, &produto.Nome, &produto.Quantidade, &produto.Preco, &produto.Oferta)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Produto não encontrado"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar produto"})
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

	_, err := db.Exec(`
        UPDATE produtos
        SET nome = $1, quantidade = $2, preco = $3, oferta = $4
        WHERE id = $5`,
		produtoReq.Nome, produtoReq.Quantidade, produtoReq.Preco, produtoReq.Oferta, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar produto"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Produto atualizado com sucesso"})
}

func DeletarProduto(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*sql.DB)

	_, err := db.Exec("DELETE FROM produtos WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar produto"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Produto deletado com sucesso"})
}
