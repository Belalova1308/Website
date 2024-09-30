package titles

import (
	"database/sql"
	"net/http"
	"strconv"
	"text/template"
)

type Article struct {
	ID      int
	Title   string
	Content string
}

func InputHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			title := r.FormValue("title")
			content := r.FormValue("content")

			_, err := db.Exec("INSERT INTO articles (title, content) VALUES ($1, $2)", title, content)
			if err != nil {
				http.Error(w, "Failed to save article", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/articles", http.StatusSeeOther)
			return
		}
		tmpl := template.Must(template.ParseFiles("./static/form.html"))
		tmpl.Execute(w, nil)
	}
}
func ArticleHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			http.Error(w, "Invalid article ID", http.StatusBadRequest)
			return
		}

		article, err := GetArticleByID(db, id)
		if err != nil {
			http.Error(w, "Article not found", http.StatusNotFound)
			return
		}

		tmpl := template.Must(template.ParseFiles("./static/article.html"))
		tmpl.Execute(w, article)
	}
}

func GetArticles(db *sql.DB, limit, offset int) ([]Article, error) {
	rows, err := db.Query("SELECT id, title FROM articles LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var article Article
		if err := rows.Scan(&article.ID, &article.Title); err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func GetArticleByID(db *sql.DB, id int) (Article, error) {
	var article Article
	err := db.QueryRow("SELECT id, title, content FROM articles WHERE id = $1", id).Scan(&article.ID, &article.Title, &article.Content)
	if err != nil {
		return article, err
	}
	return article, nil
}
func GetArticlesCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func IndexHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		page := 1
		if pageStr != "" {
			p, err := strconv.Atoi(pageStr)
			if err == nil && p > 0 {
				page = p
			}
		}
		offset := (page - 1) * 3
		articles, err := GetArticles(db, 3, offset)
		if err != nil {
			http.Error(w, "Unable to load articles", http.StatusInternalServerError)
			return
		}
		totalCount, err := GetArticlesCount(db)
		if err != nil {
			http.Error(w, "Unable to count articles", http.StatusInternalServerError)
			return
		}
		hasPrev := page > 1
		hasNext := page*3 < totalCount
		previousPage := page - 1
		nextPage := page + 1
		data := struct {
			Articles     []Article
			Page         int
			HasPrev      bool
			HasNext      bool
			PreviousPage int
			NextPage     int
		}{
			Articles:     articles,
			Page:         page,
			HasPrev:      hasPrev,
			HasNext:      hasNext,
			PreviousPage: previousPage,
			NextPage:     nextPage,
		}
		tmpl := template.Must(template.ParseFiles("./static/index.html"))
		tmpl.Execute(w, data)
	}
}
