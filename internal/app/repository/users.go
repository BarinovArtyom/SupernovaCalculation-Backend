package repository

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"lab1/internal/app/ds"
	"lab1/internal/app/role"

	"gorm.io/gorm"
)

func (r *Repository) GetUserByID(id int) (*ds.Users, error) {
	var user ds.Users
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUser(newUser ds.Users, id int) error {
	var user ds.Users

	if err := r.db.First(&user, id).Error; err != nil {
		return fmt.Errorf("user %d not found", id)
	}

	if newUser.Login != "" {
		user.Login = newUser.Login
	}
	if newUser.Password != "" {
		user.Password = newUser.Password
	}

	if err := r.db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

func (r *Repository) ModStatusCheck(id int) (role.Role, error) {
	user, err := r.GetUserByID(id)
	if err != nil {
		return 0, err
	}
	return user.ModStatus, nil
}

func (r *Repository) GetUserByLogin(login string) (ds.Users, error) {
	var user ds.Users
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если пользователь не найден - это нормально, возвращаем пустого пользователя
			return ds.Users{}, nil
		}
		// Другие ошибки базы данных
		return ds.Users{}, err
	}
	return user, nil
}

func (r *Repository) RegisterUser(Users *ds.Users) error {
	if Users.Login == "" || Users.Password == "" {
		return fmt.Errorf("login and password are required")
	}

	candidate, err := r.GetUserByLogin(Users.Login)
	if err != nil {
		return err
	}

	// Проверяем, найден ли пользователь с таким логином
	if candidate.Login != "" && candidate.Login == Users.Login {
		return fmt.Errorf("user with such login already exists")
	}

	err = r.db.Table("users").Create(&Users).Error
	if err != nil {
		return fmt.Errorf("failed to create user: %s", err)
	}
	return nil
}

func (r *Repository) GenerateHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
