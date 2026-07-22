package action

import (
	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/pkg/protocol"
)

type Action struct {
	DB         *db.DB
	ActionType protocol.ActionType
	Key        db.Key
	Body       []byte
}

func (a *Action) Perform() error {
	switch a.ActionType {
	case protocol.Set:
		return a.DB.Set(a.Key, *db.NewVal(a.Body).AddTTL(db.TTLNever{}))
	case protocol.Get:
	case protocol.Del:
	case protocol.Exist:
	case protocol.TTL_Set:
	case protocol.TTL_Get:
	case protocol.TTL_Del:
	case protocol.TTL_Exist:

	}
	return nil
}
