package api

// checkRequest est le corps attendu par POST /v1/checks. Les pointeurs des
// options distinguent l'absence d'un champ (valeur par défaut) d'un zéro
// explicite (qui sera rejeté par la validation des bornes).
type checkRequest struct {
	URLs    []string        `json:"urls"`
	Options *requestOptions `json:"options"`
}

type requestOptions struct {
	Concurrency *int `json:"concurrency"`
	TimeoutMs   *int `json:"timeout_ms"`
}

// errorBody est le contrat d'erreur uniforme de l'API.
type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
