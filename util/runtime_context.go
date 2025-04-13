package util

import (
	"context"
	"encoding/json"

	"github/M2A96/Monopoly.git/object"

	"github.com/google/uuid"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"golang.org/x/text/language"
	"google.golang.org/grpc/metadata"
)

type (
	// RuntimeContexter is an interface.
	RuntimeContexter interface {
		// GetLanguageTag is a function.
		GetLanguageTag() language.Tag
		// GetMetadataMD is a function.
		GetMetadataMD() metadata.MD
		// GetUserID is a function.
		GetUserID() uuid.UUID
	}

	runtimeContext struct {
		languageTag language.Tag
		metadataMD  metadata.MD
		userID      uuid.UUID
	}
)

var (
	_ RuntimeContexter = (*runtimeContext)(nil)
	_ json.Marshaler   = (*runtimeContext)(nil)
)

// NewRuntimeContext is a function.
func NewRuntimeContext(
	ctx context.Context,
) *runtimeContext {
	values := grpc_ctxtags.Extract(ctx).Values()

	languageTag, ok := values[object.URIRuntimeContextLanguage].(language.Tag)
	if !ok {
		languageTag = language.Und
	}

	metadataMD, ok := values[object.URIRuntimeContextMetadata].(metadata.MD)
	if !ok {
		metadataMD = metadata.MD{}
	}

	userUUID, ok := values[object.URIRuntimeContextUserID].(uuid.UUID)
	if !ok {
		userUUID = uuid.Nil
	}

	runtimeContext := &runtimeContext{
		languageTag: languageTag,
		metadataMD:  metadataMD,
		userID:      userUUID,
	}

	return runtimeContext
}

// GetLanguageTag is a function.
func (util *runtimeContext) GetLanguageTag() language.Tag {
	return util.languageTag
}

// GetMetadataMD is a function.
func (util *runtimeContext) GetMetadataMD() metadata.MD {
	return util.metadataMD
}

// GetUserID is a function.
func (util *runtimeContext) GetUserID() uuid.UUID {
	return util.userID
}

// GetMap is a function.
func (util *runtimeContext) GetMap() map[string]any {
	return map[string]any{
		"metadata_md":  util.GetMetadataMD(),
		"language_tag": util.GetLanguageTag(),
		"user_id":      util.GetUserID(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (util *runtimeContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(util.GetMap())
}
