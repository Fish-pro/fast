package plugins

import "context"

const (
	DefaultUser     = "admin"
	DefaultPassword = "admin"
)

type Authorization struct {
	User     string
	Password string
}

func (a *Authorization) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"user": a.User, "password": a.Password}, nil
}

func (a *Authorization) RequireTransportSecurity() bool {
	return false
}
