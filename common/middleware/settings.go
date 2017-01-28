package middleware

type MiddlewareSettings struct {
	Driver              string
	EncryptionKey       string
	User                string
	Host                string
	Port                int
	Pass                string
	RoutePublic         string
	RouteServer         string
	RouteFileServer     string
	RouteClient         string
	RouteInternalServer string
	PemFile             string
}
