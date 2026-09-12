package user

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (u *UserService) CreateUser() (*User, error) {

	return nil, nil
}