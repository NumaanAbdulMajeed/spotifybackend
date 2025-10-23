package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"spotifybackend/internal/config"
	"spotifybackend/internal/models"
	"spotifybackend/internal/storage"
)

type createTrackReq struct {
	Title       string `json:"title" binding:"required"`
	AlbumID     string `json:"album_id"`
	ArtistID    string `json:"artist_id"`
	DurationSec int    `json:"duration_seconds"`
	MimeType    string `json:"mime_type" binding:"required"`
	BitrateKbps int    `json:"bitrate_kbps"`
}

func RegisterTrackRoutes(rg *gin.RouterGroup, db *gorm.DB, s3client *storage.S3Client, cfg *config.Config) {
	tracks := rg.Group("/tracks")
	{
		tracks.POST("", func(c *gin.Context) {
			// create a track metadata record and return server upload endpoint
			var req createTrackReq
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			// create DB record with temporary audio_key
			key := "uploads/" + time.Now().Format("20060102-150405-000") + "-" + req.Title

			// Replace db.Create to avoid inserting empty strings into UUID columns:
			var newID string
			var albumVal interface{} = nil
			var artistVal interface{} = nil
			if req.AlbumID != "" {
				albumVal = req.AlbumID
			}
			if req.ArtistID != "" {
				artistVal = req.ArtistID
			}

			// raw insert returning id so empty album/artist become NULL
			if err := db.Raw(
				`INSERT INTO "tracks" ("album_id","artist_id","title","duration_sec","audio_key","mime_type","bitrate_kbps","created_at")
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING "id"`,
				albumVal, artistVal, req.Title, req.DurationSec, key, req.MimeType, req.BitrateKbps, time.Now(),
			).Scan(&newID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db create: " + err.Error()})
				return
			}

			var track models.Track
			if err := db.First(&track, "id = ?", newID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db fetch after create: " + err.Error()})
				return
			}

			// Return local upload endpoint instead of calling s3client (avoid nil deref)
			uploadURL := fmt.Sprintf("/api/v1/tracks/%v/upload", track.ID)
			c.JSON(http.StatusCreated, gin.H{"track": track, "upload_url": uploadURL})
		})

		tracks.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			var t models.Track
			if err := db.First(&t, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"track": t})
		})

		// upload file to local ./uploads directory
		tracks.POST("/:id/upload", func(c *gin.Context) {
			id := c.Param("id")
			var t models.Track
			if err := db.First(&t, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
				return
			}

			file, header, err := c.Request.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "file required: " + err.Error()})
				return
			}
			defer file.Close()

			uploadDir := "uploads"
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir: " + err.Error()})
				return
			}

			safeName := filepath.Base(header.Filename)
			destName := fmt.Sprintf("%v-%s-%s", t.ID, time.Now().Format("20060102-150405"), safeName)
			destPath := filepath.Join(uploadDir, destName)

			out, err := os.Create(destPath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "create file: " + err.Error()})
				return
			}
			defer out.Close()

			if _, err := io.Copy(out, file); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "save file: " + err.Error()})
				return
			}

			// update DB record with local path
			if err := db.Model(&t).Update("audio_key", destPath).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db update: " + err.Error()})
				return
			}
			// refresh
			if err := db.First(&t, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db fetch after update: " + err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{"track": t, "path": destPath})
		})

		// stream (serve file from local disk)
		tracks.GET("/:id/stream", func(c *gin.Context) {
			id := c.Param("id")
			var t models.Track
			if err := db.First(&t, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			if t.AudioKey == "" {
				c.JSON(http.StatusNotFound, gin.H{"error": "no audio uploaded"})
				return
			}
			c.File(t.AudioKey)
		})

		// record plays
		tracks.POST("/plays", func(c *gin.Context) {
			var p models.Play
			if err := c.ShouldBindJSON(&p); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			p.PlayedAt = time.Now()
			p.UserAgent = c.GetHeader("User-Agent")
			if err := db.Create(&p).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"play": p})
		})
	}
}
