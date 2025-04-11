package data

type ApiErrorResponse struct {
	Message string `json:"msg"`
}

type RegisterFilePath struct {
	FilePath string `json:"filepath"`
}
