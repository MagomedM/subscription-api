package models

type SubUser struct {
	ServiceName string `json:"service_name"`
	Price       int64  `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type SubsFull struct {
	Subs []*SubUser `json:"subs"`
}

type SubSum struct {
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	UserID      string  `json:"user_id,omitempty"`
	ServiceName string  `json:"service_name,omitempty"`
	TotalSum    float64 `json:"total_sum"`
}

type PaginatedResponse struct {
	Subs       []*SubUser `json:"subs"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page    int   `json:"page"`
	Limit   int   `json:"limit"`
	Total   int64 `json:"total"`
	Pages   int64 `json:"pages"`
	HasNext bool  `json:"has_next"`
	HasPrev bool  `json:"has_prev"`
}
