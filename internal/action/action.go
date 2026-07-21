package action

import "github.com/anteiro255/gedis/internal/db"

type ActionType uint8

const (
	Set ActionType = iota
	Get
	Del
	Exist
	TTL_Set
	TTL_Get
	TTL_Del
	TTL_Exist
)

type Action struct {
	DB         *db.DB
	ActionType ActionType
	Key        db.Key
	Body       []byte
}

func (a *Action) Perform() error {
	switch a.ActionType {
	case Set:
		return a.DB.Set(a.Key, *db.NewVal(a.Body).AddTTL(db.TTLNever{}))
	case Get:
	case Del:
	case Exist:
	case TTL_Set:
	case TTL_Get:
	case TTL_Del:
	case TTL_Exist:

	}
	return nil
}
