# Projeto de Backup de Diretórios

## Visão Geral

Este projeto é uma aplicação em Go projetada para copiar recursivamente o conteúdo de um diretório para outro, mantendo a estrutura do diretório e copiando arquivos de forma concorrente. O aplicativo garante que o número de goroutines simultâneas não exceda um limite especificado para evitar uso excessivo de recursos.

## Instalação

1. Certifique-se de ter o Go instalado em seu sistema. Você pode baixá-lo em [golang.org](https://golang.org/dl/).


1. Clone o repositório:

   ```bash
   git clone https://github.com/seuusuario/backup-directory.git
   ```
   
2. Navegue até o diretório do projeto:
   ```bash
      cd backup-directory
   ```
   
3. Compile o projeto:
   ```bash
      go build -o copyPaste
   ```
   
## Uso

1. Execute a aplicação
   ```bash
      ./copyPaste
   ```
Por padrão, a aplicação tentará copiar o conteúdo do diretório `./teste` para o diretório `./teste2`.

1. Configuração:
   * **Quantidade de Goroutines:** Ajuste o número máximo de goroutines editando a constante maxGoroutines no arquivo main.go.
   * **Diretório de Origem e Destino:** Modifique os diretórios de origem (oriDir) e destino (dstDir) editando as variáveis na função main do arquivo main.go.
   