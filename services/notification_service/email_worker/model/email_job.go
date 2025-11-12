package models

type EmailJob struct {
    UserID     string `json:"user_id"`
    To         string `json:"to"`
    Subject    string `json:"subject"`
    HtmlBody   string `json:"html"`
    TextBody   string `json:"text,omitempty"`
    TemplateID string `json:"template_id"`
    RequestID  string `json:"request_id"`
}
