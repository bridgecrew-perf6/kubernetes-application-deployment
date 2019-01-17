package utils

import (
	"github.com/gin-gonic/gin"

	"kubernetes-services-deployment/types"
)

func StringPtr(input string) *string {
	return &input
}

func NewError(ctx *gin.Context, status int, err error) {
	er := types.HTTPError{
		Code:    status,
		Message: err.Error(),
	}
	ctx.JSON(status, er)
}
