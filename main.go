package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {
	// .envファイルを読み込む
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	c := &cobra.Command{Use: "api [command]"}

	c.AddCommand(
		&cobra.Command{Use: "start", Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting API")

			file, err := os.Create("example.txt")
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()

			_, err = file.WriteString("This is a sample text.")
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}

			fmt.Println("File created successfully")
		}},
		&cobra.Command{Use: "create-json", Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Creating empty JSON file")

			file, err := os.Create("api.json")
			if err != nil {
				fmt.Println("Error creating JSON file:", err)
				return
			}
			defer file.Close()

			_, err = file.WriteString("{}")
			if err != nil {
				fmt.Println("Error writing to JSON file:", err)
				return
			}

			fmt.Println("Empty JSON file created successfully")
		}},
		&cobra.Command{Use: "get", Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Stopping API")

			databaseID := os.Getenv("NOTION_DATABASE_ID")
			url := fmt.Sprintf("https://api.notion.com/v1/databases/%s", databaseID)

			req, _ := http.NewRequest("GET", url, nil)

			NOTION_ACCESS_TOKEN := os.Getenv("NOTION_ACCESS_TOKEN")

			req.Header.Add("accept", "application/json")
			req.Header.Add("Notion-Version", "2022-06-28")
			req.Header.Add("Authorization", "Bearer "+NOTION_ACCESS_TOKEN)

			res, _ := http.DefaultClient.Do(req)

			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)

			fmt.Println(string(body))
		}},
		&cobra.Command{Use: "create-page", Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Creating page from api.json")

			file, err := os.Open("api.json")
			if err != nil {
				fmt.Println("Error opening JSON file:", err)
				return
			}
			defer file.Close()

			var data []map[string]interface{}
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&data); err != nil {
				fmt.Println("Error decoding JSON:", err)
				return
			}

			NOTION_ACCESS_TOKEN := os.Getenv("NOTION_ACCESS_TOKEN")

			for _, pageData := range data {
				url := "https://api.notion.com/v1/pages"

				jsonData, err := json.Marshal(pageData)
				if err != nil {
					fmt.Println("Error marshalling JSON:", err)
					continue
				}

				req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
				if err != nil {
					fmt.Println("Error creating request:", err)
					continue
				}

				req.Header.Add("Content-Type", "application/json")
				req.Header.Add("Notion-Version", "2022-06-28")
				req.Header.Add("Authorization", "Bearer "+NOTION_ACCESS_TOKEN)

				res, err := http.DefaultClient.Do(req)
				if err != nil {
					fmt.Println("Error making request:", err)
					continue
				}
				defer res.Body.Close()

				body, err := io.ReadAll(res.Body)
				if err != nil {
					fmt.Println("Error reading response:", err)
					continue
				}

				fmt.Println("Response from Notion API:", string(body))
			}

			// 処理が完了したらapi.jsonを空にする
			file, err = os.Create("api.json")
			if err != nil {
				fmt.Println("Error creating JSON file:", err)
				return
			}
			defer file.Close()
		}},
	)

	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
