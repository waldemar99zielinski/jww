package schema

import "jww/internal/domain"

type CurrencyCreateRequest struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	IsBase bool   `json:"is_base"`
}

func (r CurrencyCreateRequest) Validate() error {
	return firstError(
		exactLength("code", r.Code, 3),
		minLength("name", r.Name, 1),
	)
}

type CurrencyUpdateRequest struct {
	Name   string `json:"name"`
	IsBase bool   `json:"is_base"`
}

func (r CurrencyUpdateRequest) Validate() error {
	return minLength("name", r.Name, 1)
}

type CurrencyResponse struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	IsBase bool   `json:"is_base"`
}

func CurrencyToResponse(c domain.Currency) CurrencyResponse {
	return CurrencyResponse{Code: c.Code, Name: c.Name, IsBase: c.IsBase}
}
