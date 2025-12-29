package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagernet/sing-box/common/srs"
	"github.com/sagernet/sing-box/option"
)

func main() {
	// Папки для сохранения результатов
	os.MkdirAll("sing-geosite", 0755)
	os.MkdirAll("sing-geoip", 0755)

	// 1. Обработка Доменов (Geosite)
	fmt.Println("--- Processing Domains ---")
	processFiles("allow-domains/Services", "sing-geosite", true)

	// 2. Обработка IP (Geoip)
	fmt.Println("\n--- Processing IPs ---")
	processFiles("allow-domains/Subnets/IPv4", "sing-geoip", false)
}

func processFiles(inputDir string, outputDir string, isDomain bool) {
	files, err := filepath.Glob(filepath.Join(inputDir, "*.lst"))
	if err != nil {
		fmt.Println("Error reading glob:", err)
		return
	}

	for _, file := range files {
		fileName := filepath.Base(file)
		ruleName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		fmt.Printf("Processing: %s -> %s.srs\n", fileName, ruleName)

		lines := readLines(file)
		if len(lines) == 0 {
			continue
		}

		// Создаем строго типизированную структуру опций
		// Вместо map используем конкретный struct из библиотеки
		ruleOptions := option.DefaultHeadlessRule{}
		
		if isDomain {
			ruleOptions.DomainSuffix = lines
		} else {
			ruleOptions.IPCIDR = lines
		}

		// Создаем структуру правила
		plainRuleSet := option.PlainRuleSet{
			Rules: []option.HeadlessRule{
				{
					Type:           "default", // Используем строку вместо константы
					DefaultOptions: ruleOptions,
				},
			},
		}

		// Открываем файл для записи
		outputPath := filepath.Join(outputDir, ruleName+".srs")
		f, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", outputPath, err)
			os.Exit(1)
		}

		// Компилируем и записываем
		// Третий аргумент false - это параметр upgrade (нам он не нужен для новых файлов)
		err = srs.Write(f, plainRuleSet, false)
		f.Close() 

		if err != nil {
			fmt.Printf("Error compiling %s: %v\n", ruleName, err)
			os.Exit(1)
		}
	}
}

func readLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
