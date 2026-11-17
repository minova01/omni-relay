package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	requestValidator = validator.New(validator.WithRequiredStructEnabled())
	usernamePattern  = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)
)

func init() {
	_ = requestValidator.RegisterValidation("username", func(field validator.FieldLevel) bool {
		return usernamePattern.MatchString(field.Field().String())
	})
	_ = requestValidator.RegisterValidation("password_bytes", func(field validator.FieldLevel) bool {
		length := len([]byte(field.Field().String()))
		return length >= 8 && length <= 72
	})
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, destination any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return requestValidator.Struct(destination)
}
