package product

import (
	e "github.com/a-trium/domain-driven-design/implementation-1ambda/service-gateway/internal/exception"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type Repository interface {
	AddCategory(record *Category) (*Category, e.Exception)
	FindCategoryById(id uint) (*Category, e.Exception)

	AddImage(record *Image) (*Image, e.Exception)
	FindImageById(id uint) (*Image, e.Exception)

	AddProduct(record *Product) (*Product, e.Exception)
	FindProduct(id uint) (*Product, e.Exception)
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

type repositoryImpl struct {
	db *gorm.DB
}

func (r *repositoryImpl) AddCategory(record *Category) (*Category, e.Exception) {
	err := r.db.Create(record).Error

	if err != nil {
		wrap := errors.Wrap(err, "Failed to create Category")
		return nil, e.NewInternalServerException(wrap)
	}

	return record, nil
}

func (r *repositoryImpl) AddImage(record *Image) (*Image, e.Exception) {
	err := r.db.Create(record).Error

	if err != nil {
		wrap := errors.Wrap(err, "Failed to create Image")
		return nil, e.NewInternalServerException(wrap)
	}

	return record, nil
}

func (r *repositoryImpl) FindCategoryById(id uint) (*Category, e.Exception) {
	record := &Category{}
	err := r.db.Where("id = ?", id).First(record).Error

	if err != nil {
		wrap := errors.Wrap(err, "Failed to find Category by id")

		if gorm.IsRecordNotFoundError(err) {
			return nil, e.NewNotFoundException(wrap)
		}

		return nil, e.NewInternalServerException(wrap)
	}

	return record, nil
}

func (r *repositoryImpl) FindImageById(id uint) (*Image, e.Exception) {
	record := &Image{}
	err := r.db.Where("id = ?", id).First(record).Error

	if err != nil {
		wrap := errors.Wrap(err, "Failed to find Image by id")

		if gorm.IsRecordNotFoundError(err) {
			return nil, e.NewNotFoundException(wrap)
		}

		return nil, e.NewInternalServerException(wrap)
	}

	return record, nil
}

func (r *repositoryImpl) AddProduct(record *Product) (*Product, e.Exception) {
	err := r.db.Create(record).Error

	if err != nil {
		wrap := errors.Wrap(err, "Failed to create Product")
		return nil, e.NewInternalServerException(wrap)
	}

	return record, nil
}

func (r *repositoryImpl) FindProduct(id uint) (*Product, e.Exception) {
	record := &Product{}
	err := r.db.Where("id = ?", id).First(record).Error

	if err != nil {
		wrap := errors.Wrap(err, "Failed to find product by id")

		if gorm.IsRecordNotFoundError(err) {
			return nil, e.NewNotFoundException(wrap)
		}

		return nil, e.NewInternalServerException(wrap)
	}

	return record, nil
}
