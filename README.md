# Documentação Técnica do Sistema Byte Bros.TI

**Localização:** Toda a documentação técnica está disponível no repositório GitHub do projeto na pasta `/docs`.

## 1\. Arquitetura do Sistema

O sistema Byte Bros.TI é uma aplicação web de três camadas (frontend, backend, banco de dados) seguindo um modelo de arquitetura de microsserviços para o backend, o que permite maior escalabilidade e flexibilidade no desenvolvimento e deploy.

### 1.1. Diagrama de Componentes

```mermaid
graph TD
    subgraph Frontend (Cliente)
        A[Navegador Web] --> B[HTML/CSS/JS]
    end

    subgraph Backend (Servidor Go)
        C[API Gateway/Load Balancer] --> D[Módulo de Autenticação]
        C --> E[Módulo de Produtos]
        C --> F[Módulo de Pedidos]
        C --> G[Módulo de Suporte]
        C --> H[Módulo de Orçamentos]
        C --> I[Módulo de Notícias]
        C --> J[Módulo de Serviços]
        C --> K[Módulo de Chatbot AI]
        C --> L[Módulo Admin]
    end

    subgraph Banco de Dados (PostgreSQL)
        M[Servidor PostgreSQL]
    end

    B -- Requisições HTTP/REST --> C
    D -- Autenticação/Dados --> M
    E -- Dados de Produtos --> M
    F -- Dados de Pedidos/Itens --> M
    G -- Dados de Suporte --> M
    H -- Dados de Orçamentos --> M
    I -- Dados de Notícias --> M
    J -- Dados de Serviços --> M
    K -- AI/Suporte --> M
    L -- Gerenciamento Admin --> M

    K -- Integração API --> N[Google Gemini AI]
```

### 1.2. Tecnologias Utilizadas

  * **Frontend:**
      * **HTML5:** Estrutura das páginas.
      * **CSS3:** Estilização (incluindo variáveis CSS para temas e media queries para responsividade).
      * **JavaScript (ES6+):** Lógica interativa, consumo de API, manipulação de DOM, gerenciamento de estado local (LocalStorage). Utiliza módulos JS.
      * **Font Awesome:** Biblioteca de ícones.
  * **Backend:**
      * **Go (Golang):** Linguagem de programação principal para a lógica de negócios e API RESTful.
      * **Gin Gonic:** Framework web para Go, utilizado para roteamento, middlewares e tratamento de requisições HTTP.
      * **`github.com/lib/pq`:** Driver PostgreSQL para Go.
      * **`golang.org/x/crypto/bcrypt`:** Para criptografia de senhas.
      * **`github.com/golang-jwt/jwt/v4`:** Para geração e validação de JSON Web Tokens (JWT).
      * **`github.com/gin-contrib/cors`:** Middleware para Cross-Origin Resource Sharing (CORS).
      * **`github.com/joho/godotenv`:** Para carregamento de variáveis de ambiente de arquivos `.env` localmente.
      * **`github.com/google/generative-ai-go/genai`:** SDK oficial do Google para integração com a API Gemini AI.
  * **Banco de Dados:**
      * **PostgreSQL:** Sistema de gerenciamento de banco de dados relacional.
  * **Hospedagem (Planejado):**
      * **Frontend:** Netlify
      * **Backend:** Render.com (ou Heroku/Google Cloud Run/AWS Elastic Beanstalk)
      * **Banco de Dados:** Supabase

## 2\. APIs REST

As APIs REST são o coração da comunicação entre o frontend e o backend. Os endpoints seguem um padrão RESTful e são protegidos por JWT onde a autenticação é necessária.

**Base URL:** `http://localhost:8080/api` (para desenvolvimento local)

### 2.1. Autenticação (`/api/auth`)

  * **`POST /auth/registrar`**

      * **Descrição:** Registra um novo usuário no sistema.
      * **Parâmetros (Body - JSON):**
        ```json
        {
          "nome": "Nome Completo do Usuário",
          "email": "usuario@example.com",
          "telefone": "999999999",
          "senha": "senhaSegura123"
        }
        ```
      * **Respostas:**
          * `201 Created`: `{"id": 1, "nome": "Nome Completo", "email": "usuario@example.com", "token": "jwt_token", "telefone": "999999999"}`
          * `400 Bad Request`: `{ "erro": "Mensagem de erro de validação ou email já registrado" }`
          * `500 Internal Server Error`: `{ "erro": "Erro interno do servidor" }`

  * **`POST /auth/login`**

      * **Descrição:** Autentica um usuário e retorna um token JWT.
      * **Parâmetros (Body - JSON):**
        ```json
        {
          "email": "usuario@example.com",
          "senha": "senhaSegura123"
        }
        ```
      * **Respostas:**
          * `200 OK`: `{"id": 1, "nome": "Nome Completo", "email": "usuario@example.com", "token": "jwt_token", "telefone": "999999999"}`
          * `401 Unauthorized`: `{ "erro": "Credenciais inválidas" }`
          * `400 Bad Request`: `{ "erro": "Mensagem de erro de validação" }`
          * `500 Internal Server Error`: `{ "erro": "Erro interno do servidor" }`

  * **`PUT /usuarios/email`** (Protegida)

      * **Descrição:** Permite que um usuário logado altere seu e-mail.
      * **Auth:** `Authorization: Bearer <user_token>`
      * **Parâmetros (Body - JSON):**
        ```json
        {
          "email_atual": "emailantigo@example.com",
          "novo_email": "novoemail@example.com",
          "confirmar_email": "novoemail@example.com",
          "senha": "senhaAtual123"
        }
        ```
      * **Respostas:**
          * `200 OK`: `{"mensagem": "Email atualizado com sucesso!", "novo_email": "novoemail@example.com", "token": "novo_jwt_token"}`
          * `400 Bad Request`: `{ "erro": "Validação falha" }`
          * `401 Unauthorized`: `{ "erro": "Token inválido/ausente" }`
          * `403 Forbidden`: `{ "erro": "Email atual incorreto / Senha incorreta" }`
          * `409 Conflict`: `{ "erro": "Este novo email já está em uso" }`
          * `500 Internal Server Error`: `{ "erro": "Erro interno do servidor" }`

  * **`PUT /usuarios/telefone`** (Protegida)

      * **Descrição:** Permite que um usuário logado altere seu telefone.
      * **Auth:** `Authorization: Bearer <user_token>`
      * **Parâmetros (Body - JSON):**
        ```json
        {
          "telefone_atual": "999999999",
          "novo_telefone": "888888888",
          "confirmar_telefone": "888888888",
          "senha": "senhaAtual123"
        }
        ```
      * **Respostas:**
          * `200 OK`: `{"mensagem": "Telefone atualizado com sucesso!", "novo_telefone": "888888888"}`
          * `400 Bad Request`: `{ "erro": "Validação falha" }`
          * `401 Unauthorized`: `{ "erro": "Token inválido/ausente" }`
          * `403 Forbidden`: `{ "erro": "Telefone atual incorreto / Senha incorreta" }`
          * `500 Internal Server Error`: `{ "erro": "Erro interno do servidor" }`

### 2.2. Produtos (`/api/produtos`)

  * **`GET /produtos`**

      * **Descrição:** Lista todos os produtos disponíveis. Pode ser filtrado por produtos em oferta.
      * **Parâmetros (Query):** `?ofertas=true` (opcional, para listar apenas produtos em oferta).
      * **Respostas:** `200 OK`: `[ { "id": 1, "name": "Produto X", "quantity": 10, "value": 150.00, "oferta": false, "details": "Detalhes do produto X", "image": "url_imagem.jpg" } ]`

  * **`GET /produtos/{id}`**

      * **Descrição:** Obtém detalhes de um produto específico.
      * **Parâmetros (Path):** `id` (ID do produto).
      * **Respostas:** `200 OK` (objeto Produto), `404 Not Found` (produto não encontrado).

  * **`POST /produtos`** (Protegida - Admin)

      * **Descrição:** Adiciona um novo produto.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Body - JSON):** `{"name": "Novo Produto", "quantity": 5, "value": 200.00, "oferta": false, "details": "Detalhes do novo produto.", "image": "url_da_imagem.jpg"}`
      * **Respostas:** `201 Created` (objeto Produto criado), `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`.

  * **`PUT /produtos/{id}`** (Protegida - Admin)

      * **Descrição:** Atualiza um produto existente.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id` (ID do produto). **Parâmetros (Body - JSON):** Objeto Produto com campos a serem atualizados.
      * **Respostas:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

  * **`DELETE /produtos/{id}`** (Protegida - Admin)

      * **Descrição:** Exclui um produto.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id` (ID do produto).
      * **Respostas:** `200 OK`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

### 2.3. Notícias (`/api/noticias`)

  * **`GET /noticias`**

      * **Descrição:** Lista todas as notícias.
      * **Respostas:** `200 OK`: `[ { "id": 1, "titulo": "Título", "subtitulo": "Sub", "conteudo": "...", "autor": "Autor", "data": "2025-01-01T10:00:00Z" } ]`

  * **`GET /noticias/{id}`**

      * **Descrição:** Obtém detalhes de uma notícia específica.
      * **Respostas:** `200 OK` (objeto Notícia), `404 Not Found`.

  * **`POST /admin/noticias`** (Protegida - Admin)

      * **Descrição:** Adiciona uma nova notícia.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Body - JSON):** `{"titulo": "Novo", "subtitulo": "Subtítulo", "conteudo": "...", "autor": "Admin"}`
      * **Respostas:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`.

  * **`PUT /admin/noticias/{id}`** (Protegida - Admin)

      * **Descrição:** Atualiza uma notícia.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id`. **Parâmetros (Body - JSON):** Objeto Notícia com campos a serem atualizados.
      * **Respostas:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

  * **`DELETE /admin/noticias/{id}`** (Protegida - Admin)

      * **Descrição:** Exclui uma notícia.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id`.
      * **Respostas:** `200 OK`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

### 2.4. Orçamentos (`/api/orcamentos`)

  * **`POST /orcamentos`**

      * **Descrição:** Cria uma nova solicitação de orçamento.
      * **Parâmetros (Body - JSON):** `{"nome_cliente": "Fulano", "email_cliente": "fulano@email.com", "telefone": "999999999", "descricao": "Quero orçamento para PC gamer", "servico_nome": "Montagem de Computadores"}`
      * **Respostas:** `201 Created`, `400 Bad Request`, `500 Internal Server Error`.

  * **`GET /admin/orcamentos`** (Protegida - Admin)

      * **Descrição:** Lista todas as solicitações de orçamento. Pode ser filtrado.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Query):** `?status=pendente` (opcional), `?email=cliente@email.com` (opcional).
      * **Respostas:** `200 OK`: `[ { "id": 1, "nome_cliente": "Fulano", "email_cliente": "...", "telefone": "...", "descricao": "...", "servico_nome": "...", "status": "pendente", "criado_em": "..." } ]`

  * **`PUT /admin/orcamentos/{id}/status`** (Protegida - Admin)

      * **Descrição:** Atualiza o status de uma solicitação de orçamento.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id`. **Parâmetros (Body - JSON):** `{"status": "aprovado"}`
      * **Status Permitidos:** `pendente`, `em_analise`, `aprovado`, `rejeitado`.
      * **Respostas:** `200 OK`, `400 Bad Request` (status inválido), `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

  * **`DELETE /admin/orcamentos/{id}`** (Protegida - Admin)

      * **Descrição:** Exclui uma solicitação de orçamento.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id`.
      * **Respostas:** `200 OK`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

### 2.5. Pedidos de Loja (`/api/pedidos`)

  * **`POST /pedidos`** (Protegida - Usuário Logado)

      * **Descrição:** Cria um novo pedido de loja com itens.
      * **Auth:** `Authorization: Bearer <user_token>`
      * **Parâmetros (Body - JSON):**
        ```json
        {
          "itens": [
            { "produto_id": 1, "nome_produto": "Core i9", "quantidade": 1, "valor_unitario": 449.90 }
          ],
          "endereco_entrega": "Rua X, 123 - Bairro Y",
          "tipo_frete": "padrao",
          "valor_frete": 25.00,
          "valor_total": 474.90,
          "forma_pagamento": "credito",
          "prazo_entrega": "25/06/2025"
        }
        ```
      * **Respostas:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`.

  * **`GET /meus-pedidos`** (Protegida - Usuário Logado)

      * **Descrição:** Lista os pedidos de loja do usuário logado.
      * **Auth:** `Authorization: Bearer <user_token>`
      * **Respostas:** `200 OK`: `[ { "id": 1, "cliente_email": "...", "data_pedido": "...", "status": "Processando", "itens": [{...}], "valor_total": 100.00 } ]`

  * **`GET /admin/pedidos`** (Protegida - Admin)

      * **Descrição:** Lista todos os pedidos de loja. Pode ser filtrado.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Query):** `?status=Processando` (opcional), `?cliente_email=cliente@email.com` (opcional).
      * **Respostas:** `200 OK` (array de objetos Pedido).

  * **`PUT /admin/pedidos/{id}/status`** (Protegida - Admin)

      * **Descrição:** Atualiza o status de um pedido de loja.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id`. **Parâmetros (Body - JSON):** `{"status": "Entregue"}`
      * **Respostas:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

  * **`DELETE /admin/pedidos/{id}`** (Protegida - Admin)

      * **Descrição:** Exclui um pedido de loja.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id`.
      * **Respostas:** `200 OK`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

### 2.6. Suporte (`/api/suporte`)

  * **`POST /suporte`** (Protegida - Usuário Logado ou Admin - para `cliente_email`)

      * **Descrição:** Cria uma nova mensagem de suporte. `tipo_interacao` será "suporte" (da página de atendimento) ou "chatbot\_suporte" (do chatbot).
      * **Auth:** `Authorization: Bearer <user_token>` (se quiser que `cliente_email` seja preenchido automaticamente).
      * **Parâmetros (Body - JSON):** `{"nome": "Fulano", "email": "fulano@email.com", "mensagem": "Problema com meu PC", "tipo_interacao": "suporte"}`
      * **Respostas:** `201 Created`, `400 Bad Request`, `500 Internal Server Error`.

  * **`GET /minhas-interacoes`** (Protegida - Usuário Logado)

      * **Descrição:** Lista interações de suporte/contato E orçamentos do usuário logado.
      * **Auth:** `Authorization: Bearer <user_token>`
      * **Respostas:** `200 OK`: `[ { "id": 1, "nome": "Fulano", "email": "...", "mensagem": "...", "status": "aberto", "tipo_interacao": "suporte", "criado_em": "..." }, { "id": 2, "nome": "Cicrano", "email": "...", "mensagem": "...", "status": "pendente", "tipo_interacao": "orcamento", "criado_em": "...", "servico_nome": "..." } ]`

  * **`GET /admin/suporte`** (Protegida - Admin)

      * **Descrição:** Lista todas as mensagens de suporte/contato. Pode ser filtrado.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Query):** `?status=aberto` (opcional), `?tipo_interacao=suporte` (opcional), `?cliente_email=cliente@email.com` (opcional).
      * **Respostas:** `200 OK` (array de objetos Suporte).

  * **`PUT /admin/suporte/{id}/status`** (Protegida - Admin)

      * **Descrição:** Atualiza o status de uma mensagem de suporte/contato.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id`. **Parâmetros (Body - JSON):** `{"status": "resolvido"}`
      * **Status Permitidos:** `aberto`, `em_andamento`, `resolvido`.
      * **Respostas:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

  * **`DELETE /admin/suporte/{id}`** (Protegida - Admin)

      * **Descrição:** Exclui uma mensagem de suporte/contato.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id`.
      * **Respostas:** `200 OK`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

### 2.7. Serviços (`/api/servicos`)

  * **`GET /servicos`**

      * **Descrição:** Lista todos os serviços. Pode ser filtrado por serviços em oferta.
      * **Parâmetros (Query):** `?ofertas=true` (opcional).
      * **Respostas:** `200 OK`: `[ { "id": 1, "nome": "Serviço X", "preco": 100.00, "oferta": false, "detalhes": "Detalhes do serviço X" } ]`

  * **`GET /servicos/{id}`**

      * **Descrição:** Obtém detalhes de um serviço específico.
      * **Respostas:** `200 OK` (objeto Serviço), `404 Not Found`.

  * **`POST /servicos`** (Protegida - Admin)

      * **Descrição:** Adiciona um novo serviço.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Body - JSON):** `{"nome": "Nova Limpeza", "preco": 80.00, "oferta": false, "detalhes": "Detalhes da limpeza."}`
      * **Respostas:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`.

  * **`PUT /servicos/{id}`** (Protegida - Admin)

      * **Descrição:** Atualiza um serviço.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id`. **Parâmetros (Body - JSON):** Objeto Serviço com campos a serem atualizados.
      * **Respostas:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

  * **`DELETE /servicos/{id}`** (Protegida - Admin)

      * **Descrição:** Exclui um serviço.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Path):** `id`.
      * **Respostas:** `200 OK`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`.

### 2.8. Admin (`/api/admin`)

  * **`POST /admin/administradores`** (Protegida - Super Admin)

      * **Descrição:** Adiciona um novo usuário administrador.
      * **Auth:** `Authorization: Bearer <super_admin_token>`
      * **Parâmetros (Body - JSON):** `{"nome": "Novo Admin", "email": "novo@admin.com", "senha": "senhaSeguraAdmin", "is_admin": true}`
      * **Respostas:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`.

  * **`DELETE /admin/administradores/{id}`** (Protegida - Super Admin)

      * **Descrição:** Exclui um usuário administrador (não o próprio super admin).
      * **Auth:** `Authorization: Bearer <super_admin_token>`
      * **Parâmetros (Path):** `id`.
      * **Respostas:** `200 OK`, `401 Unauthorized`, `403 Forbidden` (se tentar excluir super admin ou a si mesmo), `404 Not Found`.

  * **`POST /admin/login`**

      * **Descrição:** Autentica um administrador.
      * **Parâmetros (Body - JSON):** `{"email": "admin@example.com", "senha": "senhaAdmin123"}`
      * **Respostas:** `200 OK`: `{"id": 1, "nome": "Nome Admin", "email": "admin@example.com", "is_admin": true, "token": "jwt_token"}`
      * `401 Unauthorized`, `400 Bad Request`, `500 Internal Server Error`.

  * **`GET /admin/dashboard`** (Protegida - Admin)

      * **Descrição:** Retorna informações básicas do painel administrativo.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Respostas:** `200 OK`: `{"mensagem": "Bem-vindo ao painel administrativo", "usuario": "admin@example.com", "is_admin": true}`

### 2.9. Chatbot (`/api/chatbot`)

  * **`POST /chatbot`**

      * **Descrição:** Envia uma mensagem para o chatbot AI e recebe uma resposta.
      * **Parâmetros (Body - JSON):** `{"message": "Meu computador está lento.", "history": [...]}` (history é opcional, mas útil para contexto)
      * **Respostas:** `200 OK`: `{"response": "Resposta do chatbot."}`
      * `400 Bad Request`, `500 Internal Server Error`.

  * **`POST /chatbot/suporte`** (Pode ser protegido para pegar `cliente_email` do token)

      * **Descrição:** Cria um pedido de suporte através da interação do chatbot.
      * **Auth:** `Authorization: Bearer <user_token>` (opcional, mas recomendado para vincular ao usuário)
      * **Parâmetros (Body - JSON):** `{"nome": "Nome Usuário", "email": "email@usuario.com", "mensagem": "Problema detalhado pelo chatbot"}`
      * **Respostas:** `201 Created`, `400 Bad Request`, `500 Internal Server Error`.

### 2.10. Listagem de Clientes e Funcionários (Admin)

  * **`GET /admin/usuarios`** (Protegida - Admin)

      * **Descrição:** Lista usuários (clientes). Filtrável por email/telefone.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Parâmetros (Query):** `?busca=termo_de_busca` (email ou telefone).
      * **Respostas:** `200 OK`: `[ { "id": 1, "nome": "Fulano Cliente", "email": "cliente@email.com", "telefone": "999999999" } ]`

  * **`GET /admin/funcionarios`** (Protegida - Admin)

      * **Descrição:** Lista funcionários.
      * **Auth:** `Authorization: Bearer <admin_token>`
      * **Respostas:** `200 OK`: `[ { "id": 1, "nome": "João Func", "cargo": "Tecnico", "email": "joao@bytebros.com" } ]`

## 3\. Banco de Dados

### 3.1. Diagrama ER (Entidade-Relacionamento)

**Tabelas Principais:**

  * `usuarios`
  * `admin`
  * `funcionarios`
  * `produtos`
  * `servicos`
  * `noticias`
  * `orcamentos`
  * `suporte`
  * `pedidos`
  * `pedido_itens`

**Relacionamentos Chave:**

  * `usuarios` 1:N `pedidos` (Um usuário pode ter muitos pedidos). `pedidos.cliente_email` referencia `usuarios.email`.
  * `pedidos` 1:N `pedido_itens` (Um pedido tem muitos itens). `pedido_itens.pedido_id` referencia `pedidos.id`.
  * `produtos` 1:N `pedido_itens` (Um produto pode estar em muitos itens de pedido). `pedido_itens.produto_id` referencia `produtos.id`.
  * `usuarios` 1:N `suporte` (Um usuário pode ter muitas mensagens de suporte). `suporte.cliente_email` referencia `usuarios.email`.
  * `usuarios` 1:N `orcamentos` (Um usuário pode ter muitas solicitações de orçamento). `orcamentos.email_cliente` referencia `usuarios.email`.
  * `admin` (tabela separada para administradores com `is_admin` para superadmin).

*(**Nota:** Um diagrama visual ER (como um `draw.io` ou `dbdiagram.io`) seria ideal aqui, mas não posso gerá-lo diretamente. O ideal seria você criar um e incluí-lo na pasta `docs/images` ou similar.)*

### 3.2. Scripts de Criação de Tabelas

Os scripts DDL para criação das tabelas estão no arquivo `database.go`, na função `CreateTables()`. Quando o aplicativo Go é iniciado e se conecta ao banco de dados, ele verifica e cria as tabelas se elas ainda não existirem.

Exemplo de estrutura de tabela (`usuarios`):

```sql
CREATE TABLE IF NOT EXISTS usuarios (
    id SERIAL PRIMARY KEY,
    nome_completo VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    senha_hash VARCHAR(100) NOT NULL,
    telefone VARCHAR(20), -- Permite NULL se o usuário não informar
    criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    atualizado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 4\. Ambientes

### 4.1. Configuração do Ambiente Local (Docker)

Para replicar o ambiente local de desenvolvimento com Docker, siga estes passos:

1.  **Instale Docker e Docker Compose:** Certifique-se de ter Docker Desktop (Windows/macOS) ou Docker Engine e Compose (Linux) instalados em sua máquina.

2.  **Crie um arquivo `docker-compose.yml` na raiz do seu projeto:**

    ```yaml
    version: '3.8'
    services:
      db:
        image: postgres:15-alpine # Ou a versão que você usa no Supabase
        container_name: bytebros_db
        environment:
          POSTGRES_DB: ${DB_NAME}
          POSTGRES_USER: ${DB_USER}
          POSTGRES_PASSWORD: ${DB_PASS}
        ports:
          - "5432:5432" # Mapeia a porta do container para a porta local
        volumes:
          - db_data:/var/lib/postgresql/data # Persiste os dados do banco
        healthcheck: # Opcional: verifica a saúde do banco de dados
          test: ["CMD-SHELL", "pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB"]
          interval: 5s
          timeout: 5s
          retries: 5

      backend:
        build:
          context: ./bytebros.ti # Contexto de build: onde está seu Dockerfile (supondo que main.go está em bytebros.ti/)
          dockerfile: Dockerfile # Nome do Dockerfile (você precisará criar um)
        container_name: bytebros_backend
        environment:
          PORT: 8080 # Porta que seu app Go escuta
          DB_HOST: db # Nome do serviço do Docker Compose para o banco de dados
          DB_PORT: 5432
          DB_USER: ${DB_USER}
          DB_PASS: ${DB_PASS}
          DB_NAME: ${DB_NAME}
          JWT_SECRET: ${JWT_SECRET}
          GEMINI_API_KEY: ${GEMINI_API_KEY}
        ports:
          - "8080:8080" # Mapeia a porta do container para a porta local
        depends_on:
          db:
            condition: service_healthy # Espera o banco estar saudável
        volumes:
          - ./bytebros.ti:/app # Mapeia seu código Go para dentro do container (para live-reloading se usar)
        command: ["./server"] # Comando para iniciar (precisa ser o nome do executável)

    volumes:
      db_data: {}
    ```

3.  **Crie um `Dockerfile` para sua aplicação Go (na pasta `bytebros.ti/`):**

    ```dockerfile
    # Dockerfile (na pasta bytebros.ti/)
    # Use a imagem oficial do Go para build
    FROM golang:1.22-alpine AS builder

    # Define o diretório de trabalho dentro do container
    WORKDIR /app

    # Copia os arquivos go.mod e go.sum
    COPY go.mod ./
    COPY go.sum ./

    # Baixa as dependências
    RUN go mod download

    # Copia o código-fonte
    COPY . .

    # Compila a aplicação
    # -o server: define o nome do executável como 'server'
    # .: Compila o código no diretório atual
    RUN go build -o server .

    # Use uma imagem menor para o runtime final para reduzir o tamanho da imagem
    FROM alpine:latest

    WORKDIR /app

    # Copia o executável e o arquivo .env (se você usa em produção)
    COPY --from=builder /app/server .
    # Se você usa .env no deploy, deve copiar. Render usa variáveis de ambiente direto.
    # COPY --from=builder /app/.env .

    # Expõe a porta que o aplicativo Go escuta
    EXPOSE 8080

    # Comando para executar a aplicação
    CMD ["./server"]
    ```

4.  **Crie um arquivo `.env` na raiz do projeto** (no mesmo nível do `docker-compose.yml`) e preencha as variáveis de ambiente necessárias para o Docker Compose:

    ```
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=postgres
    DB_PASS=sua_senha_segura_local
    DB_NAME=bytebros_db
    PORT=8080
    JWT_SECRET=sua_jwt_secret_muito_segura_aqui
    GEMINI_API_KEY=sua_gemini_api_key_aqui
    ```

5.  **Inicie o Ambiente:**

      * Abra o terminal na raiz do seu projeto (onde está o `docker-compose.yml`).
      * Execute: `docker-compose up --build`
      * Isso irá construir as imagens, iniciar os contêineres e o seu backend Go estará acessível em `http://localhost:8080`.

## 5\. Requisitos Técnicos

Para desenvolver e executar este projeto, as seguintes dependências e ferramentas são necessárias:

  * **Go (Golang):** Versão 1.22 ou superior.
  * **Git:** Para controle de versão.
  * **Node.js e npm/yarn:** Para gerenciamento de pacotes frontend (se aplicável para ferramentas de build, embora seu projeto seja HTML/CSS/JS puro, pode ser útil para ferramentas de formatação ou linters).
  * **PostgreSQL:** Banco de dados (local ou em nuvem como Supabase).
  * **Docker e Docker Compose:** (Opcional, mas recomendado para ambiente de desenvolvimento).
  * **VS Code (ou IDE de sua preferência):** Com extensões de Go, HTML, CSS, JavaScript.

**Bibliotecas Go (go.mod):**

```go
require (
	github.com/gin-gonic/gin v1.10.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.1.1 // Driver PostgreSQL
	golang.org/x/crypto v0.23.0 // Para bcrypt
	github.com/golang-jwt/jwt/v4 v4.5.0 // JWT
	github.com/gin-contrib/cors v1.7.1 // Middleware CORS
	github.com/google/generative-ai-go/genai v0.12.0 // Gemini AI
	google.golang.org/api v0.180.0 // Dependência do Gemini SDK
)

// Inconsistências ou dependências indiretas podem aparecer com `go mod tidy`
```

-----

Com esta documentação, você terá um guia completo para o seu projeto. Lembre-se de substituir os placeholders `[Seu Nome/Gerente de Projetos]` e os valores de exemplo por seus dados reais.
