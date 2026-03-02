package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"backend/models"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

// GetPosts returns all published blog posts
func GetPosts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT id, slug, title, content, excerpt, published_at, created_at, updated_at, is_published, tags
			FROM blog_posts
			WHERE is_published = true
			ORDER BY published_at DESC
		`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		posts := []models.BlogPost{}
		for rows.Next() {
			var post models.BlogPost
			err := rows.Scan(
				&post.ID,
				&post.Slug,
				&post.Title,
				&post.Content,
				&post.Excerpt,
				&post.PublishedAt,
				&post.CreatedAt,
				&post.UpdatedAt,
				&post.IsPublished,
				pq.Array(&post.Tags),
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			posts = append(posts, post)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	}
}

// GetPostBySlug returns a single blog post by slug
func GetPostBySlug(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		slug := vars["slug"]

		var post models.BlogPost
		err := db.QueryRow(`
			SELECT id, slug, title, content, excerpt, published_at, created_at, updated_at, is_published, tags
			FROM blog_posts
			WHERE slug = $1 AND is_published = true
		`, slug).Scan(
			&post.ID,
			&post.Slug,
			&post.Title,
			&post.Content,
			&post.Excerpt,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.IsPublished,
			pq.Array(&post.Tags),
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(post)
	}
}

// CreatePost creates a new blog post
func CreatePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var id int
		var publishedAt *time.Time
		if req.IsPublished {
			now := time.Now()
			publishedAt = &now
		}

		err := db.QueryRow(`
			INSERT INTO blog_posts (slug, title, content, excerpt, is_published, published_at, tags)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`, req.Slug, req.Title, req.Content, req.Excerpt, req.IsPublished, publishedAt, pq.Array(req.Tags)).Scan(&id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	}
}

// UpdatePost updates an existing blog post
func UpdatePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var req models.UpdatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		query := "UPDATE blog_posts SET updated_at = CURRENT_TIMESTAMP"
		args := []interface{}{}
		argPos := 1

		if req.Slug != nil {
			query += ", slug = $" + strconv.Itoa(argPos)
			args = append(args, *req.Slug)
			argPos++
		}
		if req.Title != nil {
			query += ", title = $" + strconv.Itoa(argPos)
			args = append(args, *req.Title)
			argPos++
		}
		if req.Content != nil {
			query += ", content = $" + strconv.Itoa(argPos)
			args = append(args, *req.Content)
			argPos++
		}
		if req.Excerpt != nil {
			query += ", excerpt = $" + strconv.Itoa(argPos)
			args = append(args, *req.Excerpt)
			argPos++
		}
		if req.IsPublished != nil {
			query += ", is_published = $" + strconv.Itoa(argPos)
			args = append(args, *req.IsPublished)
			argPos++
			
			if *req.IsPublished {
				query += ", published_at = CURRENT_TIMESTAMP"
			}
		}
		if req.Tags != nil {
			query += ", tags = $" + strconv.Itoa(argPos)
			args = append(args, pq.Array(*req.Tags))
			argPos++
		}

		query += " WHERE id = $" + strconv.Itoa(argPos)
		args = append(args, id)

		result, err := db.Exec(query, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Post updated successfully"})
	}
}

// DeletePost deletes a blog post
func DeletePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		result, err := db.Exec("DELETE FROM blog_posts WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted successfully"})
	}
}
