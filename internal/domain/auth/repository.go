package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const RedisBlacklistKeyPrefix = "blacklist:jti:"

type TokenBlacklistRepository interface {
	BlacklistToken(ctx context.Context, jti string, userID uuid.UUID, tokenType TokenType, expiresAt time.Time, reason string) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}

type APIKeyRepository interface {
	Create(ctx context.Context, key *APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*APIKey, error)
	GetByHash(ctx context.Context, keyHash string) (*APIKey, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*APIKey, error)
	Update(ctx context.Context, key *APIKey) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// GormTokenBlacklistRepository implements TokenBlacklistRepository with Redis cache and DB fallback
type GormTokenBlacklistRepository struct {
	redisClient *redis.Client
	db          *gorm.DB
}

func NewTokenBlacklistRepository(redisClient *redis.Client, db *gorm.DB) *GormTokenBlacklistRepository {
	return &GormTokenBlacklistRepository{
		redisClient: redisClient,
		db:          db,
	}
}

func (r *GormTokenBlacklistRepository) BlacklistToken(ctx context.Context, jti string, userID uuid.UUID, tokenType TokenType, expiresAt time.Time, reason string) error {
	if jti == "" {
		return nil
	}

	remaining := time.Until(expiresAt)
	if remaining <= 0 {
		return nil
	}

	if r.redisClient != nil {
		key := fmt.Sprintf("%s%s", RedisBlacklistKeyPrefix, jti)
		_ = r.redisClient.Set(ctx, key, "revoked", remaining).Err()
	}

	if r.db != nil {
		entry := BlacklistedToken{
			JTI:       jti,
			UserID:    userID,
			TokenType: string(tokenType),
			ExpiresAt: expiresAt,
			Reason:    reason,
		}
		_ = r.db.WithContext(ctx).Create(&entry).Error
	}

	return nil
}

func (r *GormTokenBlacklistRepository) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}

	if r.redisClient != nil {
		key := fmt.Sprintf("%s%s", RedisBlacklistKeyPrefix, jti)
		count, err := r.redisClient.Exists(ctx, key).Result()
		if err == nil {
			return count > 0, nil
		}
	}

	if r.db != nil {
		var blacklisted BlacklistedToken
		err := r.db.WithContext(ctx).Where("jti = ? AND expires_at > ?", jti, time.Now()).First(&blacklisted).Error
		if err == nil {
			return true, nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return false, nil
}

// GormAPIKeyRepository implements APIKeyRepository with PostgreSQL
type GormAPIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *GormAPIKeyRepository {
	return &GormAPIKeyRepository{db: db}
}

func (r *GormAPIKeyRepository) Create(ctx context.Context, key *APIKey) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *GormAPIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*APIKey, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var key APIKey
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *GormAPIKeyRepository) GetByHash(ctx context.Context, keyHash string) (*APIKey, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var key APIKey
	if err := r.db.WithContext(ctx).Where("key_hash = ?", keyHash).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *GormAPIKeyRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*APIKey, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}
	var keys []*APIKey
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

func (r *GormAPIKeyRepository) Update(ctx context.Context, key *APIKey) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Save(key).Error
}

func (r *GormAPIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&APIKey{}).Where("id = ?", id).Update("last_used_at", &now).Error
}

func (r *GormAPIKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&APIKey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     APIKeyStatusRevoked,
		"updated_at": now,
	}).Error
}

func (r *GormAPIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&APIKey{}).Error
}
