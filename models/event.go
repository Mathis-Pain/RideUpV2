package models

import (
	"database/sql"
	"time"
)

type Event struct {
	ID             int            `json:"id"`
	Title          string         `json:"title"`
	Description    sql.NullString `json:"-"`           // ✅ SANS pointeur
	DescriptionStr string         `json:"description"` // ✅ Pour le JSON
	CreatedBy      int            `json:"created_by"`
	CreatorName    string         `json:"creator_name"`
	CreatedAt      time.Time      `json:"created_at"`
	Latitude       float64        `json:"latitude"`
	Longitude      float64        `json:"longitude"`
	Address        string         `json:"address"`
	StartDatetime  time.Time      `json:"start_datetime"`
	EndDatetime    *time.Time     `json:"end_datetime,omitempty"`
	Participants   int            `json:"participants"`
	UserJoined     bool           `json:"user_joined"`
	FormattedStart string         `json:"-"` // ✅ Pour le template HTML
}
