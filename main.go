package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	goRoutines        = 20
	maxMemory  uint64 = 1000 >> 20 // 1GB default
	debugMode         = false
	overwrite  bool
	oriDirs    []string
)

func filesEqual(file1, file2 string) (bool, error) {
	info1, err := os.Stat(file1)
	if err != nil {
		return false, err
	}
	info2, err := os.Stat(file2)
	if err != nil {
		return false, err
	}
	return info1.ModTime() == info2.ModTime(), nil
}

func copyFile(ori, dst string, wg *sync.WaitGroup, controlador chan struct{}) {
	defer wg.Done()
	controlador <- struct{}{}
	defer func() { <-controlador }()

	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			equal, err := filesEqual(ori, dst)
			if err != nil {
				log.Printf("Erro ao comparar arquivo %s, %v\n", ori, err)
				return
			}
			if equal {
				log.Printf("Arquivo de destino já existe, não será copiado: %s\n", dst)
				return
			}
		}
	}

	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if m.Alloc < maxMemory {
			break
		}
		if debugMode {
			fmt.Println("[DEBUG] Memória usada acima do limite, esperando...")
		}
		time.Sleep(1 * time.Second)
	}

	log.Printf("Copiando arquivo: %s para %s\n", ori, dst)
	if debugMode {
		fmt.Println("[DEBUG] Número atual de Goroutines: ", runtime.NumGoroutine())
	}
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

		origInfo, err := origem.Stat()
		if err == nil {
			err := os.Chtimes(dst, origInfo.ModTime(), origInfo.ModTime())
			if err != nil {
				log.Printf("Erro ao copiar data de modificação do arquivo de origem pro arquivo destino: %s , %v\n", ori, err)
				return
			}
		}
	}
}

func copyDir(origemDir, dstDir string, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()

	log.Printf("Copiando diretório: %s para %s\n", origemDir, dstDir)

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

		for runtime.NumGoroutine() > goRoutines {
			if debugMode {
				fmt.Println("[DEBUG] Numero máximo de goroutines atingido, aguardando liberar.... ", runtime.NumGoroutine(), " goroutines")
			}
		}

		if entry.IsDir() {
			wg.Add(1)
			copyDir(origemPath, dstPath, wg, sem)
		} else {
			wg.Add(1)
			go copyFile(origemPath, dstPath, wg, sem)
		}
	}
}

func main() {
	var (
		oriDir       string
		dstDir       string
		resp         string
		logName      string
		newdirString string
		generateLog  bool
	)

	for {
		if newdirString == "N" {
			break
		}
		fmt.Print("Digite o diretório de origem: ")
		fmt.Scanln(&oriDir)
		if _, err := os.Stat(oriDir); os.IsNotExist(err) {
			fmt.Println("Diretório de origem não existe.")
		} else {
			oriDirs = append(oriDirs, oriDir)
			for {
				fmt.Print("Vai adicionar outro diretório? (s/n) ")
				fmt.Scanln(&newdirString)

				newdirString = strings.ToUpper(newdirString)
				if newdirString == "S" {
					break
				} else if newdirString == "N" {
					fmt.Println("Lista de diretórios de origem:", oriDirs)
					break
				} else {
					fmt.Println("Resposta inválida. Digite 's' para sim ou 'n' para não.")
				}
			}
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
	for {
		fmt.Print("Quantidade de memória RAM máxima que poderá ser utilizada (GB): ")
		_, err := fmt.Scanln(&maxMemory)
		if err != nil {
			fmt.Println("Digite somente números.")
			continue
		}
		if maxMemory <= 1 {
			fmt.Println("Quantidade inválida.")
		} else {
			maxMemory = (maxMemory * 1000) << 20
			break
		}
	} // mem ram limitação
	for {
		var overwriteString string
		fmt.Print("Substituir arquivos existentes? Se 'não' for escolhido irá ignorar arquivos já existentes. (s/n) ")
		_, err := fmt.Scanln(&overwriteString)
		if err != nil {
			fmt.Println("Digite somente letras.")
			continue
		}
		overwriteString = strings.ToUpper(overwriteString)
		if overwriteString == "S" {
			overwrite = true
			break
		} else if overwriteString == "N" {
			overwrite = false
			break
		} else {
			fmt.Println("Resposta inválida.")
		}
	} // Susbtituir
	for {
		var debugString string
		fmt.Print("Debug Mode? (s/n) ")
		_, err := fmt.Scanln(&debugString)
		if err != nil {
			fmt.Println("Digite somente letras.")
			continue
		}
		debugString = strings.ToUpper(debugString)
		if debugString == "S" {
			debugMode = true
			break
		} else if debugString == "N" {
			debugMode = false
			break
		} else {
			fmt.Println("Resposta inválida.")
		}
	} // debug mode

	controlador := make(chan struct{}, goRoutines)
	var wg sync.WaitGroup

	//TODO: Fazer um controle para que todas as goroutines trabalhem em cada item da lista por vez.

	for _, dir := range oriDirs {
		wg.Add(1)
		dstSubDir := filepath.Join(dstDir, filepath.Base(dir))
		go copyDir(dir, dstSubDir, &wg, controlador)
	}

	wg.Wait()
	log.Println("Backup concluído.")
}
