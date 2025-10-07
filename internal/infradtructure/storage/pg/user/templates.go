package user

const (
	CreateUserTemplate = `
	INSERT INTO users 
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;`

	GetUserByEmailTemplate = `
	SELECT id, first_name, last_name, email, password_salt 
	FROM users WHERE email = $1;`

	UpdateUserTemplate = `
	UPDATE users
	SET first_name = $2, last_name = $3, password_salt = $4
	WHERE email = $1
	RETURNING id;`

	DeleteUserTemplate = `
	DELETE FROM users
	WHERE email = $1
	RETURNING id;`
)
