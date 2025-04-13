package util

//go:generate mockgen -destination=../test/v2/validation.go -package=test -mock_names=Validationer=MockValidation . Validationer

import (
	"context"

	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"

	localesEN "github.com/go-playground/locales/en"
	universalTranslator "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	translationsEN "github.com/go-playground/validator/v10/translations/en"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/maps"
)

type (
	// GetValidationer is an interface.
	GetValidationer interface {
		// GetValidationer is a function.
		GetValidationer() Validationer
	}

	// Validationer is an interface.
	Validationer interface {
		// ValidateMetadata is a function.
		ValidateMetadata(
			context.Context,
			map[string]any,
		) error
	}

	validation struct {
		configConfigger  config.Configger
		logRuntimeLogger log.RuntimeLogger
		objectUUIDer     object.UUIDer
		traceTracer      trace.Tracer
	}
)

var (
	_ GetTracer            = (*validation)(nil)
	_ Validationer         = (*validation)(nil)
	_ config.GetConfigger  = (*validation)(nil)
	_ log.GetRuntimeLogger = (*validation)(nil)
	_ object.GetUUIDer     = (*validation)(nil)
)

// NewValidation is a function.
func NewValidation(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
) *validation {
	return &validation{
		configConfigger:  configConfigger,
		logRuntimeLogger: logRuntimeLogger,
		objectUUIDer:     objectUUIDer,
		traceTracer:      traceTracer,
	}
}

// GetTracer is a function.
func (util *validation) GetTracer() trace.Tracer {
	return util.traceTracer
}

// GetConfigger is a function.
func (util *validation) GetConfigger() config.Configger {
	return util.configConfigger
}

// GetRuntimeLogger is a function.
func (util *validation) GetRuntimeLogger() log.RuntimeLogger {
	return util.logRuntimeLogger
}

// GetUUIDer is a function.
func (util *validation) GetUUIDer() object.UUIDer {
	return util.objectUUIDer
}

// ValidateMetadata is a function.
func (util *validation) ValidateMetadata(
	ctx context.Context,
	data map[string]any,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = util.GetTracer().Start(
		ctx,
		"ValidateMetadata",
		trace.WithSpanKind(trace.SpanKindProducer),
	)
	defer traceSpan.End()

	runtimeContext := NewRuntimeContext(ctx)
	spanContext := NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "ValidateMetadata",
		"rt_ctx": runtimeContext,
		"sp_ctx": spanContext,
		"config": util.GetConfigger(),
		"data":   data,
	}

	util.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	dataVersion, ok := data["version"]
	if !ok {
		util.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, object.ErrDataVersion).
			Error(object.ErrDataVersion.Error())
		traceSpan.RecordError(object.ErrDataVersion)
		traceSpan.SetStatus(codes.Error, object.ErrDataVersion.Error())

		return object.ErrDataVersion
	}

	util.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDataVersion, dataVersion).
		Debug(object.URIEmpty)

	selectedValidateMapRules := make(map[string]any, 1)

	for key, validateMapRule := range util.GetConfigger().
		GetRuntimeConfigger().
		GetValidateMapRules()["rules"] {
		util.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIValidateMapRule, validateMapRule).
			Debug(object.URIEmpty)

		if validateMapRule["version"].(string) == dataVersion {
			util.GetRuntimeLogger().
				WithFields(fields).
				Debug(`validateMapRule["version"].(string) == dataVersion`)

			maps.Copy(selectedValidateMapRules, validateMapRule)

			break
		}
	}

	util.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldSelectedValidateMapRule, selectedValidateMapRules).
		Debug(object.URIEmpty)

	if len(selectedValidateMapRules) == 0 {
		util.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, object.ErrSelectedValidateMapRule).
			Error(object.ErrSelectedValidateMapRule.Error())
		traceSpan.RecordError(object.ErrSelectedValidateMapRule)
		traceSpan.SetStatus(codes.Error, object.ErrSelectedValidateMapRule.Error())

		return object.ErrSelectedValidateMapRule
	}

	delete(selectedValidateMapRules, "version")

	util.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldSelectedValidateMapRule, selectedValidateMapRules).
		Debug(object.URIEmpty)

	validatorValidate := validator.New()
	localesTranslator := localesEN.New()
	utUniversalTranslator := universalTranslator.New(localesTranslator)
	universalTranslatorTranslator, _ := utUniversalTranslator.GetTranslator("en")

	if err := translationsEN.RegisterDefaultTranslations(
		validatorValidate,
		universalTranslatorTranslator,
	); err != nil {
		util.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrRegisterDefaultTranslations.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrRegisterDefaultTranslations.Error())

		return err
	}

	errValidateMap := validatorValidate.ValidateMap(data, selectedValidateMapRules)

	util.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldErrValidateMap, errValidateMap).
		Debug(object.URIEmpty)

	validateMetadatas := make([]ValidateMetadataErrorer, 0)

	for key, value := range errValidateMap {
		util.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldValue, value).
			Debug(object.URIEmpty)

		validateMetadataError := NewValidateMetadataError(
			value.(validator.ValidationErrors).Translate(universalTranslatorTranslator)[object.URIEmpty],
			key,
		)

		util.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldValidateMetadataError, validateMetadataError).
			Debug(object.URIEmpty)

		validateMetadatas = append(validateMetadatas, validateMetadataError)
	}

	util.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldValidateMetadatas, validateMetadatas).
		Debug(object.URIEmpty)

	newValidateMetadataErrors := NewValidateMetadataErrors(validateMetadatas)

	util.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldValidateMetadataErrors, newValidateMetadataErrors).
		Debug(object.URIEmpty)

	if len(newValidateMetadataErrors.ToArray()) > 0 {
		util.GetRuntimeLogger().
			WithFields(fields).
			Debug(`len(validateMetadataErrors.ToArray()) > 0`)

		return newValidateMetadataErrors
	}

	return nil
}
