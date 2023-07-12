package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
)

func (s *Service) SaveFileApi(ctx *gin.Context) {
	filepathName := ctx.PostForm("filepath")
	code := ctx.PostForm("code")

	dir := filepath.Dir(filepathName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create directory"})
			return
		}
	}

	err := os.WriteFile(filepathName, []byte(code), 0644)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save the file"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "File successfully saved"})
}
