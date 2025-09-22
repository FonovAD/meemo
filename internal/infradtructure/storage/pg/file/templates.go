package file

// TODO: Очень часто повторяются джоины. Можно для улучшения производительности денормализовать таблицу
// и хранить вместо user_id email. Условно пользователь редко меняет email. Даже если захочет поменять
// это займет не много времени
const (
	SaveFileTemplate = `
	INSERT INTO file
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
	RETURNING id;`

	// пользователей может быть много. У каждого пользователя может быть 100+ файлов. Итоговая
	// временная таблица получиться большой. Поэтому сначала отсеиваем пользователей, а потом уже джоиним
	DeleteFileTemplate = `
	DELETE FROM file f
	INNER JOIN (SELECT * FROM users WHERE email = $1) u ON f.user_id = u.id
	WHERE f.original_name = $2;`

	GetFileTemplate = `
	SELECT * FROM file f
	INNER JOIN (SELECT * FROM users WHERE email = $1) u ON f.user_id = u.id
	WHERE f.original_name = $2;`

	ChangeVisibilityTemplate = `
	UPDATE file f
	SET f.is_public = $1
	INNER JOIN (SELECT * FROM users WHERE email = $2) u ON f.user_id = u.id
	WHERE f.original_name = $3;`

	SetStatusTemplate = `
	UPDATE file f
	SET f.status = $1 
	INNER JOIN (SELECT * FROM users WHERE email = $2) u ON f.user_id = u.id
	WHERE f.original_name = $3;
	RETURNING id;`
)
