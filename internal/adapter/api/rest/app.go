package rest

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/domain/service"
	"github.com/zrp9/launchl/internal/request"
)

type AppHandler struct {
	s         service.AppService
	logger    crane.Zlogrus
	validator validator.Validate
}

func NewAppHandler(s service.AppService, v validator.Validate, l crane.Zlogrus) AppHandler {
	return AppHandler{
		s:         s,
		logger:    l,
		validator: v,
	}
}

func (a AppHandler) HandleLogging(hn APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := hn(w, r); err != nil {
			if e, ok := err.(APIErr); ok {
				request.WriteErr(w, e.Status, e)
			}
			a.logger.MustError(err)
		}
	}
}

func (a AppHandler) RegisterRoutes(m *http.ServeMux) {
	m.HandleFunc("GET /api/v1/features/{pg}/{limit}", a.HandleLogging(a.HandleGetFeatures))
}

func (a AppHandler) HandleGetFeatures(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return ReturnErr(http.StatusRequestTimeout, request.ErrReqTimeout)
	}

	page, err := request.ParsePagenation(r)
	if err != nil {
		return err
	}

	if err = a.validator.Struct(page); err != nil {
		return err
	}

	features, err := a.s.GetFeatures(r.Context(), page.Number, page.Limit)
	if err != nil {
		return err
	}

	res := request.JSON{
		"features": features,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}
