package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"bytebros.ti/models"
	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var geminiClient *genai.GenerativeModel

func InitializeGeminiClient() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("GEMINI_API_KEY não definida. Chatbot com IA não funcionará.")
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Erro ao criar cliente Gemini: %v", err)
	}
	geminiClient = client.GenerativeModel("gemini-1.5-flash")

	log.Println("Cliente Gemini inicializado com sucesso.")
}

func ChatbotHandler(c *gin.Context) {
	if geminiClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Chatbot IA não inicializado. Verifique a configuração da chave de API."})
		return
	}

	var chatReq struct {
		Message string `json:"message"`
		History []struct {
			Role  string   `json:"role"`
			Parts []string `json:"parts"`
		} `json:"history"`
	}

	if err := c.ShouldBindJSON(&chatReq); err != nil {
		log.Printf("ERRO: Falha ao fazer bind JSON para ChatbotHandler: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	cs := geminiClient.StartChat()

	for _, msg := range chatReq.History {
		var parts []genai.Part
		for _, partStr := range msg.Parts {
			parts = append(parts, genai.Text(partStr))
		}
		cs.History = append(cs.History, &genai.Content{
			Role:  msg.Role,
			Parts: parts,
		})
	}

	log.Printf("DEBUG: Enviando mensagem para Gemini. Última mensagem: '%s'", chatReq.Message)
	resp, err := cs.SendMessage(c.Request.Context(), genai.Text(chatReq.Message))
	if err != nil {
		log.Printf("ERRO Gemini SendMessage: %v", err)
		if strings.Contains(err.Error(), "blocked") || strings.Contains(err.Error(), "safety") {
			c.JSON(http.StatusOK, gin.H{"response": "Desculpe, sua mensagem contém conteúdo que não posso processar de acordo com minhas diretrizes de segurança. Por favor, reformule."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Desculpe, não consegui processar sua solicitação no momento. Tente novamente mais tarde."})
		return
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		botResponse := resp.Candidates[0].Content.Parts[0]
		if text, ok := botResponse.(genai.Text); ok {
			log.Printf("DEBUG: Resposta do chatbot: '%s'", string(text))
			c.JSON(http.StatusOK, gin.H{"response": string(text)})
			return
		}
	}

	log.Println("DEBUG: Nenhuma resposta válida do Gemini, enviando fallback.")
	c.JSON(http.StatusOK, gin.H{"response": "Desculpe, não consegui gerar uma resposta significativa."})
}

func ChatbotSupportRequest(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	clienteEmail, exists := c.Get("email")
	clienteEmailStr := ""
	if exists && clienteEmail != nil {
		clienteEmailStr = clienteEmail.(string)
	}

	var supportReq struct {
		Nome     string `json:"nome"`
		Email    string `json:"email"`
		Mensagem string `json:"mensagem"`
	}

	if err := c.ShouldBindJSON(&supportReq); err != nil {
		log.Printf("ERRO: Falha ao fazer bind JSON para ChatbotSupportRequest: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	var suporte models.Suporte
	err := db.QueryRow(`
        INSERT INTO suporte (nome, email, mensagem, status, tipo_interacao, cliente_email)
        VALUES ($1, $2, $3, 'aberto', $4, $5)
        RETURNING id, criado_em`,
		supportReq.Nome, supportReq.Email, supportReq.Mensagem, "chatbot_suporte", sql.NullString{String: clienteEmailStr, Valid: clienteEmailStr != ""}).
		Scan(&suporte.ID, &suporte.CriadoEm)

	if err != nil {
		log.Printf("ERRO BD: Erro ao registrar pedido de suporte via chatbot: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao registrar pedido de suporte via chatbot", "detalhes": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"mensagem": "Pedido de suporte via chatbot enviado com sucesso!", "id": suporte.ID})
}
