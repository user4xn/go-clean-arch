package dto

type (
	ResponseRequestOtp struct {
		LastRequestOn string `json:"last_request_on"`
		NextRequestAt string `json:"next_request_at"`
	}

	ResponseFailVerifyOtp struct {
		AttemptLeft string `json:"attempt_left"`
	}

	PayloadOtp struct {
		Email string `json:"email" binding:"required"`
	}

	PayloadVerifyOtp struct {
		Email string `json:"email" binding:"required"`
		OTP   string `json:"otp" binding:"required"`
	}

	PayloadVerifyOtpTraced struct {
		Email     string `json:"email" binding:"required"`
		OTP       string `json:"otp" binding:"required"`
		IP        string `json:"ip"`
		UserAgent string `json:"user_agent"`
	}

	DataOtpEmail struct {
		OTP   string `json:"otp"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	OtpUpdateAttempt struct {
		Attempt int `json:"attempt"`
	}
)
