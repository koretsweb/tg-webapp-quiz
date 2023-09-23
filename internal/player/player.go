package player

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/rs/xid"
)

type Player struct {
	ID        xid.ID    `json:"id" bson:"_id"`
	Version   xid.ID    `json:"version" bson:"version"`
	Email     string    `json:"email" bson:"email"`
	Name      string    `json:"name" bson:"name"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type playerJSON struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

func (p Player) MarshalJSON() ([]byte, error) {
	pl := playerJSON{
		ID:        p.ID.String(),
		Version:   p.Version.String(),
		Email:     p.Email,
		Name:      p.Name,
		UpdatedAt: p.UpdatedAt.UTC().Format(time.RFC3339),
		CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339),
	}

	return sonic.ConfigFastest.Marshal(pl)
}

type FilterRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}
