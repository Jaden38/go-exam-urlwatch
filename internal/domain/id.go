package domain

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID génère un identifiant de lot de la forme "b_4f3c1a".
func NewID() string {
	b := make([]byte, 3)
	// crypto/rand.Read ne renvoie une erreur qu'en cas de défaillance système.
	_, _ = rand.Read(b)
	return "b_" + hex.EncodeToString(b)
}
