package db

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
