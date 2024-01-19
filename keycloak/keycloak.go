// https://www.keycloak.org/docs-api/21.0.1/rest-api
package keycloak

type Config struct {
	BaseUrl      string
	Realm        string
	ClientId     string
	ClientSecret string
}
