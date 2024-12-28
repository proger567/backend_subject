package app

type Subject struct {
	ID             int    `json:"id"`
	Comment        string `json:"comment,omitempty"`
	DateCreate     string `json:"date_create"`
	Description    string `json:"description,omitempty"`
	LastTimeUpdate string `json:"last_time_update"`
	Name           string `json:"name"`
	Type           string `json:"type,omitempty"`
	ParentID       int    `json:"parent_id,omitempty"`
}
