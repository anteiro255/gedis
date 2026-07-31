package action

import (
	"encoding/binary"

	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/pkg/protocol/action"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

type Action struct {
	DB         *db.DB
	ActionType action.Action
	Key        db.Key
	Body       []byte
}

func (a *Action) Perform() ([]byte, status.Status) {
	switch a.ActionType {
	case action.Set:
		a.DB.Set(a.Key, db.Val(a.Body))
		return nil, status.OK

	case action.Get:
		return a.DB.Get(a.Key)

	case action.Del:
		return nil, a.DB.Del(a.Key)

	case action.Exist:
		return nil, a.DB.Exists(a.Key)

	case action.TTL_Set:
		if len(a.Body) != 4 {
			return nil, status.WrongInput
		}
		ttl := db.TTL(binary.BigEndian.Uint32(a.Body))
		return nil, a.DB.SetTTL(a.Key, ttl)

	case action.TTL_Get:
		ttl, sts := a.DB.GetTTL(a.Key)

		ttlBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(ttlBytes, uint32(ttl))
		return ttlBytes, sts

	case action.TTL_Del:
		return nil, a.DB.DelTTL(a.Key)

	case action.TTL_Expire:
		if len(a.Body) != 4 {
			return nil, status.WrongInput
		}
		return nil, a.DB.Expire(a.Key, db.TTL(binary.BigEndian.Uint32(a.Body)))

	default:
		return nil, status.WrongInput
	}
}
