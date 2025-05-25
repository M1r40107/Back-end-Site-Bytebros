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
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado - usando variáveis de ambiente do sistema")
	}

	database.InitDB()
	defer database.CloseDB()

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Set("db", database.DB)
		c.Next()
	})

	noticiaRoutes := router.Group("/api/noticias")
	{
		noticiaRoutes.POST("/", handlers.CriarNoticia)
		noticiaRoutes.GET("/", handlers.ListarNoticias)
		noticiaRoutes.GET("/:id", handlers.ObterNoticia)
		noticiaRoutes.PUT("/:id", handlers.AtualizarNoticia)
		noticiaRoutes.DELETE("/:id", handlers.DeletarNoticia)
	}

	produtoRoutes := router.Group("/api/produtos")
	{
		produtoRoutes.POST("/", handlers.CriarProduto)
		produtoRoutes.GET("/", handlers.ListarProdutos)
		produtoRoutes.GET("/:id", handlers.ObterProduto)
		produtoRoutes.PUT("/:id", handlers.AtualizarProduto)
		produtoRoutes.DELETE("/:id", handlers.DeletarProduto)
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

		adminRoutes := protected.Group("/admin")
		adminRoutes.Use(handlers.AdminMiddleware())
		{
			adminRoutes.GET("/funcionarios", handlers.ListarFuncionarios)
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
		suporteRoutes.POST("/", handlers.CriarMensagemSuporte)

		adminSuporte := suporteRoutes.Group("/")
		adminSuporte.Use(handlers.AuthMiddleware(), handlers.AdminMiddleware())
		{
			adminSuporte.GET("/", handlers.ListarMensagensSuporte)
			adminSuporte.GET("/:id", handlers.ObterMensagemSuporte)
			adminSuporte.PUT("/:id/status", handlers.AtualizarStatusSuporte)
		}
	}

	adminRoutes := router.Group("/api/admin")
	adminRoutes.Use(handlers.AuthMiddleware(), handlers.AdminMiddleware())
	{
		adminRoutes.POST("/administradores", handlers.CriarAdministrador)
		adminRoutes.GET("/dashboard", handlers.AdminDashboard)
	}

	router.POST("/api/admin/login", handlers.LoginAdmin)
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
