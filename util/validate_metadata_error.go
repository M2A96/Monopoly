package util

import (
	"fmt"
	"strings"
)

type (
	// ValidateMetadataErrorer is an interface.
	ValidateMetadataErrorer interface {
		error
		// GetField is a function.
		GetField() string
		// GetDescription is a function.
		GetDescription() string
	}

	// ValidateMetadataErrorers is an interface.
	ValidateMetadataErrorers interface {
		error
		ToArray() []ValidateMetadataErrorer
	}

	validateMetadataError struct {
		description string
		field       string
	}

	validateMetadataErrors []ValidateMetadataErrorer
)

var (
	_ ValidateMetadataErrorers = (*validateMetadataErrors)(nil)
	_ error                    = (*validateMetadataError)(nil)
	_ error                    = (*validateMetadataErrors)(nil)
)

// NewValidateMetadataError is a function.
func NewValidateMetadataError(
	description string,
	field string,
) *validateMetadataError {
	return &validateMetadataError{
		description: description,
		field:       field,
	}
}

// NewValidateMetadataErrors is a function.
func NewValidateMetadataErrors(
	errorers []ValidateMetadataErrorer,
) validateMetadataErrors {
	return validateMetadataErrors(errorers)
}

// GetDescription is a function.
func (util validateMetadataError) GetDescription() string {
	return util.description
}

// GetField is a function.
func (util validateMetadataError) GetField() string {
	return util.field
}

// Error is a function.
func (util validateMetadataError) Error() string {
	return fmt.Sprintf(
		`%s: %s`,
		util.GetField(),
		util.GetDescription(),
	)
}

// Error is a function.
func (util validateMetadataErrors) Error() string {
	errors := make([]string, 0, len(util))

	for _, value := range util {
		errors = append(errors, value.Error())
	}

	return strings.Join(errors, "; ")
}

// ToArray is a function.
func (util validateMetadataErrors) ToArray() []ValidateMetadataErrorer {
	return util
}
