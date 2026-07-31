package action

type Action uint8

const (
	Set Action = iota
	Get
	Del
	Exist
	TTL_Set
	TTL_Get
	TTL_Del
	TTL_Expire
)
