package user

const (
	CreateUserTemplate = `
	INSERT INTO user 
	VALUES ($1, $2, $3, $4, $5, $6, $7);`

	GetUserByEmailTemplate = `
	SELECT id, email, first_name, last_name, password_salt 
	FROM user WHERE email = $1;`

	UpdateUserTemplate = `
	UPDATE user 
	SET email = $2, first_name = $3, last_name = $4, password_salt = $5
	WHERE email = $1;`

	DeleteUserTemplate = `
	DELETE FROM user
	WHERE email = $1;`
)
