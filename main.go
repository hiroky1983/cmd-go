package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

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

	databaseID := os.Getenv("NOTION_DATABASE_ID")
	NOTION_ACCESS_TOKEN := os.Getenv("NOTION_ACCESS_TOKEN")

	c.AddCommand(
		&cobra.Command{Use: "start", Short: "example.txtを作成するコマンド", Run: func(cmd *cobra.Command, args []string) {
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
		&cobra.Command{Use: "create-json", Short: "空のapi.jsonを作成するコマンド", Run: func(cmd *cobra.Command, args []string) {
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
		&cobra.Command{Use: "get", Short: "Notionのデータベースを取得するコマンド", Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Stopping API")

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
		&cobra.Command{Use: "create-page", Short: "api.jsonからNotionのページを作成するコマンド", Run: func(cmd *cobra.Command, args []string) {
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
			// file, err = os.Create("api.json")
			// if err != nil {
			// 	fmt.Println("Error creating JSON file:", err)
			// 	return
			// }
			// defer file.Close()
		}},
		&cobra.Command{Use: "gen-json", Short: "index.mdから直接api.jsonを作成するコマンド", Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Generating api.json directly from index.md")

			// index.mdファイルを開く
			file, err := os.Open("index.md")
			if err != nil {
				fmt.Println("Error opening index.md:", err)
				return
			}
			defer file.Close()

			var entries []map[string]interface{}
			scanner := bufio.NewScanner(file)
			var crmName, serviceName, methodName, description string
			servicePattern := regexp.MustCompile(`### (\w+Service)`)
				methodPattern := regexp.MustCompile(`\| (\w+) \| \[.*\]\(#.*\) \| \[.*\]\(#.*\) \| (.*) \|`)

			for scanner.Scan() {
				line := scanner.Text()

				// CRM名を取得
				if strings.Contains(line, "admin/v1") {
					crmName = "admin"
				} else if strings.Contains(line, "operation/v1") {
					crmName = "operation"
				} else if strings.Contains(line, "customer/v1") {
					crmName = "customer"
				}

				// Service名を取得
				if matches := servicePattern.FindStringSubmatch(line); matches != nil {
					serviceName = matches[1]
				}

				// Method名とDescriptionを取得
				if matches := methodPattern.FindStringSubmatch(line); matches != nil {
					methodName = matches[1]
					description = matches[2]

					// JSONエントリを作成
					entry := map[string]interface{}{
						"parent": map[string]string{
							"database_id": os.Getenv("NOTION_DATABASE_ID"),
						},
						"properties": map[string]interface{}{
							"エンドポイント": map[string]interface{}{
								"title": []map[string]interface{}{
									{
										"text": map[string]string{
											"content": fmt.Sprintf("/%s.v1.%s/%s", crmName, serviceName, methodName),
										},
									},
								},
							},
							"概要": map[string]interface{}{
								"rich_text": []map[string]interface{}{
									{
										"text": map[string]string{
											"content": description,
										},
									},
								},
							},
						},
					}

					entries = append(entries, entry)
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading index.md:", err)
				return
			}

			// api.jsonファイルを作成
			apiFile, err := os.Create("api.json")
			if err != nil {
				fmt.Println("Error creating api.json:", err)
				return
			}
			defer apiFile.Close()

			encoder := json.NewEncoder(apiFile)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(entries); err != nil {
				fmt.Println("Error writing to api.json:", err)
			}
		}},
	)

	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
