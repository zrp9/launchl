// Package request has util methods for http handlers.
package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	EmlRgx             = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	PhneRgx            = regexp.MustCompile(`^\d{3}-\d{3}-\d{4}$/`)
	DefaultRecordLimit = 10
	maxSize            = int64(1024000)
	ErrMaxSize         = errors.New("image to large, max size")
	ErrReqTimeout      = errors.New("request timeout")
	ErrJSONParse       = errors.New("failed to parse json")
	ErrInvalidData     = errors.New("invalid data")
	ErrUnAuthorized    = errors.New("Unauthorized")
)

type JSON map[string]any

type Pager struct {
	Page  int
	Limit int
}

type AuthHeaderErr struct {
	url *url.URL
}

func NewAuthHeaderErr(url *url.URL) AuthHeaderErr {
	return AuthHeaderErr{url: url}
}

func (a AuthHeaderErr) Error() string {
	return fmt.Sprintf("request has missing authorization header, url: %v", a.url)
}

func WriteTimeoutResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusRequestTimeout)

	errMsg := map[string]any{
		"message": "request timeout",
		"status":  fmt.Sprintf("%v", http.StatusRequestTimeout),
		"success": false,
	}

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(errMsg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func ParseJSON(r *http.Request, payload any) error {
	log.Printf("parsing body: %v", r.Body)
	defer r.Body.Close() //nolint:errcheck

	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

func SetJSONHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func WriteResponse(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

func WriteJSON(w http.ResponseWriter, status int, msg any) error {
	SetJSONHeader(w)
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(msg)
}

func WriteErr(w http.ResponseWriter, status int, err error) {
	err = WriteJSON(w, status, map[string]string{"error": err.Error()})
	if err != nil {
		http.Error(w, err.Error(), status)
	}
}

func WriteManyErrors(w http.ResponseWriter, status int, errs []error) {
	e := make([]string, 0, len(errs))
	for _, err := range errs {
		e = append(e, err.Error())
	}

	err := WriteJSON(w, status, map[string]string{"errors": strings.Join(e, ",")})
	if err != nil {
		http.Error(w, err.Error(), status)
	}
}

func HandleTimeout(w http.ResponseWriter) {
	WriteErr(w, http.StatusRequestTimeout, ErrReqTimeout)
}

func FormatErrMsg(msg string, err error) string {
	return fmt.Sprintf(msg, err)
}

func IsValidEmail(email string) bool {
	return EmlRgx.MatchString(email)
}

func IsValidPhone(phne string) bool {
	return PhneRgx.MatchString(phne)
}

func CharCount(str string) int {
	return utf8.RuneCountInString(str)
}

func ConvertUUID(str string) uuid.UUID {
	if str == "" {
		log.Println("uuuid is empty")
		return uuid.Nil
	}

	uid, err := uuid.Parse(str)
	if err != nil {
		log.Println("failed to parse uuid setting to nil")
		return uuid.Nil
	}

	return uid
}

func DeterminRecordLimit(limt int) int {
	if limt <= 0 {
		return DefaultRecordLimit
	}

	return limt
}

func ParseFile(r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	err := r.ParseMultipartForm(maxSize)

	if err != nil {
		return nil, nil, ErrMaxSize
	}

	file, header, err := r.FormFile("image")

	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			return nil, nil, err
		} else {
			return nil, nil, nil
		}
	}

	return file, header, nil
}

func ParseFloatOrZero(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1, err
	}

	return f, nil
}

func ParseIntOrZero(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return -1, err
	}

	return i, nil
}

func ParseBool(s string) bool {
	return s == "true" || s == "on"
}

func ParseID(r *http.Request) (int, error) {
	id := r.PathValue("id")

	if id == "" {
		return 0, errors.New("failed to parse id from request")
	}

	ID, err := strconv.Atoi("id")
	if err != nil {
		return 0, err
	}

	return ID, nil
}

func ParseUUID(r *http.Request) (string, error) {
	uid := r.PathValue("id")

	if uid == "" {
		return "", errors.New("user id is required")
	}

	return uid, nil
}

func ParseEmail(r *http.Request) (string, error) {
	email := r.PathValue("email")
	if email == "" {
		return "", errors.New("email is required")
	}
	return email, nil
}

func ParseUsername(r *http.Request) (string, error) {
	usrname := r.PathValue("username")
	if usrname == "" {
		return "", errors.New("username is required")
	}

	return usrname, nil
}

func ParseURLID(r *http.Request) (string, error) {
	id := r.PathValue("urlId")
	if id == "" {
		return "", errors.New("url identifier is required")
	}
	return id, nil
}

func ParsePagenation(r *http.Request) (Pager, error) {
	query := r.URL.Query()
	page, err := strconv.Atoi(query.Get("page"))

	if err != nil {
		return Pager{}, fmt.Errorf("page was not included with request %v", err)
	}

	lmt, err := strconv.Atoi(query.Get("limit"))
	if err != nil {
		lmt = DeterminRecordLimit(0)
	}

	if lmt == 0 {
		lmt = DefaultRecordLimit
	}

	return Pager{Page: page, Limit: lmt}, nil
}

func WriteBadResponse(w http.ResponseWriter, err error) {
	WriteErr(w, http.StatusBadRequest, err)
}

func WriteServerError(w http.ResponseWriter, err error) {
	WriteErr(w, http.StatusInternalServerError, err)
}

// TODO: have something like this to make a dto handler for validating dtos?
// func ValidateDto(w http.ResponseWriter, dto DTO) error {
// 	if err := dto.Validate(); err != nil {
// 		WriteErr(w, http.StatusBadRequest, ErrInvalidData)
// 		return ErrInvalidData
// 	}
// 	return nil
