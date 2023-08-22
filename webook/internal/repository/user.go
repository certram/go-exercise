package repository

import (
	"context"

	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"github.com/gin-gonic/gin"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE `email`=?
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Email:        u.Email,
		NickName:     u.NickName,
		Birthday:     u.Birthday,
		Introduction: u.Introduction,
		Location:     u.Location,
		Avatar:       u.Avatar,
	}, nil
	// 先从 cache 里面找
	// 再从 dao 里面找
	// 找到了回写 cache
}
func (r *UserRepository) Edit(ctx *gin.Context, u domain.User) error {
	return r.dao.Update(ctx, dao.User{
		Id:           u.Id,
		Birthday:     u.Birthday,
		NickName:     u.NickName,
		Location:     u.Location,
		Introduction: u.Introduction,
		Avatar:       u.Avatar,
	})
}
