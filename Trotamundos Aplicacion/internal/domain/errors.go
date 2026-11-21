package domain

import "errors"

var (
	ErrNotFound            = errors.New("recurso no encontrado")
	ErrForbidden           = errors.New("no tienes permisos para acceder a este viaje")
	ErrUnauthorized        = errors.New("sesión inválida o expirada")
	ErrInvalidCredentials  = errors.New("correo electrónico o contraseña incorrectos")
	ErrEmailAlreadyUsed    = errors.New("ya existe una cuenta registrada con ese correo electrónico")
	ErrValidation          = errors.New("datos de entrada inválidos")
	ErrMailboxUnauthorized = errors.New("debes autorizar el acceso a tu correo antes de escanear confirmaciones")
	ErrSenderNotAllowed    = errors.New("el remitente no se encuentra en tu lista de dominios confiables")
	ErrNoValidReservation  = errors.New("el mensaje no contiene una reserva válida o código de confirmación")
	ErrProviderUnavailable = errors.New("el proveedor externo no respondió a tiempo")
)
