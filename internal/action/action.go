package action

import (
	"encoding/binary"

	"github.com/anteiro255/gedis/internal/db"
	"github.com/anteiro255/gedis/pkg/protocol"
	"github.com/anteiro255/gedis/pkg/protocol/status"
)

type Action struct {
	DB         *db.DB
	ActionType protocol.ActionType
	Key        db.Key
	Body       []byte
}

func (a *Action) Perform() ([]byte, status.Status) {
	switch a.ActionType {
	case protocol.Set:
		a.DB.Set(a.Key, db.Val(a.Body))
		return nil, status.OK

	case protocol.Get:
		return a.DB.Get(a.Key)

	case protocol.Del:
		return nil, a.DB.Del(a.Key)

	case protocol.Exist:
		return nil, a.DB.Exists(a.Key)

	case protocol.TTL_Set:
		if len(a.Body) != 4 {
			return nil, status.WrongInput
		}
		ttl := db.TTL(binary.BigEndian.Uint32(a.Body))
		return nil, a.DB.SetTTL(a.Key, ttl)

	case protocol.TTL_Get:
		ttl, sts := a.DB.GetTTL(a.Key)

		ttlBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(ttlBytes, uint32(ttl))
		return ttlBytes, sts

	case protocol.TTL_Del:
		return nil, a.DB.DelTTL(a.Key)

	default:
		return nil, status.WrongInput
	}
}
