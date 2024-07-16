package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	content := fmt.Sprintf("Dolar: %s", string(body))

	err = writeFileWithContext(ctx, "cotacao.txt", content)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	fmt.Println("Arquivo com a cotação do dólar criado com sucesso!")
}

func writeFileWithContext(ctx context.Context, filename, content string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	resultChan := make(chan error, 1)

	go func() {
		_, err := f.Write([]byte(content))
		resultChan <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled or timed out: %w", ctx.Err())
	case err := <-resultChan:
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
		return nil
	}
}
