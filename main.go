package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func copyFile(ori, dst string, wg *sync.WaitGroup, controlador chan struct{}) {
	defer wg.Done()
	controlador <- struct{}{}

	origem, err := os.Open(ori)
	if err != nil {
		log.Printf("Erro ao abrir o arquivo de origem %s: %v\n", ori, err)
		<-controlador
		return
	}
	defer origem.Close()

	destino, err := os.Create(dst)
	if err != nil {
		log.Printf("Erro ao criar o arquivo de destino %s: %v\n", dst, err)
		<-controlador
		return
	}
	defer destino.Close()

	_, err = io.Copy(destino, origem)
	if err != nil {
		log.Printf("Erro ao copiar o arquivo de %s para %s: %v\n", ori, dst, err)
	} else {
		log.Printf("Arquivo copiado com sucesso: %s para %s\n", ori, dst)
	}

	<-controlador
}

func copyDir(origemDir, dstDir string, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()

	entries, err := ioutil.ReadDir(origemDir)
	if err != nil {
		log.Printf("Erro ao ler o diretório de origem %s: %v\n", origemDir, err)
		return
	}

	err = os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		log.Printf("Erro ao criar o diretório de destino %s: %v\n", dstDir, err)
		return
	}

	for _, entry := range entries {
		origemPath := filepath.Join(origemDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			wg.Add(1)
			go copyDir(origemPath, dstPath, wg, sem)
		} else {
			wg.Add(1)
			go copyFile(origemPath, dstPath, wg, sem)
		}
	}
}

func main() {
	var (
		oriDir      string
		dstDir      string
		resp        string
		logName     string
		generateLog bool
		goRoutines  int
	)

	for {
		fmt.Print("Digite o diretório de origem: ")
		fmt.Scanln(&oriDir)
		if _, err := os.Stat(oriDir); os.IsNotExist(err) {
			fmt.Println("Diretório de origem não existe.")
		} else {
			break
		}
	} // oriDir

	for {
		fmt.Print("Digite o diretório de destino: ")
		fmt.Scanln(&dstDir)
		if _, err := os.Stat(dstDir); os.IsNotExist(err) {
			fmt.Println("Diretório de destino não existe.")
		} else {
			break
		}
	} // dstDir

	for {
		fmt.Print("Deseja gerar um arquivo de log? (s/n): ")
		_, err := fmt.Scanln(&resp)
		if err != nil {
			fmt.Println("Digite somente letras aqui.")
			return
		}
		resp = strings.ToUpper(resp)
		if resp == "S" {
			generateLog = true
			break
		} else if resp == "N" {
			generateLog = false
			break
		} else {
			fmt.Println("Resposta inválida.")
		}
	} // generateLog

	if generateLog {
		for {
			fmt.Print("Digite o nome do arquivo de log: ")
			_, err := fmt.Scanln(&logName)
			if err != nil {
				fmt.Println("Nome do arquivo inválido.")
				continue
			}
			break
		} // logName
		fileLog, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Erro ao criar arquivo de log: %v\n", err)
		} else {
			defer fileLog.Close()
			log.SetOutput(fileLog)
		}
	}

	for {
		fmt.Print("Quantidade de arquivos que serão copiados simultaneamente: ")
		_, err := fmt.Scanln(&goRoutines)
		if err != nil {
			fmt.Println("Digite somente números.")
			continue
		}
		if goRoutines <= 0 {
			fmt.Println("Quantidade inválida.")
		} else {
			break
		}
	} // goRoutines

	controlador := make(chan struct{}, goRoutines)
	var wg sync.WaitGroup

	wg.Add(1)
	go copyDir(oriDir, dstDir, &wg, controlador)

	wg.Wait()
	log.Println("Backup concluído.")
}
