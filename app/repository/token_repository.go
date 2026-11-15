package repository

import (
	"time"

	"kosku-api/app/model"

	"gorm.io/gorm"
)

/* ------------------------------------------------------------------ */
/*  Contract                                                          */
/* ------------------------------------------------------------------ */

type TokenRepository interface {
	Create(token *model.RefreshToken) error
	FindByHash(hash string) (*model.RefreshToken, error)
	RevokeByHash(hash string) error
	RevokeAllByUser(userID uint) error
}

/* ------------------------------------------------------------------ */
/*  Implementation                                                    */
/* ------------------------------------------------------------------ */

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *tokenRepository) FindByHash(hash string) (*model.RefreshToken, error) {
	var t model.RefreshToken
	err := r.db.Where("token_hash = ?", hash).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tokenRepository) RevokeByHash(hash string) error {
	now := time.Now()
	return r.db.Model(&model.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", now).Error
}

func (r *tokenRepository) RevokeAllByUser(userID uint) error {
	now := time.Now()
	return r.db.Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}
