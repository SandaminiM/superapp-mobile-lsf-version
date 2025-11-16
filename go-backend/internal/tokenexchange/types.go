package tokenexchange

type TokenRequest struct {
	MicroAppID string `json:"microAppId" validate:"required,min=1"`
}

type JSONWebKey struct {
	Algorithm string `json:"alg"`
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	KeyID     string `json:"kid"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

type JSONWebKeySet struct {
	Keys []JSONWebKey `json:"keys"`
}
