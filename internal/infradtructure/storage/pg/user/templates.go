package user

const (
	CreateUserTemplate = `
	INSERT INTO users 
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id;`

	GetUserByEmailTemplate = `
	SELECT id, email, first_name, last_name, password_salt 
	FROM users WHERE email = $1;`

	UpdateUserTemplate = `
	UPDATE users
	SET email = $2, first_name = $3, last_name = $4, password_salt = $5
	WHERE email = $1
	RETURNING id;`

	DeleteUserTemplate = `
	DELETE FROM users
	WHERE email = $1
	RETURNING id;`
)
