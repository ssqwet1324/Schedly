package entity

type EmailMessage struct {
	Subject   string `json:"title"`
	To        string `json:"email"`
	Content   string `json:"body"`
	EventTime string `json:"event_time"`
}
