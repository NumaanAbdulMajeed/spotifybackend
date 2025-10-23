package models

import (
	"time"

	// "gorm.io/driver/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID           string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	DisplayName  string    `json:"display_name"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Artist struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Album struct {
	ID          string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ArtistID    string     `gorm:"type:uuid" json:"artist_id"`
	Title       string     `gorm:"not null" json:"title"`
	ReleaseDate *time.Time `json:"release_date"`
	CoverURL    string     `json:"cover_url"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

type Track struct {
	ID          string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AlbumID     string    `gorm:"type:uuid" json:"album_id"`
	ArtistID    string    `gorm:"type:uuid" json:"artist_id"`
	Title       string    `gorm:"not null" json:"title"`
	DurationSec int       `json:"duration_seconds"`
	AudioKey    string    `gorm:"not null" json:"audio_key"` // object key
	MimeType    string    `json:"mime_type"`
	BitrateKbps int       `json:"bitrate_kbps"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type Play struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid" json:"user_id"`
	TrackID   string    `gorm:"type:uuid" json:"track_id"`
	PlayedAt  time.Time `gorm:"autoCreateTime" json:"played_at"`
	Position  int       `json:"position_seconds"`
	UserAgent string    `json:"user_agent"`
}

func NewGormDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Artist{}, &Album{}, &Track{}, &Play{})
}
