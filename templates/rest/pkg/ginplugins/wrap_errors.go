package ginplugins

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	customErrors "gotemplate/pkg/errors"
)

func WrapError(err error, ginContext *gin.Context) {
	statusCode := http.StatusInternalServerError
	errType := customErrors.ErrSystem
	switch {
	case errors.Is(err, customErrors.ErrDataNotFound):
		errType = customErrors.ErrDataNotFound
		statusCode = http.StatusNotFound
	case errors.Is(err, customErrors.ErrValidation):
		errType = customErrors.ErrValidation
		statusCode = http.StatusBadRequest
	case errors.Is(err, customErrors.ErrExternalService):
		errType = customErrors.ErrExternalService
		statusCode = http.StatusExpectationFailed
	case errors.Is(err, customErrors.ErrSystem):
		errType = customErrors.ErrSystem
	case errors.Is(err, customErrors.ErrPermissionDenied):
		errType = customErrors.ErrPermissionDenied
		statusCode = http.StatusForbidden
	}
	ginContext.Writer.Header().Set("X-Error-Type", errType.Error())
	ginContext.JSON(statusCode, gin.H{"error": err.Error()})
}
