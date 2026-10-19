package domain

import "time"

type Player struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	PrimaryRole    string    `json:"primary_role"`
	SecondaryRoles []string  `json:"secondary_roles"`
	Birthdate      time.Time `json:"birthdate"`
	Club           string    `json:"club"`
	Category       string    `json:"category"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
