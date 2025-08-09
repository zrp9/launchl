package userservice

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/zrp9/launchl/internal/auth"
	"github.com/zrp9/launchl/internal/crane"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/dto"
	"github.com/zrp9/launchl/internal/repos"
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
	m.HandleFunc("GET /user", u.HandleGetUser)
	m.HandleFunc("GET /user/{id}", u.HandleAddUser)
	m.HandleFunc("PUT /user/{id}", u.HandleEditUser)
	m.HandleFunc("PATCH /user/{id}", u.HandleEditUser)
	m.HandleFunc("DELTE /user/{id}", u.HandleDeleteUser)
	m.HandleFunc("PUT /sign-up/{role}", u.HandleSignup)
	m.HandleFunc("PUT /sign-in", u.HandleLogin)
	m.HandleFunc("GET /whoami", u.HandleWhoAmI)
}

func (u UserAPI) HandleLogin(w http.ResponseWriter, r *http.Request) {
	request.SetJSONHeader(w)
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		var payload dto.LoginDto
		if err := request.ParseJSON(r, &payload); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := u.s.validator.Struct(payload); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		authenticated, user, err := u.s.Authenticate(r.Context(), payload.Username, payload.Password)

		if err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusUnauthorized, request.ErrUnAuthorized)
			return
		}

		if !authenticated {
			request.WriteErr(w, http.StatusUnauthorized, request.ErrUnAuthorized)
			return
		}

		token, err := auth.GenerateToken(user.ID.String(), user.Username, user.Role.Name)
		if err != nil {
			u.logger.MustDebugErr(fmt.Errorf("error generating token %v", err))
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		// for session based
		// http.SetCookie(w, &http.Cookie{
		// 	Name: "token",
		// 	Value: token,
		// 	Expires: auth.Expirey,
		// 	HttpOnly: true,
		// 	Secure: true,
		// 	SameSite: http.SameSiteStrictMode,
		// })

		// TODO: use refresh tokens also
		res := request.JSON{
			"token": token,
			"user":  user,
		}
		if err = request.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (u UserAPI) HandleSignup(w http.ResponseWriter, r *http.Request) {
	request.SetJSONHeader(w)
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		var payload dto.SignupDto
		if err := request.ParseJSON(r, &payload); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusBadRequest, err)
		}

		if err := u.s.validator.Struct(payload); err != nil {
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		roleName := r.PathValue("role")
		if roleName == "" {
			u.logger.MustDebugErr(errors.New("signup request did not have a specified role"))
		}

		role, err := u.s.cfgRepo.GetRole(r.Context(), roleName)
		if err != nil {
			u.logger.MustDebugErr(fmt.Errorf("role %v does not exist %w", role, err))
			request.WriteErr(w, http.StatusBadRequest, err)
		}

		usr := u.s.SignupAdapter(payload)
		usr.Role = &role
		nUser, err := u.s.Create(r.Context(), usr)

		if err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		// TODO: use refresh tokens also
		token, err := auth.GenerateToken(nUser.ID.String(), nUser.Username, nUser.Role.Name)
		if err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := request.JSON{
			"token": token,
			"user":  nUser,
		}

		if err = request.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (u UserAPI) HandleWhoAmI(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			err := request.NewAuthHeaderErr(r.URL)
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusUnauthorized, request.ErrUnAuthorized)
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			err := errors.New("token was not on header")
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusUnauthorized, request.ErrUnAuthorized)
			return
		}

		token := tokenParts[1]
		usr, err := u.s.ValidateClaims(r.Context(), token)

		if err != nil {
			request.WriteErr(w, http.StatusUnauthorized, request.ErrUnAuthorized)
		}

		res := request.JSON{
			"user": usr,
		}

		if err := request.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (u UserAPI) HandleAddUser(w http.ResponseWriter, r *http.Request) {
	request.SetJSONHeader(w)
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		var payload domain.User
		if err := request.ParseJSON(r, &payload); err != nil {
			u.logger.MustDebugErr(errors.Join(request.ErrJSONParse, err))
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		// TODO: make a dto type and call request.ValidateDto
		// to validate diferent dtos generically
		usr, err := u.s.Create(r.Context(), &payload)
		if err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := request.JSON{
			"user": usr,
		}

		if err := request.WriteJSON(w, http.StatusOK, res); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (u UserAPI) HandleEditUser(w http.ResponseWriter, r *http.Request) {
	request.SetJSONHeader(w)
	select {
	case <-r.Context().Done():
		u.logger.MustDebugErr(request.ErrReqTimeout)
		request.HandleTimeout(w)
	default:
		var payload domain.User
		if err := request.ParseJSON(r, &payload); err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		usr, err := u.s.Update(r.Context(), payload)
		if err != nil {
			u.logger.MustDebugErr(err)
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := request.JSON{
			"user": usr,
		}
		if err := request.WriteJSON(w, http.StatusOK, res); err != nil {
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
		id, err := request.ParseUUID(r)
		if err != nil {
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		usr, err := u.s.Get(r.Context(), id)
		if err != nil {
			if err != repos.ErrNoRecords {
				request.WriteErr(w, http.StatusInternalServerError, err)
				return
			}
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
