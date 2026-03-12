package main

import (
	"errors"
	"fmt"
	"net/http"

	"greenlight.haodm.net/internal/data"
	"greenlight.haodm.net/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	validator := validator.New()
	if data.ValidateUser(validator, user); !validator.Valid() {
		app.failedValidationResponse(w, r, validator.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		fmt.Println(err.Error())
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			validator.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(w, r, validator.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Set response header so that client know where to find the created movie
	header := make(http.Header)
	header.Set("Location", fmt.Sprintf("/v1/users/%d", user.ID))

	// Write and return http response to client
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, header)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
