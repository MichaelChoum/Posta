package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Follow struct {
	ID             int64 `gorm:"primary_key"`
	UserID         int64
	FollowedUserID int64
	FollowStatus   int
	CreateTime     time.Time
	UpdateTime     time.Time
}

func (m *Follow) TableName() string {
	return "follow"
}

type FollowModel struct {
	db *gorm.DB
}

func NewFollowModel(db *gorm.DB) *FollowModel {
	return &FollowModel{
		db: db,
	}
}

func (m *FollowModel) Insert(ctx context.Context, data *Follow) error {
	return m.db.WithContext(ctx).Create(data).Error
}

func (m *FollowModel) FindOne(ctx context.Context, id int64) (*Follow, error) {
	var result Follow
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&result).Error
	return &result, err
}

func (m *FollowModel) Update(ctx context.Context, data *Follow) error {
	return m.db.WithContext(ctx).Save(data).Error
}

func (m *FollowModel) UpdateFields(ctx context.Context, id int64, values map[string]interface{}) error {
	return m.db.WithContext(ctx).Model(&Follow{}).Where("id = ?", id).Updates(values).Error
}

func (m *FollowModel) FindByUserIDAndFollowedUserID(ctx context.Context, userId, followedUserId int64) (*Follow, error) {
	var result Follow
	err := m.db.WithContext(ctx).
		Where("user_id = ? AND followed_user_id = ?", userId, followedUserId).
		First(&result).Error
	// 如果是没找到的话，gorm是算成error的，但是FindByUserIDAndFollowedUserID不会返回error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &result, err
}

//
//	func (m *FollowModel) FindByUserId(ctx context.Context, userId int64, limit int) ([]*Follow, error) {
//		var result []*Follow
//		err := m.db.WithContext(ctx).
//			Where("user_id = ? AND follow_status = ?", userId, 1).
//			Order("id desc").
//			Limit(limit).
//			Find(&result).Error
//
//		return result, err
//	}

// 根据关注者的id找limit条关注记录
func (m *FollowModel) FindByUserId(ctx context.Context, userId int64, cursorId int64, limit int) ([]*Follow, error) {
	var result []*Follow
	query := m.db.WithContext(ctx).
		Where("user_id = ? AND follow_status = ?", userId, 1)

	if cursorId > 0 {
		query = query.Where("id < ?", cursorId)
	}

	err := query.Order("id desc").
		Limit(limit).
		Find(&result).Error

	return result, err
}

// 根据关注者的id和被关注者们的id找关注记录
func (m *FollowModel) FindByFollowedUserIds(ctx context.Context, userId int64, followedUserIds []int64) ([]*Follow, error) {
	var result []*Follow
	err := m.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Where("followed_user_id in (?)", followedUserIds).
		Find(&result).Error

	return result, err
}

func (m *FollowModel) FindByFollowedUserId(ctx context.Context, userId int64, limit int) ([]*Follow, error) {
	var result []*Follow
	err := m.db.WithContext(ctx).
		Where("followed_user_id = ? AND follow_status = ?", userId, 1).
		Order("id desc").
		Limit(limit).
		Find(&result).Error
	return result, err
}
