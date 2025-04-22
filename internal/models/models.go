package models

type Meta struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Notification struct {
	DocID int32 `json:"doc_id"`
}
