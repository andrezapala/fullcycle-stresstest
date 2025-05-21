package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type result struct {
	statusCode int
	duration   time.Duration
}

func worker(jobs <-chan int, results chan<- result, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	for range jobs {
		start := time.Now()
		resp, err := http.Get(url)
		duration := time.Since(start)
		if err != nil {
			results <- result{statusCode: 0, duration: duration}
			continue
		}
		results <- result{statusCode: resp.StatusCode, duration: duration}
		resp.Body.Close()
	}
}

func main() {
	url := flag.String("url", "", "URL do serviço a ser testado")
	totalRequests := flag.Int("requests", 100, "Número total de requisições")
	concurrency := flag.Int("concurrency", 10, "Número de requisições simultâneas")
	flag.Parse()

	if *url == "" || *totalRequests <= 0 || *concurrency <= 0 {
		flag.Usage()
		return
	}

	jobs := make(chan int, *totalRequests)
	results := make(chan result, *totalRequests)
	var wg sync.WaitGroup

	startTime := time.Now()

	// Inicia workers
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go worker(jobs, results, *url, &wg)
	}

	// Envia jobs
	for i := 0; i < *totalRequests; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	close(results)

	// Processa os resultados
	total200 := 0
	statusCounts := make(map[int]int)
	var totalDuration time.Duration

	for res := range results {
		if res.statusCode == 200 {
			total200++
		}
		statusCounts[res.statusCode]++
		totalDuration += res.duration
	}

	fmt.Println("===== Relatório de Teste de Carga =====")
	fmt.Printf("Tempo total: %v\n", time.Since(startTime))
	fmt.Printf("Total de requisições: %d\n", *totalRequests)
	fmt.Printf("Requisições com status 200: %d\n", total200)
	fmt.Println("Distribuição dos status HTTP:")
	for code, count := range statusCounts {
		fmt.Printf("Status %d: %d\n", code, count)
	}
}
