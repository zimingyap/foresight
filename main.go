package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type NewsArticle struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	PublishedAt time.Time `json:"publishedAt"`
	SourceName  string    `json:"sourceName"`
	URL         string    `json:"url"`
}

type NewsResponse struct {
	Status       string        `json:"status"`
	TotalResults int           `json:"totalResults"`
	Articles     []NewsArticle `json:"articles"`
}

func fetchNews(query string) ([]NewsArticle, error) {
	apiKey := os.Getenv("NEWS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("NEWS_API_KEY environment variable not set")
	}

	url := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&apiKey=%s", query, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Reading response body failed: %w", err)
	}

	var newsResponse NewsResponse
	err = json.Unmarshal(body, &newsResponse)
	if err != nil {
		return nil, fmt.Errorf("JSON unmarshaling failed: %w", err)
	}

	if newsResponse.Status != "ok" {
		return nil, fmt.Errorf("API returned an error: %s", newsResponse.Status)
	}

	return newsResponse.Articles, nil
}
func newsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		query = "finance"
	}

	articles, err := fetchNews(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching news: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("news").Parse(`
        <!DOCTYPE html>
        <html>
        <head>
                <title>Financial News</title>
                <script src="https://unpkg.com/htmx.org@1.9.4"></script>
        </head>
        <body>
            <form hx-get="/" hx-target="#news-container" hx-indicator="#loading">
                <input type="text" name="q" placeholder="Enter query" value="{{.Query}}">
                <button type="submit">Search</button>
            </form>
            <div id="loading" class="htmx-indicator">Loading...</div>
                <div id="news-container">
                        {{range .Articles}}
                                <div>
                                        <h3><a href="{{.URL}}" target="_blank">{{.Title}}</a></h3>
                    <p>{{.PublishedAt.Format "2006-01-02 15:04:05"}} - {{.SourceName}}</p>
                                        <p>{{.Description}}</p>
                                        <p>{{.Content}}</p>
                                        <hr>
                                </div>
                        {{end}}
                </div>
        </body>
        </html>
        `)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
		return
	}

	data := struct {
		Articles []NewsArticle
		Query    string
	}{
		Articles: articles,
		Query:    query,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/", newsHandler)
	fmt.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

	// articles, err := fetchNews("Nvidia")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for _, article := range articles {
	// 	fmt.Println("Title:", article.Title)
	// 	fmt.Println("PublishedAt:", article.PublishedAt)
	// 	fmt.Println("---")
	// }
}
