package services

// Service describes a service that adds things together.
type Service struct {
	UserService IUserService
	AuthService IAuthService
}

// New returns a basic Service.
func NewService(user IUserService, auth IAuthService) Service {
	return Service{UserService: user, AuthService: auth}
}
