// Package launch contains the api handlers, and service for the launch list apps main functionality I could have broken it up into user/survey services but survey has limitedd functionality
package launch

import (
	"errors"
	"net/http"
	"sync"

	"github.com/zrp9/launchl/internal/crane"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/dto"
	"github.com/zrp9/launchl/internal/request"
	"golang.org/x/sync/errgroup"
)

type LaunchAPI struct {
	s      LaunchService
	logger *crane.Zlogrus
}

func Initialize(s LaunchService, l *crane.Zlogrus) LaunchAPI {
	return LaunchAPI{
		s:      s,
		logger: l,
	}
}

func (u LaunchAPI) Name() string {
	return "user"
}

// I was going to originally seperate the services and have a user service and a survey service that had a similar package structure as launch
// but to keep the app simple, build it quicker and survey having limited publicly exposed functionality I decided to just create one service that excepts diferent repos
// basically I combined the user and survey service
// survey has only one endpoint expose so didn't see much of a point to seperate it out

func (u LaunchAPI) RegisterRoutes(m *http.ServeMux) {
	// this is how i could have the main registerRoutes func call pass in prefixes
	//m.HandleFunc(fmt.Sprintf("GET /%v", prefix), u.HandleFetchUsers)
	m.HandleFunc("POST /user/subscribe", u.HandleLogging(u.HandleSubscribe))
	m.HandleFunc("GET /user/{username}", u.HandleLogging(u.HandleGetUser))
	// get users number in queue
	m.HandleFunc("GET /user/{username}/position", u.HandleLogging(u.HandleCheckQueue))
	m.HandleFunc("POST /user/{username}/survey", u.HandleLogging(u.HandleSurvey))
	m.HandleFunc("POST /user/referred/{urlId}", u.HandleLogging(u.HandleSubscribeRefered))
}

type APIHandler func(w http.ResponseWriter, r *http.Request) error

type APIErr struct {
	Status int
	Err    error
}

func (a APIErr) Error() string {
	return a.Err.Error()
}

func (u LaunchAPI) HandleLogging(hn APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := hn(w, r); err != nil {
			if e, ok := err.(APIErr); ok {
				request.WriteErr(w, e.Status, e)
			}
			u.logger.MustError(err)
		}
	}
}

func (u LaunchAPI) HandleSubscribe(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return u.ReturnErr(http.StatusRequestTimeout, request.ErrReqTimeout)
	}

	var payload domain.User
	if err := request.ParseJSON(r, &payload); err != nil {
		return u.ReturnErr(http.StatusBadRequest, err)
	}

	if err := u.s.validator.Struct(payload); err != nil {
		return u.ReturnErr(http.StatusBadRequest, err)
	}

	role, err := u.s.cfgRepo.Get(r.Context(), "subscriber")
	if err != nil {
		return u.ReturnErr(http.StatusInternalServerError, err)
	}

	payload.Role = &role
	nUser, err := u.s.CreateUser(r.Context(), &payload)

	if err != nil {
		return u.ReturnErr(http.StatusInternalServerError, err)
	}

	res := request.JSON{
		"user": nUser,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (u LaunchAPI) HandleDeleteUser(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return u.ReturnErr(http.StatusRequestTimeout, request.ErrReqTimeout)
	}

	usrname, err := request.ParseUsername(r)
	if err != nil {
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}

	err = u.s.DeleteUserByUsername(r.Context(), usrname)
	if err != nil {
		return APIErr{Status: http.StatusInternalServerError, Err: err}
	}

	res := request.JSON{
		"success": true,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (u LaunchAPI) HandleGetUser(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return APIErr{
			Status: http.StatusGatewayTimeout,
			Err:    err,
		}
	}

	usrname, err := request.ParseUsername(r)
	if err != nil {
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}

	usr, err := u.s.GetUserByUsername(r.Context(), usrname)
	if err != nil {
		return APIErr{Status: http.StatusInternalServerError, Err: err}
	}

	res := request.JSON{
		"user": usr,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (u LaunchAPI) HandleFetchUsers(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return APIErr{Status: http.StatusGatewayTimeout, Err: err}
	}
	usrs, err := u.s.GetAllUsers(r.Context())
	if err != nil {
		return APIErr{Status: http.StatusInternalServerError, Err: err}
	}

	res := request.JSON{
		"users": usrs,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (u LaunchAPI) HandleCheckQueue(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return APIErr{Status: http.StatusGatewayTimeout, Err: err}
	}

	usrname, err := request.ParseUsername(r)
	if err != nil {
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}

	position, err := u.s.CheckQue(r.Context(), usrname)
	if err != nil {
		request.WriteErr(w, http.StatusBadRequest, err)
	}

	res := request.JSON{
		"quePosition": position,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (u LaunchAPI) HandleSurvey(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return APIErr{Status: http.StatusGatewayTimeout, Err: err}
	}
	var payload dto.SurveyResponses
	if err := request.ParseJSON(r, &payload); err != nil {
		request.WriteBadResponse(w, err)
	}

	if errs := payload.Validate(); len(errs) > 0 {
		return APIErr{Status: http.StatusBadRequest, Err: errors.Join(errs...)}
	}

	var wg sync.WaitGroup
	for _, anwser := range payload {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := u.s.AddSurveyResponses(r.Context(), anwser.UserID, anwser.QuestionID, anwser.OptionID, anwser.TextAnwser); err != nil {
				u.logger.MustError(err)
			}
		}()
	}

	wg.Wait()
	res := request.JSON{
		"status": 201,
	}

	return request.WriteJSON(w, http.StatusCreated, res)
}

func (u LaunchAPI) HandleSubscribeRefered(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return APIErr{Status: http.StatusGatewayTimeout, Err: err}
	}

	var payload domain.User

	usrname, err := request.ParseUsername(r)
	if err != nil {
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}

	urlID, err := request.ParseURLID(r)
	if err != nil {
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}

	if err = request.ParseJSON(r, &payload); err != nil {
		return APIErr{Status: http.StatusInternalServerError, Err: err}
	}

	usr, err := u.s.CreateUser(r.Context(), &payload)
	if err != nil {
		return APIErr{Status: http.StatusInternalServerError, Err: err}
	}

	referer, err := u.s.GetReferer(r.Context(), usrname, urlID)
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(r.Context())

	eg.Go(func() error {
		return u.s.RewardReferer(ctx, referer)
	})

	eg.Go(func() error {
		return u.s.CreateReferal(ctx, usr.ID, referer.ID)
	})

	if err := eg.Wait(); err != nil {
		return APIErr{Status: http.StatusInternalServerError, Err: err}
	}

	res := request.JSON{
		"users": "",
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (u LaunchAPI) ReturnErr(status int, err error) APIErr {
	return APIErr{
		Status: status,
		Err:    err,
	}
}
