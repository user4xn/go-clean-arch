package consts

import "time"

type (
	SessionStatus int
)

const (
	SessionActive  SessionStatus = 0
	SessionRevoked SessionStatus = 1

	TokenDurationRelease = time.Minute * 15
	TokenDurationDev     = time.Minute * 60

	RefreshTokenDayAgeRelease = 7
	RefreshTokenDayAgeDev     = 30
)
