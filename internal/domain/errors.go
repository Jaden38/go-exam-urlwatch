package domain

import (
	"errors"
	"fmt"
)

// ErrBatchNotFound est renvoyée par Store.Get lorsqu'aucun lot ne correspond à
// l'identifiant demandé.
var ErrBatchNotFound = errors.New("lot introuvable")

// ValidationError signale qu'un champ de la requête est invalide. Elle porte le
// nom du champ fautif afin que la couche API la traduise en réponse 400.
type ValidationError struct {
	Field   string
	Message string
}

// Error implémente l'interface error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
