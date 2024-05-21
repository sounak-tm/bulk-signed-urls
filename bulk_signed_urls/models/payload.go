package models

type FileInfo struct {
	ID   string
	Name string
}

type RequestPayload struct {
	Files map[string]string `json:"files"`
}

type ResponsePayload struct {
	URLs map[string]string `json:"urls"`
}
