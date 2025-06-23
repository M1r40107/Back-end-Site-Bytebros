package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bytebros.ti/handlers"

	"bytebros.ti/database"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado - usando variáveis de ambiente do sistema")
	}

	database.InitDB()
	defer database.CloseDB()

	if err := database.CreateTables(); err != nil {
		log.Fatalf("Erro ao criar tabelas: %v", err)
	}

	handlers.InitializeGeminiClient()
	log.SetOutput(os.Stderr)

	router := gin.Default()

	router.RedirectTrailingSlash = false

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"https://bytebros.netlify.app/"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour
	config.AddAllowHeaders("x-requested-with")
	router.Use(cors.New(config))

	router.Use(func(c *gin.Context) {
		c.Set("db", database.DB)
		c.Next()
	})

	router.GET("/api/noticias", handlers.ListarNoticias)
	router.GET("/api/noticias/:id", handlers.ObterNoticia)

	produtoRoutes := router.Group("/api/produtos")
	{
		produtoRoutes.POST("", handlers.CriarProduto)
		produtoRoutes.GET("", handlers.ListarProdutos)
		produtoRoutes.GET("/:id", handlers.ObterProduto)
		produtoRoutes.PUT("/:id", handlers.AtualizarProduto)
		produtoRoutes.DELETE("/:id", handlers.DeletarProduto)
	}

	orcamentoRoutes := router.Group("/api/orcamentos")
	{
		orcamentoRoutes.POST("", handlers.CriarOrcamento)
	}

	authRoutes := router.Group("/api/auth")
	{
		authRoutes.POST("/registrar", handlers.RegistrarUsuario)
		authRoutes.POST("/login", handlers.LoginUsuario)

		authRoutes.POST("/funcionarios/registrar", handlers.RegistrarFuncionario)
		authRoutes.POST("/funcionarios/login", handlers.LoginFuncionario)
	}

	protected := router.Group("/api")
	protected.Use(handlers.AuthMiddleware())
	{
		protected.GET("/perfil", handlers.ObterPerfil)
		protected.POST("/pedidos", handlers.CriarPedido)
		protected.GET("/meus-pedidos", handlers.ListarPedidosCliente)
		protected.GET("/minhas-interacoes", handlers.ListarInteracoesCliente)
		protected.POST("/chatbot/suporte", handlers.ChatbotSupportRequest)
		protected.PUT("/usuarios/email", handlers.AtualizarEmailUsuario)
		protected.PUT("/usuarios/telefone", handlers.AtualizarTelefoneUsuario)

		adminRoutes := protected.Group("/admin")
		adminRoutes.Use(handlers.AdminMiddleware())
		{
			adminRoutes.DELETE("/administradores/:id", handlers.DeletarAdministrador)
			adminRoutes.GET("/funcionarios", handlers.ListarFuncionarios)
			adminRoutes.GET("/usuarios", handlers.ListarUsuarios)
			adminRoutes.GET("/pedidos", handlers.ListarPedidosAdmin)
			adminRoutes.PUT("/pedidos/:id/status", handlers.AtualizarStatusPedido)
			adminRoutes.DELETE("/pedidos/:id", handlers.DeletarPedido)

			adminRoutes.POST("/noticias", handlers.CriarNoticia)
			adminRoutes.PUT("/noticias/:id", handlers.AtualizarNoticia)
			adminRoutes.DELETE("/noticias/:id", handlers.DeletarNoticia)

			adminRoutes.GET("/orcamentos", handlers.ListarOrcamentos)
			adminRoutes.GET("/orcamentos/:id", handlers.ObterOrcamento)
			adminRoutes.PUT("/orcamentos/:id/status", handlers.AtualizarStatusOrcamento)
			adminRoutes.DELETE("/orcamentos/:id", handlers.DeletarOrcamento)
		}
	}

	servicosRoutes := router.Group("/api/servicos")
	{
		servicosRoutes.GET("/", handlers.ListarServicos)
		servicosRoutes.GET("/:id", handlers.ObterServico)

		adminServicos := servicosRoutes.Group("/")
		adminServicos.Use(handlers.AuthMiddleware(), handlers.AdminMiddleware())
		{
			adminServicos.POST("/", handlers.CriarServico)
			adminServicos.PUT("/:id", handlers.AtualizarServico)
			adminServicos.DELETE("/:id", handlers.DeletarServico)
		}
	}

	suporteRoutes := router.Group("/api/suporte")
	{
		suporteRoutes.POST("", handlers.CriarMensagemSuporte)

		adminSuporte := suporteRoutes.Group("")
		adminSuporte.Use(handlers.AuthMiddleware(), handlers.AdminMiddleware())
		{
			adminSuporte.GET("", handlers.ListarMensagensSuporte)
			adminSuporte.GET("/:id", handlers.ObterMensagemSuporte)
			adminSuporte.PUT("/:id/status", handlers.AtualizarStatusSuporte)
			adminSuporte.DELETE("/:id", handlers.DeletarSuporte)
		}
	}

	adminRoutes := router.Group("/api/admin")
	adminRoutes.Use(handlers.AuthMiddleware(), handlers.AdminMiddleware())
	{
		adminRoutes.POST("/administradores", handlers.CriarAdministrador)
		adminRoutes.GET("/dashboard", handlers.AdminDashboard)
	}

	router.POST("/api/admin/login", handlers.LoginAdmin)

	router.POST("/api/chatbot", handlers.ChatbotHandler)

	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Servidor iniciado na porta %s", os.Getenv("PORT"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Falha ao iniciar servidor: %v", err)
		}
	}()

	<-quit
	log.Println("Desligando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Falha ao desligar servidor: %v", err)
	}

	log.Println("Servidor desligado com sucesso")
}
