package dto

type (
	// Define common response struct
	Common struct {
		Status  string      `json:"status"`
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	PayloadBasicTable struct {
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
		Search string `json:"search"`
	}

	ResponseTotalRow struct {
		TotalRow int `json:"total_row"`
	}
)
