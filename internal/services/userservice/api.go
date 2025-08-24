package userservice

import (
	"fmt"
	"net/http"

	"github.com/zrp9/launchl/internal/crane"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/request"
)

type UserAPI struct {
	s      UserService
	logger *crane.Zlogrus
}

func Initialize(s UserService, l *crane.Zlogrus) UserAPI {
	return UserAPI{
		s:      s,
		logger: l,
	}
}

func (u UserAPI) Name() string {
	return "user"
}

func (u UserAPI) RegisterRoutes(m *http.ServeMux) {
	// this is how i could have the main registerRoutes func call pass in prefixes
	//m.HandleFunc(fmt.Sprintf("GET /%v", prefix), u.HandleFetchUsers)
	m.HandleFunc("POST /user/subscribe", u.HandleAddUser)
	m.HandleFunc("GET /user/{email}", u.HandleGetUser)
	// get users number in queue
	m.HandleFunc("GET /user/{email}/position", u.HandleCheckQueue)
	m.HandleFunc("POST /user/{email}/survey", u.HandleSurvey)
	m.HandleFunc("POST /user/{email}/refered/{url-id}", u.HandleSubscribeRefered)
}

func (u UserAPI) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		var payload domain.User
		if err := request.ParseJSON(r, &payload); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteBadResponse(w, err)
			return
		}

		if err := u.s.validator.Struct(payload); err != nil {
			request.WriteBadResponse(w, err)
			return
		}

		role, err := u.s.cfgRepo.GetRole(r.Context(), "subscriber")
		if err != nil {
			u.logger.MustDebugErr(fmt.Errorf("could not find suscriber role %v", err))
			request.WriteServerError(w, err)
		}

		payload.Role = &role
		nUser, err := u.s.Create(r.Context(), &payload)

		if err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := request.JSON{
			"user": nUser,
		}

		if err = request.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (u UserAPI) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	request.SetJSONHeader(w)
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		id, err := request.ParseUUID(r)
		if err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		err = u.s.Delete(r.Context(), id)
		if err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := request.JSON{
			"success": true,
		}

		if err := request.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (u UserAPI) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	request.SetJSONHeader(w)
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		email, err := request.ParseEmail(r)
		if err != nil {
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		usr, err := u.s.GetByEmail(r.Context(), email)
		if err != nil {
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := request.JSON{
			"user": usr,
		}

		if err = request.WriteJSON(w, http.StatusOK, res); err != nil {
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (u UserAPI) HandleFetchUsers(w http.ResponseWriter, r *http.Request) {
	request.SetJSONHeader(w)
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		usrs, err := u.s.GetAll(r.Context())
		if err != nil {
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := request.JSON{
			"users": usrs,
		}
		if err := request.WriteJSON(w, http.StatusOK, res); err != nil {
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (u UserAPI) HandleCheckQueue(w http.ResponseWriter, r *http.Request) {
	request.SetJSONHeader(w)
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		usrs, err := u.s.GetAll(r.Context())
		if err != nil {
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := request.JSON{
			"users": usrs,
		}
		if err := request.WriteJSON(w, http.StatusOK, res); err != nil {
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}

}

func (u UserAPI) HandleSurvey(w http.ResponseWriter, r *http.Request) {
	request.SetJSONHeader(w)
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		usrs, err := u.s.GetAll(r.Context())
		if err != nil {
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := request.JSON{
			"users": usrs,
		}
		if err := request.WriteJSON(w, http.StatusOK, res); err != nil {
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}

}

func (u UserAPI) HandleSubscribeRefered(w http.ResponseWriter, r *http.Request) {
	request.SetJSONHeader(w)
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		usrs, err := u.s.GetAll(r.Context())
		if err != nil {
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := request.JSON{
			"users": usrs,
		}
		if err := request.WriteJSON(w, http.StatusOK, res); err != nil {
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}

}
