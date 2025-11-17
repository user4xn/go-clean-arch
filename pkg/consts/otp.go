package consts

type (
	AttemptCooldown int
)

const (
	AttemptOne   AttemptCooldown = 60
	AttemptTwo   AttemptCooldown = 120
	AttemptThree AttemptCooldown = 240
	AttemptFour  AttemptCooldown = 480
	AttemptFive  AttemptCooldown = 86400
)
