package rest

import (
	"log"
	"net/http"

	v "github.com/go-playground/validator/v10"
	"github.com/zrp9/launchl/internal/adapter/converter"
	"github.com/zrp9/launchl/internal/adapter/dto"
	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/domain/service"
	"github.com/zrp9/launchl/internal/request"
)

type UserHandler struct {
	s         service.UserService
	logger    crane.Zlogrus
	validator v.Validate
}

func NewUserHandler(s service.UserService, v v.Validate, l crane.Zlogrus) UserHandler {
	return UserHandler{
		s:         s,
		logger:    l,
		validator: v,
	}
}

func (u UserHandler) Name() string {
	return "user"
}

// I was going to originally seperate the services and have a user service and a survey service that had a similar package structure as launch
// but to keep the app simple, build it quicker and survey having limited publicly exposed functionality I decided to just create one service that excepts diferent repos
// basically I combined the user and survey service
// survey has only one endpoint expose so didn't see much of a point to seperate it out

func (u UserHandler) RegisterRoutes(m *http.ServeMux) {
	// this is how i could have the main registerRoutes func call pass in prefixes
	//m.HandleFunc(fmt.Sprintf("GET /%v", prefix), u.HandleFetchUsers)
	log.Printf("MUX in user hndler register:  %p", m)
	m.HandleFunc("POST /api/v1/user/subscribe", u.HandleLogging(u.HandleSubscribe))
	// get users number in queue
	m.HandleFunc("GET /api/v1/user/que/position/{email}", u.HandleLogging(u.HandleCheckQueue))
	m.HandleFunc("POST /api/v1/user/referred/{urlId}", u.HandleLogging(u.HandleSubscribeRefered))
	m.HandleFunc("GET /api/v1/user/{email}/rlink", u.HandleLogging(u.HandleGetReferLink))
	m.HandleFunc("POST /api/v1/user/{email}", u.HandleLogging(u.HandleDeleteUser))
}

func (u UserHandler) HandleLogging(hn APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := hn(w, r); err != nil {
			if e, ok := err.(APIErr); ok {
				request.WriteErr(w, e.Status, e)
			} else {
				request.WriteErr(w, http.StatusInternalServerError, err)
			}
			u.logger.MustError(err)
		}
	}
}

func (u UserHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) error {
	log.Printf("subscrib api")
	if err := r.Context().Err(); err != nil {
		return ReturnErr(http.StatusRequestTimeout, request.ErrReqTimeout)
	}

	var payload dto.UserCreateRequest
	if err := request.ParseJSON(r, &payload); err != nil {
		return ReturnErr(http.StatusBadRequest, err)
	}

	if err := u.validator.Struct(payload); err != nil {
		return ReturnErr(http.StatusBadRequest, err)
	}

	usr := converter.MakeCreateUser(payload)
	nUser, err := u.s.CreateUser(r.Context(), usr)

	if err != nil {
		return ReturnErr(http.StatusInternalServerError, err)
	}

	res := request.JSON{
		"user": nUser,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (u UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) error {
	log.Printf("del usr api")
	if err := r.Context().Err(); err != nil {
		return ReturnErr(http.StatusRequestTimeout, request.ErrReqTimeout)
	}

	email, err := request.ParseEmail(r)
	if err != nil {
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}

	err = u.s.DeleteUser(r.Context(), email)
	if err != nil {
		return APIErr{Status: http.StatusInternalServerError, Err: err}
	}

	res := request.JSON{
		"success": true,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (u UserHandler) HandleCheckQueue(w http.ResponseWriter, r *http.Request) error {
	log.Printf("check que api")
	if err := r.Context().Err(); err != nil {
		return APIErr{Status: http.StatusGatewayTimeout, Err: err}
	}

	email, err := request.ParseEmail(r)
	if err != nil {
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}

	position, err := u.s.CheckQuePosition(r.Context(), email)
	if err != nil {
		request.WriteErr(w, http.StatusBadRequest, err)
	}

	res := request.JSON{
		"quePosition": position,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (u UserHandler) HandleSubscribeRefered(w http.ResponseWriter, r *http.Request) error {
	log.Printf("subscribe refered api")
	if err := r.Context().Err(); err != nil {
		return APIErr{Status: http.StatusGatewayTimeout, Err: err}
	}

	var payload dto.UserCreateRequest

	urlID, err := request.ParseURLID(r)
	if err != nil {
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}

	if err = request.ParseJSON(r, &payload); err != nil {
		return APIErr{Status: http.StatusInternalServerError, Err: err}
	}

	usr := converter.MakeCreateUser(payload)
	if err = u.s.SignupReferal(r.Context(), *usr, urlID); err != nil {
		return APIErr{Status: http.StatusInternalServerError, Err: err}
	}

	res := request.JSON{
		"user": usr,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (u UserHandler) HandleGetReferLink(w http.ResponseWriter, r *http.Request) error {
	log.Printf("refer link api")
	if err := r.Context().Err(); err != nil {
		return APIErr{Status: http.StatusRequestTimeout, Err: err}
	}

	email, err := request.ParseEmail(r)
	if err != nil {
		return err
	}

	refLink, position, err := u.s.GetRefLinkAndPosition(r.Context(), email)
	if err != nil {
		return err
	}

	res := request.JSON{
		"refLink":     refLink,
		"curPosition": position,
	}
	return request.WriteJSON(w, http.StatusOK, res)
}
