package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

// stubGameService is a minimal in-test stub of input.GameServicer.
type stubGameService struct {
	getFunc        func(ctx context.Context, id uuid.UUID) (bo.Gamer, error)
	listFunc       func(ctx context.Context, p dao.Paginationer, f dao.GameFilter) ([]bo.Gamer, dao.Cursorer, error)
	createGameFunc func(ctx context.Context, g bo.Gamer) (uuid.UUID, error)
	startGameFunc  func(ctx context.Context, id uuid.UUID) error
}

func (s *stubGameService) Get(ctx context.Context, id uuid.UUID) (bo.Gamer, error) {
	if s.getFunc != nil {
		return s.getFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (s *stubGameService) List(ctx context.Context, p dao.Paginationer, f dao.GameFilter) ([]bo.Gamer, dao.Cursorer, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, p, f)
	}
	return nil, nil, errors.New("not implemented")
}

func (s *stubGameService) CreateGame(ctx context.Context, g bo.Gamer) (uuid.UUID, error) {
	if s.createGameFunc != nil {
		return s.createGameFunc(ctx, g)
	}
	return uuid.Nil, errors.New("not implemented")
}

func (s *stubGameService) StartGame(ctx context.Context, id uuid.UUID) error {
	if s.startGameFunc != nil {
		return s.startGameFunc(ctx, id)
	}
	return errors.New("not implemented")
}

// gameHandlerIface combines the two exported handler interfaces for testing.
type gameHandlerIface interface {
	api.GameHandler
	api.Handler
}

func newTestGameHandler(svc *stubGameService) gameHandlerIface {
	tracer := noop.NewTracerProvider().Tracer("test")
	return api.NewGameHandler(
		nil,
		&noopLogger{},
		object.NewUUID(),
		tracer,
		api.WithGameHandlerGameServicer(svc),
	)
}

func TestCreateGame_Success(t *testing.T) {
	e := echo.New()
	newID := uuid.New()

	svc := &stubGameService{
		createGameFunc: func(_ context.Context, _ bo.Gamer) (uuid.UUID, error) {
			return newID, nil
		},
	}
	h := newTestGameHandler(svc)

	body := `{"name":"test-game","status":"waiting"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/games", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGame(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["id"])
}

func TestCreateGame_ServiceError(t *testing.T) {
	e := echo.New()

	svc := &stubGameService{
		createGameFunc: func(_ context.Context, _ bo.Gamer) (uuid.UUID, error) {
			return uuid.Nil, errors.New("db error")
		},
	}
	h := newTestGameHandler(svc)

	body := `{"name":"test-game","status":"waiting"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/games", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGame(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetGame_Success(t *testing.T) {
	e := echo.New()
	gameID := uuid.New()

	svc := &stubGameService{
		getFunc: func(_ context.Context, id uuid.UUID) (bo.Gamer, error) {
			return bo.NewGamer(id, "test-game", "waiting", uuid.Nil, uuid.Nil), nil
		},
	}
	h := newTestGameHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())

	err := h.GetGame(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetGame_InvalidID(t *testing.T) {
	e := echo.New()

	h := newTestGameHandler(&stubGameService{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-valid-id")

	err := h.GetGame(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetGame_NotFound(t *testing.T) {
	e := echo.New()
	gameID := uuid.New()

	svc := &stubGameService{
		getFunc: func(_ context.Context, _ uuid.UUID) (bo.Gamer, error) {
			return nil, errors.New("not found")
		},
	}
	h := newTestGameHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())

	err := h.GetGame(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestStartGame_Success(t *testing.T) {
	e := echo.New()
	gameID := uuid.New()

	svc := &stubGameService{
		startGameFunc: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	h := newTestGameHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())

	err := h.StartGame(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestStartGame_InvalidID(t *testing.T) {
	e := echo.New()

	h := newTestGameHandler(&stubGameService{})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-valid-id")

	err := h.StartGame(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGameHandler_RegisterRoutes(t *testing.T) {
	e := echo.New()
	h := newTestGameHandler(&stubGameService{})
	h.RegisterRoutes(e)

	paths := make(map[string]bool)
	for _, r := range e.Routes() {
		paths[r.Method+":"+r.Path] = true
	}

	assert.True(t, paths["GET:/api/v1/games/:id"], "should have GET game route")
	assert.True(t, paths["POST:/api/v1/games/:id/start"], "should have start game route")
	assert.True(t, paths["POST:/api/v1/games"], "should have create game route")
	assert.False(t, paths["GET:/api/v1/games/:id/state"], "old GetGameState route should be removed")
}
