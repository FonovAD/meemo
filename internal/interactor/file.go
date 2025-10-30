package interactor

import (
	"meemo/internal/domain/file/repository"
	"meemo/internal/domain/file/service"
	storage "meemo/internal/infrastructure/storage/pg/file"
	"meemo/internal/infrastructure/storage/s3/file"
	handler "meemo/internal/presenter/http/handler/file"
	usecase "meemo/internal/usecase/file"
)

func (i *interactor) NewFileRepository() repository.FileRepository {
	return storage.NewFileRepository(i.conn)
}

func (i *interactor) NewFileService() service.FileService {
	return service.NewFileService()
}

func (i *interactor) NewS3Storage() file.S3Client {
	return file.NewS3Client(i.s3client, i.s3bucket)
}

func (i *interactor) NewFileUseCase() usecase.Usecase {
	return usecase.NewFileUsecase(i.NewFileRepository(), i.NewFileService(), i.NewS3Storage())
}

func (i *interactor) NewFileHandler() handler.FileHandler {
	return handler.NewFileHandler(i.NewFileUseCase())
}
