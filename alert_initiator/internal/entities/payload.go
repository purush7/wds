package entities

type BatcherPayload struct {
	FilePath string `json:"filePath"`
	Id       string `json:"id"`
}

type NotifierPayload struct {
	UserId       string `json:"userId"`
	AlertMessage string `json:"alertMessage"`
}
