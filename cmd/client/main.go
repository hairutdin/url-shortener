package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func shortenURL(input io.Reader, endpoint string) error {
	data := url.Values{}
	fmt.Println("Введите длинный URL")

	reader := bufio.NewReader(input)
	longURL, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	longURL = strings.TrimSpace(longURL)
	data.Set("url", longURL)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Статус-код", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))

	return nil
}

func main() {
	err := shortenURL(os.Stdin, "http://localhost:8080/")
	if err != nil {
		panic(err)
	}
}
