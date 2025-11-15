package handler

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"operation successful"`
}

type PaginationParams struct {
	Limit  int `json:"limit" example:"10"`
	Offset int `json:"offset" example:"0"`
}
