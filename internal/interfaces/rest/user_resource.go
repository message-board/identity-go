package rest

import (
	"net/http"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/go-chi/render"
	"github.com/message-board/identity-go/internal/interfaces/handlers/createuser"
	"github.com/message-board/identity-go/pkg/requests"
)

type UserResource struct {
	commandBus *cqrs.CommandBus
}

func NewUserResource(commandBus *cqrs.CommandBus) UserResource {
	return UserResource{
		commandBus: commandBus,
	}
}

// ListUsers godoc
// @Summary List users
// @Description get users
// @Tags users
// @Accept  json
// @Produce  json
// @Success 200 {array} rest.Users
// @Failure 400 {object} rest.Error
// @Failure 404 {object} rest.Error
// @Failure 500 {object} rest.Error
// @Router /api/users [get]
func (ur UserResource) GetUsers(w http.ResponseWriter, r *http.Request) {
	users := Users{
		Users: []User{},
	}

	render.Respond(w, r, users)
}

// GetUser godoc
// @Summary Get user
// @Description get user by id
// @Tags users
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Success 200 {object} rest.User
// @Failure 400 {object} rest.Error
// @Failure 404 {object} rest.Error
// @Failure 500 {object} rest.Error
// @Router /api/users/{id} [get]
func (ur UserResource) GetUser(w http.ResponseWriter, r *http.Request, id string) {
	user := User{
		Id:           id,
		EmailAddress: id + "@test.com",
	}

	render.Respond(w, r, user)
}

// CreateUser godoc
// @Summary Create user
// @Description create user
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body requests.CreateUserRequest true "Create user"
// @Success 200 {object} rest.User
// @Failure 400 {object} rest.Error
// @Failure 404 {object} rest.Error
// @Failure 500 {object} rest.Error
// @Router /api/users [post]
func (ur UserResource) CreateUser(w http.ResponseWriter, r *http.Request) {
	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		render.Render(w, r, ErrUnsupportedMediaType())
		return
	}

	request := &requests.CreateUserRequest{}
	if err := render.Decode(r, request); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	createUser := createuser.NewCreateUser(request.Id, request.EmailAddress, request.Password)
	if err := ur.commandBus.Send(r.Context(), createUser); err != nil {
		panic(err)
	}
	// err := h.app.Commands.CreateUserCommandHandler.Handle(r.Context(), command)
	// if err != nil {
	// 	util.WriteResponse(w, "Failed to create user "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	w.WriteHeader(http.StatusNoContent)
}

//--
// Error response payloads & renderers
//--

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrUnsupportedMediaType() render.Renderer {
	return &ErrResponse{
		Err:            nil,
		HTTPStatusCode: 415,
		StatusText:     "Unsupported Media Type",
		ErrorText:      "Unsupported Media Type",
	}
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}
