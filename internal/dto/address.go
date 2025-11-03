package dto

type CreateAddressRequest struct {
	Address    string  `json:"address" binding:"required,min=10,max=200"`
	City       *string `json:"city" binding:"omitempty,min=2,max=100"`
	PostalCode *string `json:"postal_code" binding:"omitempty,min=3,max=12"`
	Country    *string `json:"country" binding:"omitempty,oneof=Russia Kazakhstan Belarus Armenia Kyrgyzstan Uzbekistan"`
}

type UpdateAddressRequest struct {
	Address    *string `json:"address" binding:"omitempty,min=10,max=200"`
	City       *string `json:"city" binding:"omitempty,min=2,max=100"`
	PostalCode *string `json:"postal_code" binding:"omitempty,min=3,max=12"`
	Country    *string `json:"country" binding:"omitempty,oneof=Russia Kazakhstan Belarus Armenia Kyrgyzstan Uzbekistan"`
}

type AddressResponse struct {
	ID         int64   `json:"id"`
	UUID       string  `json:"uuid"`
	Address    string  `json:"address"`
	City       *string `json:"city"`
	PostalCode *string `json:"postal_code"`
	Country    *string `json:"country"`
	CreatedAt  string  `json:"created_at"`
}
