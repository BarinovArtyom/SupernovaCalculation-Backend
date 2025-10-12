package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"lab1/internal/app/ds"

	"gorm.io/gorm"

	"github.com/minio/minio-go/v7"
)

func (r *Repository) GetScopes() ([]ds.Scope, error) {
	var scopes []ds.Scope
	err := r.db.Find(&scopes).Error
	if err != nil {
		return nil, err
	}
	if len(scopes) == 0 {
		return nil, fmt.Errorf("Массив пустой")
	}
	return scopes, nil
}

func (r *Repository) GetScopeByID(id int) (ds.Scope, error) {
	var scope ds.Scope
	err := r.db.First(&scope, id).Error
	if err != nil {
		return ds.Scope{}, err
	}
	return scope, nil
}

func (r *Repository) GetScopesByTitle(title string) ([]ds.Scope, error) {
	var scopes []ds.Scope
	err := r.db.Where("name ILIKE ?", "%"+title+"%").Find(&scopes).Error
	if err != nil {
		return nil, err
	}
	return scopes, nil
}

func (r *Repository) GetActiveStar(user_id int) (int, error) {
	var star ds.Star
	err := r.db.Where("status = ? and user_id = ?", "active", user_id).First(&star).Error
	if err != nil {
		return 0, err
	}
	return star.ID, nil
}

func (r *Repository) CalcsInStar(starID int) (int, error) {
	var count int64
	err := r.db.Model(&ds.Calc{}).Where("star_id = ?", starID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) AddToStar(scopeID int) error {
	var star ds.Star
	result := r.db.Where("status = ?", "active").First(&star)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			star = ds.Star{
				Status:        "active",
				CreationDate:  time.Now(),
				FormDate:      time.Now(),
				Name:          "New Star",
				Constellation: "Unknown",
				ModID:         1,
				UserID:        1,
			}

			if err := r.db.Create(&star).Error; err != nil {
				return fmt.Errorf("failed to create star: %w", err)
			}
		} else {
			return fmt.Errorf("failed to find star: %w", result.Error)
		}
	}

	var existingCalc ds.Calc
	result = r.db.Where("star_id = ? AND scope_id = ?", star.ID, scopeID).First(&existingCalc)

	if result.Error == nil {
		return nil
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing calc record: %w", result.Error)
	}

	calc := ds.Calc{
		ScopeID: scopeID,
		StarID:  star.ID,
	}

	if err := r.db.Create(&calc).Error; err != nil {
		return fmt.Errorf("failed to create calc record: %w", err)
	}

	return nil
}

func (r *Repository) CreateScope(scope *ds.Scope) (int, error) {
	err := r.db.Create(&scope).Error
	if err != nil {
		return 0, fmt.Errorf("error creating scope: %w", err)
	}
	return scope.ID, nil
}

func (r *Repository) EditScope(scope *ds.Scope) error {
	updates := make(map[string]interface{})

	if scope.Name != "" {
		updates["name"] = scope.Name
	}
	if scope.Description != "" {
		updates["descr"] = scope.Description
	}
	if scope.Filter != "" {
		updates["filter"] = scope.Filter
	}
	if scope.ImgLink != "" {
		updates["img_link"] = scope.ImgLink
	}
	if scope.Lambda != 0 {
		updates["lambda"] = scope.Lambda
	}
	if scope.DeltaLamb != 0 {
		updates["delta_lamb"] = scope.DeltaLamb
	}
	if scope.ZeroPoint != 0 {
		updates["zero_point"] = scope.ZeroPoint
	}
	updates["status"] = scope.Status

	err := r.db.Model(&ds.Scope{}).Where("id = ?", scope.ID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("error editing scope (%d): %w", scope.ID, err)
	}
	return nil
}

func (r *Repository) DeleteScope(id int) error {
	err := r.db.Delete(&ds.Scope{}, id).Error
	if err != nil {
		return fmt.Errorf("error deleting scope with id %s: %w", id, err)
	}
	return nil
}

func (r *Repository) DeletePicture(id int, img string) error {
	log.Printf("Deleting from MinIO - Bucket: vedro, Object: %s", img)

	err := r.minioclient.RemoveObject(context.Background(), "vedro", img, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("MinIO deletion error: %v", err)
		return fmt.Errorf("img deletion error: %w", err)
	}

	log.Printf("Successfully deleted from MinIO: %s", img)
	return nil
}

func (r *Repository) UploadPicture(id string, imageName string, imageFile io.Reader, imageSize int64) error {
	var scope ds.Scope

	if err := r.db.First(&scope, id).Error; err != nil {
		return fmt.Errorf("scope (%s) not found: %w", id, err)
	}

	if scope.ImgLink != "" {
		err := r.minioclient.RemoveObject(context.Background(), "vedro", imageName, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("img delete error %s: %v", scope.ImgLink, err)
		}
	}

	_, errMinio := r.minioclient.PutObject(context.Background(), "vedro", imageName, imageFile, imageSize, minio.PutObjectOptions{
		ContentType: "image/png",
	})

	scope.ImgLink = fmt.Sprintf("http://127.0.0.1:9000/vedro/%s.jpg", id)
	errDB := r.db.Save(&scope).Error

	if errMinio != nil || errDB != nil {
		return fmt.Errorf("img upload error for scope %s", id)
	}

	return nil
}
