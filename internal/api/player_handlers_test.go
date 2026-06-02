package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github/M2A96/Monopoly.git/internal/api"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
)

// stubPlayerService is a minimal in-test stub of input.PlayerServicer.
type stubPlayerService struct {
	getFunc          func(ctx context.Context, id uuid.UUID) (bo.Player, error)
	listFunc         func(ctx context.Context, p dao.Paginationer, f dao.PlayerFilter) ([]bo.Player, dao.Cursorer, error)
	createPlayerFunc func(ctx context.Context, player bo.Player) (uuid.UUID, error)
}

func (s *stubPlayerService) Get(ctx context.Context, id uuid.UUID) (bo.Player, error) {
	if s.getFunc != nil {
		return s.getFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (s *stubPlayerService) List(ctx context.Context, p dao.Paginationer, f dao.PlayerFilter) ([]bo.Player, dao.Cursorer, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, p, f)
	}
	return nil, nil, errors.New("not implemented")
}

func (s *stubPlayerService) CreatePlayer(ctx context.Context, player bo.Player) (uuid.UUID, error) {
	if s.createPlayerFunc != nil {
		return s.createPlayerFunc(ctx, player)
	}
	return uuid.Nil, errors.New("not implemented")
}

type playerHandlerIface interface {
	api.PlayerHandler
	api.Handler
}

func newTestPlayerHandler(svc *stubPlayerService) playerHandlerIface {
	tracer := noop.NewTracerProvider().Tracer("test")
	return api.NewPlayerHandler(
		nil,
		&noopLogger{},
		object.NewUUID(),
		tracer,
		api.WithPlayerHandlerPlayerServicer(svc),
	)
}

func TestGetPlayer_Success(t *testing.T) {
	e := echo.New()
	playerID := uuid.New()

	svc := &stubPlayerService{
		getFunc: func(_ context.Context, id uuid.UUID) (bo.Player, error) {
			return bo.NewPlayer(id, uuid.Nil, "alice", 1500, 0, false, 0, false), nil
		},
	}
	h := newTestPlayerHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(playerID.String())

	err := h.GetPlayer(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetPlayer_InvalidID(t *testing.T) {
	e := echo.New()

	h := newTestPlayerHandler(&stubPlayerService{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.GetPlayer(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetPlayer_NotFound(t *testing.T) {
	e := echo.New()
	playerID := uuid.New()

	svc := &stubPlayerService{
		getFunc: func(_ context.Context, _ uuid.UUID) (bo.Player, error) {
			return nil, errors.New("not found")
		},
	}
	h := newTestPlayerHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(playerID.String())

	err := h.GetPlayer(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPlayerHandler_RegisterRoutes(t *testing.T) {
	e := echo.New()
	h := newTestPlayerHandler(&stubPlayerService{})
	h.RegisterRoutes(e)

	paths := make(map[string]bool)
	for _, r := range e.Routes() {
		paths[r.Method+":"+r.Path] = true
	}

	assert.True(t, paths["POST:/api/v1/players"], "should have create player route")
	assert.True(t, paths["GET:/api/v1/players/:id"], "should have get player route")
	assert.True(t, paths["GET:/api/v1/games/:game_id/players"], "should have list players route")
}
