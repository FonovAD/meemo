package user

const (
	CreateUserTemplate = `
	INSERT INTO users (first_name, last_name, email, password_salt) 
	VALUES (:first_name, :last_name, :email, :password_salt)
	RETURNING id;`

	GetUserByEmailTemplate = `
	SELECT id, first_name, last_name, email, password_salt 
	FROM users WHERE email = $1;`

	UpdateUserTemplate = `
	UPDATE users
	SET first_name = :first_name, last_name = :last_name, password_salt = :password_salt
	WHERE email = :email
	RETURNING id;`

	DeleteUserTemplate = `
	DELETE FROM users
	WHERE email = $1
	RETURNING id;`

	UpdateUserEmailTemplate = `
	UPDATE users
	SET email = $2
	WHERE email = $1
	RETURNING id;`

	CheckPassword = `
	SELECT EXISTS (
		SELECT 1
		FROM users
		WHERE email = $1
		  AND password_salt = $2
	) AS is_valid;)`
)
