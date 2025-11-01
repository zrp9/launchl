package rest

import (
	"fmt"
	"log"
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
				log.Printf("err %v", err)
				request.WriteErr(w, e.Status, e)
			}
			log.Printf("err %v", err)
			a.logger.MustError(err)
		}
	}
}

func (a AppHandler) RegisterRoutes(m *http.ServeMux) {
	log.Printf("MUX in app hndler register:  %p", m)
	m.HandleFunc("GET /api/v1/features", a.HandleLogging(a.HandleGetFeatures))
	m.HandleFunc("GET /api/v1/roles", a.HandleLogging(a.HandleGetRoles))
	m.HandleFunc("GET /api/v1/testimonials", a.HandleLogging(a.HandleGetTestimonials))
}

func (a AppHandler) HandleGetFeatures(w http.ResponseWriter, r *http.Request) error {
	log.Printf("hit features api")
	if err := r.Context().Err(); err != nil {
		return ReturnErr(http.StatusRequestTimeout, request.ErrReqTimeout)
	}

	page, err := request.ParsePagenation(r)
	log.Printf("page info %v %v", page.Number, page.Limit)
	if err != nil {
		return ReturnErr(http.StatusBadRequest, err)
	}

	if err = a.validator.Struct(page); err != nil {
		return ReturnErr(http.StatusBadRequest, fmt.Errorf("validation error %w", err))
	}

	features, err := a.s.GetFeatures(r.Context(), page.Number, page.Limit)
	if err != nil {
		return ReturnErr(http.StatusInternalServerError, fmt.Errorf("service error %w", err))
	}

	res := request.JSON{
		"feats": features,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (a AppHandler) HandleGetRoles(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return ReturnErr(http.StatusRequestTimeout, request.ErrReqTimeout)
	}

	roles, err := a.s.GetRoles(r.Context())
	if err != nil {
		log.Printf("get roles err %v", err)
		return ReturnErr(http.StatusInternalServerError, err)
	}

	res := request.JSON{
		"roles": roles,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (a AppHandler) HandleGetTestimonials(w http.ResponseWriter, r *http.Request) error {
	log.Printf("hit testimonials")
	if err := r.Context().Err(); err != nil {
		return ReturnErr(http.StatusRequestTimeout, request.ErrReqTimeout)
	}

	testimonials, err := a.s.GetTestimonials(r.Context())
	if err != nil {
		return ReturnErr(http.StatusInternalServerError, err)
	}

	res := request.JSON{
		"testimonials": testimonials,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}
