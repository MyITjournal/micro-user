package models

import "time"

type Template struct {
	TemplateID string             `json:"template_id"`
	Name       string             `json:"name"`
	Version    string             `json:"version"`
	Language   string             `json:"language"`
	Type       string             `json:"type"`
	Subject    string             `json:"subject,omitempty"`
	Body       TemplateBody       `json:"body"`
	Variables  []TemplateVariable `json:"variables"`
	Metadata   TemplateMetadata   `json:"metadata"`
}

type TemplateBody struct {
	HTML string `json:"html,omitempty"`
	Text string `json:"text"`
}

type TemplateVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type TemplateMetadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	Tags      []string  `json:"tags"`
}

type RenderRequest struct {
	Language    string                 `json:"language"`
	Version     string                 `json:"version,omitempty"`
	Variables   map[string]interface{} `json:"variables"`
	PreviewMode bool                   `json:"preview_mode"`
}

type RenderResponse struct {
	TemplateID    string       `json:"template_id"`
	Language      string       `json:"language"`
	Version       string       `json:"version"`
	Subject       string       `json:"subject,omitempty"`
	Rendered      TemplateBody `json:"rendered"`
	RenderedAt    time.Time    `json:"rendered_at"`
	VariablesUsed []string     `json:"variables_used"`
}
