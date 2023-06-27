package response

const (
	// ---- Part of technical errors on our side (can not unmarshal request etc) ----

	InvalidRequestStructure ErrCode = 10001

	// ---- Part of user errors. Missed argument, invalid type, etc ----

	WrongValue        ErrCode = 11000
	MissedValue       ErrCode = 11001
	UnsupportedValue  ErrCode = 11002
	WrongFormat       ErrCode = 11003
	UnsupportedAction ErrCode = 11004
)

type ErrCode uint
