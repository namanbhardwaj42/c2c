// parser.go

package utils

import (
	models "c2c/internal/models"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func ParsingParams(c *gin.Context) (*models.CloudApiRequest, error) {
	var cloudApiRequest models.CloudApiRequest
	if err := c.ShouldBindBodyWith(&cloudApiRequest, binding.JSON); err != nil {
		return nil, err
	}
	var cc models.CustomContext
	if err := json.Unmarshal([]byte(cloudApiRequest.CustomContextFull), &cc); err != nil {
		return nil, err
	}

	cloudApiRequest.CustomContext = &cc

	return &cloudApiRequest, nil
}
