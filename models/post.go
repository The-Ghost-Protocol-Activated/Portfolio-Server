package models

import (
	"time"
)

type BlogPost struct {
	ID          int       `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Excerpt     string    `json:"excerpt"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsPublished bool      `json:"is_published"`
	Tags        []string  `json:"tags"`
}

type CreatePostRequest struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Excerpt     string   `json:"excerpt"`
	IsPublished bool     `json:"is_published"`
	Tags        []string `json:"tags"`
}

type UpdatePostRequest struct {
	Slug        *string   `json:"slug,omitempty"`
	Title       *string   `json:"title,omitempty"`
	Content     *string   `json:"content,omitempty"`
	Excerpt     *string   `json:"excerpt,omitempty"`
	IsPublished *bool     `json:"is_published,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
}
