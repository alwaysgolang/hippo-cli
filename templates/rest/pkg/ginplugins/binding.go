package ginplugins

import (
	"github.com/gin-gonic/gin"

	customErrors "gotemplate/pkg/errors"
)

func MustBindJSON(ctx *gin.Context, model any) bool {
	if err := ctx.ShouldBindJSON(model); err != nil {
		WrapError(customErrors.WrapValidationError(err), ctx)
		return false
	}
	return true
}
