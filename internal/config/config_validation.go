package config

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var customValidators = map[string]validator.Func{
	"rsyncTargetHostValidation": rsyncTargetHostValidation,
}

func RegisterCustomValidations(validate *validator.Validate) {
	for tag, fn := range customValidators {
		err := validate.RegisterValidation(tag, fn)
		if err != nil {
			slog.Error("failed to register custom validation", "tag", tag, "error", err)
		}
	}
}

func rsyncTargetHostValidation(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // empty value is allowed
	}

	if !strings.HasSuffix(value, ":") {
		slog.Error("non-empty rsyncTargetHost must end with ':'", "value", value)
		return false
	}


	// match either `USER@HOST:` or `HOST:`
	pattern := `^([a-zA-Z0-9_-]+@)?(.+):$`
	re := regexp.MustCompile(pattern)
	matched := re.FindStringSubmatch(value)
	if matched == nil {
		slog.Error("rsyncTargetHost does not match required pattern `[USER@]HOST:`", "value", value)
		return false
	}
	host := matched[2]
	if host == "" {
		slog.Error("rsyncTargetHost must have a valid HOST", "value", value)
		return false
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Var(host, "hostname|ip")
	if err != nil {
		// additionally check for IPv6 in square brackets
		if !(strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]")) {
			slog.Error("HOST in rsyncTargetHost is not a valid hostname or IP address", "host", host, "value", value, "error", err)
			return false		
		}
		
		host = strings.TrimPrefix(host, "[")
		host = strings.TrimSuffix(host, "]")
		err = validate.Var(host, "ipv6")
		if err != nil {
			slog.Error("HOST in rsyncTargetHost is not a valid IPv6 address (must be enclosed in square brackets)", "host", host, "error", err)
			return false
		}
	}

	return true
}