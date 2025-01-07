package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "m102030m"
	DB_NAME     = "client"
	DB_HOST     = "localhost"
	DB_PORT     = "5432"
)

func dbConnect() (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		DB_USER, DB_PASSWORD, DB_NAME, DB_HOST, DB_PORT)
	return sql.Open("postgres", connStr)
}

func uploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file"})
		return
	}
	if err := c.SaveUploadedFile(file, "./uploads/"+file.Filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func getAllPlates(c *gin.Context) {
	db, err := dbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}
	defer db.Close()

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := `
		SELECT id, starttime, endtime, predicted_string, raw_image_path, plate_cropped_image_path, permit, camera_id 
		FROM plates 
		LIMIT $1 OFFSET $2;
	`
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plates"})
		return
	}
	defer rows.Close()

	var plates []map[string]interface{}
	for rows.Next() {
		var id, camera_id int
		var starttime, endtime, predicted_string, raw_image_path, plate_cropped_image_path sql.NullString
		var permit sql.NullBool

		if err := rows.Scan(&id, &starttime, &endtime, &predicted_string, &raw_image_path, &plate_cropped_image_path, &permit, &camera_id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to scan row: %v", err)})
			return
		}

		// Converting NULL values back to default Go types
		plate := map[string]interface{}{
			"id":                       id,
			"starttime":                starttime.String,
			"endtime":                  endtime.String,
			"predicted_string":         predicted_string.String,
			"raw_image_path":           raw_image_path.String,
			"plate_cropped_image_path": plate_cropped_image_path.String,
			"permit":                   permit.Bool,
			"camera_id":                camera_id,
		}
		plates = append(plates, plate)
	}

	c.JSON(http.StatusOK, gin.H{
		"page":   page,
		"limit":  limit,
		"plates": plates,
	})
}

func main() {
	r := gin.Default()
	r.POST("/upload", uploadImage)
	r.GET("/plates", getAllPlates)

	if err := r.Run("localhost:5001"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
