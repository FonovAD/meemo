package interactor

import (
	"time"

	tokenservice "meemo/internal/domain/token/service"
	"meemo/internal/domain/user/repository"
	userservice "meemo/internal/domain/user/service"
	storage "meemo/internal/infrastructure/storage/pg/user"
	handler "meemo/internal/presenter/http/handler/user"
	usecase "meemo/internal/usecase/user"
)

func (i *interactor) NewUserRepository() repository.UserRepository {
	return storage.NewUserRepository(i.conn)
}

func (i *interactor) NewUserService() userservice.UserService {
	return userservice.NewUserService()
}

func (i *interactor) NewJWTTokenService() tokenservice.TokenService {
	// TODO: Вынести в конфигурацию
	secretKey := "your-256-bit-secret-key-change-in-production"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour // 7 дней

	return tokenservice.NewJWTTokenService(secretKey, accessExpiry, refreshExpiry)
}

func (i *interactor) NewUserUseCase() usecase.UseCase {
	return usecase.NewUseCase(
		i.NewUserRepository(),
		i.NewUserService(),
		i.NewJWTTokenService(),
	)
}

func (i *interactor) NewUserHandler() handler.UserHandler {
	return handler.NewUserHandler(i.NewUserUseCase(), i.registrationEnabled)
}
