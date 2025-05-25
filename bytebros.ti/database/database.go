package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão com o banco: %v", err)
	}

	DB.SetMaxIdleConns(5)
	DB.SetMaxOpenConns(10)
	DB.SetConnMaxLifetime(30 * time.Minute)
	DB.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := DB.PingContext(ctx); err != nil {
		log.Fatalf("Erro ao conectar ao PostgreSQL: %v", err)
	}

	log.Println("Conectado ao PostgreSQL com sucesso!")
}

func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Erro ao fechar conexão com o banco: %v", err)
		} else {
			log.Println("Conexão com o PostgreSQL fechada com sucesso")
		}
	}
}

func GetDB() *sql.DB {
	return DB
}

func CreateTables() error {
	tables := []struct {
		name  string
		query string
	}{
		{
			name: "usuarios",
			query: `
			CREATE TABLE IF NOT EXISTS usuarios (
				id SERIAL PRIMARY KEY,
				nome_completo VARCHAR(100) NOT NULL,
				email VARCHAR(100) NOT NULL UNIQUE,
				senha_hash VARCHAR(100) NOT NULL,
				criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				atualizado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_usuarios_email ON usuarios(email);`,
		},
		{
			name: "funcionarios",
			query: `
			CREATE TABLE IF NOT EXISTS funcionarios (
				id SERIAL PRIMARY KEY,
				nome VARCHAR(100) NOT NULL,
				cargo VARCHAR(50) NOT NULL,
				email VARCHAR(100) NOT NULL UNIQUE,
				senha_hash VARCHAR(100) NOT NULL,
				criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_funcionarios_email ON funcionarios(email);
			CREATE INDEX IF NOT EXISTS idx_funcionarios_cargo ON funcionarios(cargo);`,
		},
		{
			name: "produtos",
			query: `
			CREATE TABLE IF NOT EXISTS produtos (
				id SERIAL PRIMARY KEY,
				nome VARCHAR(100) NOT NULL,
				quantidade INTEGER NOT NULL DEFAULT 0,
				preco DECIMAL(10,2) NOT NULL,
				oferta BOOLEAN NOT NULL DEFAULT false
			);
			CREATE INDEX IF NOT EXISTS idx_produtos_oferta ON produtos(oferta);
			CREATE INDEX IF NOT EXISTS idx_produtos_nome ON produtos(nome);`,
		},
		{
			name: "noticias",
			query: `
			CREATE TABLE IF NOT EXISTS noticias (
				id SERIAL PRIMARY KEY,
				titulo VARCHAR(150) NOT NULL,
				subtitulo VARCHAR(300) NOT NULL,
				conteudo TEXT NOT NULL,
				autor VARCHAR(100) NOT NULL,
				data TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_noticias_data ON noticias(data);`,
		},
		{
			name: "servicos",
			query: `
			CREATE TABLE IF NOT EXISTS servicos (
				id SERIAL PRIMARY KEY,
				nome VARCHAR(100) NOT NULL,
				preco DECIMAL(10,2) NOT NULL,
				oferta BOOLEAN NOT NULL DEFAULT false,
				detalhes TEXT NOT NULL
			);
			CREATE INDEX IF NOT EXISTS idx_servicos_oferta ON servicos(oferta);
			CREATE INDEX IF NOT EXISTS idx_servicos_nome ON servicos(nome);`,
		},
		{
			name: "suporte",
			query: `
			CREATE TABLE IF NOT EXISTS suporte (
        		id SERIAL PRIMARY KEY,
        		nome VARCHAR(100) NOT NULL,
        		email VARCHAR(100) NOT NULL,
        		mensagem TEXT NOT NULL,
        		status VARCHAR(20) NOT NULL DEFAULT 'aberto',
        		criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
    		CREATE INDEX IF NOT EXISTS idx_suporte_status ON suporte(status);
    		CREATE INDEX IF NOT EXISTS idx_suporte_email ON suporte(email);`,
		},
	}

	for _, table := range tables {
		log.Printf("Criando tabela %s...", table.name)
		if _, err := DB.Exec(table.query); err != nil {
			return fmt.Errorf("erro ao criar tabela %s: %w", table.name, err)
		}
	}

	log.Println("Todas as tabelas foram criadas com sucesso")
	return nil
}

func DropTables() error {
	tables := []string{
		"servicos",
		"noticias",
		"produtos",
		"funcionarios",
		"usuarios",
		"suporte",
	}

	for _, table := range tables {
		if _, err := DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)); err != nil {
			return fmt.Errorf("erro ao dropar tabela %s: %w", table, err)
		}
		log.Printf("Tabela %s removida", table)
	}

	return nil
}
