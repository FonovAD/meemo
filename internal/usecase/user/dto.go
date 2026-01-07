package user

type CreateUserDtoIn struct {
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
	Password  string `db:"password"`
}
type CreateUserDtoOut struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
type GetUserInfoDtoIn struct {
	AccessToken string `json:"access_token"`
}
type GetUserInfoOut struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}
type UserDtoIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type UserDtoOut struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UpdateTokenDtoIn struct {
	RefreshToken string `db:"refresh_token"`
}
type UpdateTokenDtoOut struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
type LogoutDtoIn struct {
	AccessToken string `db:"access_token"`
}
type LogoutDtoOut struct {
}
