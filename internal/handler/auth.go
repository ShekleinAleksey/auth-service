package handler

import (
	"net/http"

	"github.com/ShekleinAleksey/auth-service/internal/entity"
	"github.com/gin-gonic/gin"
)

type signUpInput struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// @Summary SignUp
// @Tags auth
// @Description create account
// @ID create-account
// @Accept  json
// @Produce  json
// @Param input body signUpInput true "account info"
// @Success 200 {object} TokenDetails
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/sign-up [post]
func (h *Handler) signUp(c *gin.Context) {
	var input signUpInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if input.Password == "" {
		newErrorResponse(c, http.StatusBadRequest, "password is required")
		return
	}

	if input.Email == "" {
		newErrorResponse(c, http.StatusBadRequest, "email is required")
		return
	}

	// Проверяем, существует ли пользователь с таким email
	users, err := h.service.AuthService.GetUsers()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	for _, u := range users {
		if u.Email == input.Email {
			newErrorResponse(c, http.StatusBadRequest, "user with this email already exists")
			return
		}
	}

	userID, err := h.service.AuthService.CreateUser(entity.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  input.Password,
	})

	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	tokenDetails, err := h.service.AuthService.CreateToken(input.Email, input.Password)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := h.service.AuthService.FindUser(userID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"user":          user,
		"token_details": tokenDetails,
	})
}

type signInInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// @Summary SignIn
// @Tags auth
// @Description login
// @ID login
// @Accept  json
// @Produce  json
// @Param input body signInInput true "credentials"
// @Success 200 {object} TokenDetails
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/sign-in [post]
func (h *Handler) signIn(c *gin.Context) {
	var input signInInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tokenDetails, err := h.service.AuthService.CreateToken(input.Email, input.Password)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, TokenDetails{
		AccessToken:      tokenDetails.AccessToken,
		ExpiresIn:        tokenDetails.ExpiresIn,
		RefreshToken:     tokenDetails.RefreshToken,
		RefreshExpiresIn: tokenDetails.RefreshExpiresIn,
	})
}

type TokenDetails struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

// @Summary GetUsers
// @Tags auth
// @Description get users
// @ID get-users
// @Accept  json
// @Produce  json
// @Success 200 {array} entity.User
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/get-users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	users, err := h.service.AuthService.GetUsers()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, users)
}

// @Summary Refresh
// @Tags auth
// @Description refresh token
// @ID refresh
// @Accept  json
// @Produce  json
// @Param input body TokenDetails true "refresh token"
// @Success 200 {object} TokenDetails
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var requestBody struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	err := c.ShouldBindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	tokenDetails, err := h.service.AuthService.RefreshToken(requestBody.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, TokenDetails{
		AccessToken:      tokenDetails.AccessToken,
		ExpiresIn:        tokenDetails.ExpiresIn,
		RefreshToken:     tokenDetails.RefreshToken,
		RefreshExpiresIn: tokenDetails.RefreshExpiresIn,
	})
}
