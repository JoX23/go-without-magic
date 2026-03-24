package domain

import "errors"

// Errores de dominio para Product — tipados para que el handler los mapee
// correctamente a códigos HTTP o gRPC.
//
// REGLA: estos errores representan casos de negocio, NO errores técnicos.
var (
	ErrProductNotFound     = errors.New("product not found")
	ErrProductDuplicated   = errors.New("product already exists")
	ErrInvalidProductSku   = errors.New("sku cannot be empty")
	ErrInvalidProductName  = errors.New("name cannot be empty")
	ErrInvalidProductPrice = errors.New("price cannot be empty")
)
