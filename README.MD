# Desafio Client-Server-API (Go Expert)

Este projeto implementa um sistema client-server em Go para consulta de cotação do Dólar (USD-BRL) aplicando conceitos de contextos para controle de timeouts, consumo de API externa, persistência em SQLite e manipulação de arquivos.

## Requisitos Prévios

- **Go** (versão 1.18 ou superior)
- **GCC / CGO** habilitado (necessário para o driver do SQLite `github.com/mattn/go-sqlite3`)

## Instalação das Dependências

Instale as dependências necessárias executando:

```bash

go mod tidy
go run server.go
go run client.go