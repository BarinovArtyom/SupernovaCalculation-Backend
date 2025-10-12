package repository

import (
	"fmt"
	"lab1/internal/app/ds"
	"lab1/internal/app/dsn"
	"strconv"
)

func (r *Repository) GetUserByID(id string) (*ds.Users, error) {
	var user ds.Users
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	if err := r.db.First(&user, intID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUser(login string, pwd string) (*ds.Users, error) {
	var user ds.Users
	if login == "" || pwd == "" {
		return nil, fmt.Errorf("empty login/password")
	}
	if err := r.db.Where("login = ? and password = ?", login, pwd).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) CreateUser(user *ds.Users) error {
	// Устанавливаем ModStatus в false по умолчанию для новых пользователей
	user.ModStatus = false
	return r.db.Create(user).Error
}

func (r *Repository) UpdateUser(newUser ds.Users, id string) error {
	var user ds.Users
	intID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	if err := r.db.First(&user, intID).Error; err != nil {
		return fmt.Errorf("user %d not found", intID)
	}

	// Обновляем только если переданы данные
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

func (r *Repository) Login(login string, pwd string) (int, error) {
	i, err := dsn.GetCurrentUserID()
	if err == nil && i != "null" { // there's an active running session
		return 0, fmt.Errorf("an already running session exists: %s", i)
	}

	user, err := r.GetUser(login, pwd)
	if err != nil {
		return 0, err
	}

	strID := strconv.Itoa(user.ID)
	err = dsn.SetCurrentUserID(strID)
	if err != nil {
		return 0, fmt.Errorf("error starting session")
	}
	return user.ID, nil
}

func (r *Repository) Logout() error {
	i, err := dsn.GetCurrentUserID()
	if i == "null" || err != nil {
		return fmt.Errorf("no running session found")
	}
	if err := dsn.SetCurrentUserID("null"); err != nil {
		return fmt.Errorf("failed to deauth the user")
	}
	return nil
}

func (r *Repository) ModStatusCheck() (bool, error) {
	userID, err := dsn.GetCurrentUserID()
	if err != nil {
		return false, err
	}
	user, err := r.GetUserByID(userID)
	if err != nil {
		return false, err
	}
	return user.ModStatus, nil
}
