package main

import (
	"bufio"
	"context"
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
	// Источник: allow-domains/Services/*.lst
	fmt.Println("--- Processing Domains ---")
	processFiles("allow-domains/Services", "sing-geosite", true)

	// 2. Обработка IP (Geoip)
	// Источник: allow-domains/Subnets/IPv4/*.lst
	fmt.Println("\n--- Processing IPs ---")
	processFiles("allow-domains/Subnets/IPv4", "sing-geoip", false)
}

func processFiles(inputDir string, outputDir string, isDomain bool) {
	files, err := filepath.Glob(filepath.Join(inputDir, "*.lst"))
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fileName := filepath.Base(file)
		ruleName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		
		fmt.Printf("Processing: %s -> %s.srs\n", fileName, ruleName)

		lines := readLines(file)
		if len(lines) == 0 {
			continue
		}

		// Создаем структуру правила
		headlessRule := option.HeadlessRule{
			Type:           option.HeadlessRuleTypeDefault,
			DefaultOptions: option.HeadlessRuleDefaultOptions{},
		}

		// Заполняем либо домены, либо IP
		if isDomain {
			// Для доменов используем domain_suffix
			headlessRule.DefaultOptions.DomainSuffix = lines
		} else {
			// Для IP используем ip_cidr
			headlessRule.DefaultOptions.IPCIDR = lines
		}

		// Компиляция в srs
		outputPath := filepath.Join(outputDir, ruleName+".srs")
		err := srs.Write(context.Background(), outputPath, headlessRule)
		if err != nil {
			fmt.Printf("Error compiling %s: %v\n", ruleName, err)
			os.Exit(1)
		}
	}
}

// Функция чтения файла построчно с очисткой
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
		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
