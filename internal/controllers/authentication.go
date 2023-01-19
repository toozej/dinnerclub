package controllers

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/gocondor/core/jwt"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginCreds struct {
	Username string `form:"username" json:"username" binding:"required,alphanum"`
	Password string `form:"password" json:"-" binding:"required,min=10"`
}

func Login(c *gin.Context) {
	Auth := authentication.Resolve()
	JWT := jwt.Resolve()

	// validate and bind user input
	var loginData LoginCreds
	if err := c.ShouldBind(&loginData); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"message": err.Error(),
		})
		return
	}

	// get the user record by email from db
	var user models.User
	result := database.DB.Where("username = ?", loginData.Username).First(&user)
	// check if the record not found
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"message": "wrong credentials",
		})
		return
	}

	// handle database error incase there is any
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "someting went wrong getting user from database",
		})
		return
	}

	// compare the password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		// wrong password
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"message": "wrong credentials",
		})
		return
	}

	// generate the jwt token
	_, err = JWT.CreateToken(user.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "something went wrong creating JWT token",
		})
	}

	// generate the token
	_, err = JWT.CreateRefreshToken(user.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "something went wrong refreshing JWT token",
		})
	}

	// login the user
	err = Auth.Login(user.ID, c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "something went wrong logging in",
		})
	}

	// set user as logged in via Gin context
	c.Set("is_logged_in", true)

	log.Debugf("User %s successfully logged in", user.Username)

	// TODO add flash for successful login

	// render response
	// TODO respond with JSON if selected, or HTML if selected
	// c.JSON(http.StatusOK, gin.H{
	// 	"data": map[string]string{
	// 		"token":        token,
	// 		"refreshToken": refreshToken,
	// 	},
	// })
	redirectPath := "/profile/"
	c.Redirect(http.StatusFound, redirectPath)
}

func Logout(c *gin.Context) {
	Auth := authentication.Resolve()

	err := Auth.Logout(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "something went wrong logging out",
		})
	}

	// TODO add flash for successful login

	// render response
	// TODO respond with JSON if selected, or HTML if selected
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "logged out successfully",
	// })
	redirectPath := "/entries/"
	c.Redirect(http.StatusFound, redirectPath)
}

func Register(c *gin.Context) {
	// bind the input to the user's model
	var user models.User
	if err := c.ShouldBind(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"message": err.Error(),
		})
		return
	}

	// check if there is a record with the given username
	res := database.DB.Where("username = ?", user.Username).First(&models.User{})
	if res.Error == nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"message": "user already signed up",
		})
		return
	}

	// hash the passowrd
	hahsedPWD, err := hashPassword(user.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err,
		})
		return
	}

	// use the hashed password
	user.Password = hahsedPWD
	// create the db record
	res = database.DB.Create(&user)
	if res.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": res.Error.Error(),
		})
		return
	}

	// TODO add flash for successful registration

	// render response
	// TODO respond with JSON if selected, or HTML if selected
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "signup successful",
	// })
	redirectPath := "/auth/login"
	c.Redirect(http.StatusFound, redirectPath)
}

// hashPassword hashs passwords
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func SetUserStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		Auth := authentication.Resolve()
		if authed, err := Auth.UserID(c); err == nil && authed != 0 {
			c.Set("is_logged_in", true)
		} else {
			c.Set("is_logged_in", false)
		}
	}
}

// This function ensures that a request will be aborted with an error
// if the user is not logged in
func EnsureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		loggedInInterface, _ := c.Get("is_logged_in")
		loggedIn := loggedInInterface.(bool)
		if !loggedIn {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

// This function ensures that a request will be aborted with an error
// if the user is already logged in
func EnsureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		loggedInInterface, _ := c.Get("is_logged_in")
		loggedIn := loggedInInterface.(bool)
		if loggedIn {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
