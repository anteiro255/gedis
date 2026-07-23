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
		v := db.NewVal(a.Body)
		// No need to explicitly call SetTTL(db.TTLNever{}) as NewVal already initializes it to TTLNever
		return nil, a.DB.Set(a.Key, v)

	case protocol.Get:
		v, s := a.DB.Get(a.Key)
		if s != status.OK {
			return nil, s
		}
		// In a real system, we should check if the TTL is expired before returning
		if _, expired := v.TTL().(db.TTLExpired); expired {
			return nil, status.NoSuchKey
		}
		return v.Data(), status.OK

	case protocol.Del:
		return nil, a.DB.Del(a.Key)

	case protocol.Exist:
		if !a.DB.Exists(a.Key) {
			return nil, status.NoSuchKey
		}
		return nil, status.OK

	case protocol.TTL_Set:
		if len(a.Body) < 4 {
			return nil, status.WrongInput
		}
		ttl := db.TTLSeconds{Seconds: binary.BigEndian.Uint32(a.Body)}
		return nil, a.DB.SetTTL(a.Key, ttl)

	case protocol.TTL_Get:
		ttl, s := a.DB.GetTTL(a.Key)
		if s != status.OK {
			return nil, s
		}

		switch t := ttl.(type) {
		case db.TTLSeconds:
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, t.Seconds)
			return b, status.OK
		case db.TTLNever:
			// Define a protocol-specific value for "Never" or return a specific status
			return nil, status.OK
		case db.TTLExpired:
			return nil, status.NoSuchKey
		default:
			return nil, status.InternalError
		}

	case protocol.TTL_Del:
		return nil, a.DB.DelTTL(a.Key)

	case protocol.TTL_Exist:
		if !a.DB.ExistsTTL(a.Key) {
			return nil, status.NoSuchKey
		}
		return nil, status.OK

	default:
		return nil, status.WrongInput
	}
}
