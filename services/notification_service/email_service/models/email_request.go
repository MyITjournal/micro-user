package models

type EmailRequest struct {
    To       string
    Subject  string
    HtmlBody string
    TextBody string
}
