package controllers

import (
	"errors"
	"fmt"
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

func LoginGet(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/login.html", gin.H{"citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

func LoginPost(c *gin.Context) {
	Auth := authentication.Resolve()
	JWT := jwt.Resolve()

	// validate and bind user input
	var loginData LoginCreds
	if err := c.ShouldBind(&loginData); err != nil {
		flashMessage(c, "The username or password you entered is incorrect.")
		log.Debugf("Error binding login credentials to login struct: %s", err.Error())
		return
	}

	// get the user record by email from db
	var user models.User
	result := database.DB.Where("username = ?", loginData.Username).First(&user)
	// check if the record not found
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.Redirect(http.StatusFound, "/auth/login")
		flashMessage(c, "The username or password you entered is incorrect.")
		log.Debugf("%s", result.Error)
		return
	}

	// handle database error incase there is any
	if result.Error != nil {
		c.Redirect(http.StatusFound, "/auth/login")
		flashMessage(c, "The username or password you entered is incorrect.")
		log.Debugf("Error getting user from database: %s", result.Error)
		return
	}

	// compare the password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		// wrong password
		c.Redirect(http.StatusFound, "/auth/login")
		flashMessage(c, "The username or password you entered is incorrect.")
		log.Debugf("Wrong password entered: %s", err)
		return
	}

	// generate the jwt token
	_, err = JWT.CreateToken(user.ID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Debugf("Error creating JWT token: %s", err)
		return
	}

	// generate the token
	_, err = JWT.CreateRefreshToken(user.ID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Debugf("Error refreshing JWT token: %s", err)
		return
	}

	// login the user
	err = Auth.Login(user.ID, c)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Debugf("Error logging in the user: %s", err)
	}

	// set user as logged in via Gin context
	c.Set("is_logged_in", true)

	log.Debugf("User %s successfully logged in", user.Username)

	// render response
	// TODO respond with JSON if selected, or HTML if selected
	// c.JSON(http.StatusOK, gin.H{
	// 	"data": map[string]string{
	// 		"token":        token,
	// 		"refreshToken": refreshToken,
	// 	},
	// })
	flashMessage(c, fmt.Sprintf("User '%s' logged in successfully.", user.Username))
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

	// render response
	// TODO respond with JSON if selected, or HTML if selected
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "logged out successfully",
	// })
	flashMessage(c, "User logged out successfully.")
	redirectPath := "/entries/"
	c.Redirect(http.StatusFound, redirectPath)
}

// GET /auth/register
// Display registration HTML form
func RegisterGet(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/register.html", gin.H{"citycode": c.MustGet("citycode").(string), "messages": flashes(c)})
}

// POST /auth/register
// Create a new user / register a new user
func RegisterPost(c *gin.Context) {
	// bind the input to the user's model
	var user models.User
	if err := c.ShouldBind(&user); err != nil {
		flashMessage(c, "The registration information you entered is incorrect, please try again.")
		log.Debugf("Error binding user credentials to user model struct: %s", err.Error())
		return
	}

	// check if there is a record with the given username
	res := database.DB.Where("username = ?", user.Username).First(&models.User{})
	if res.Error == nil {
		c.Redirect(http.StatusFound, "/auth/register")
		flashMessage(c, "User already signed up.")
		log.Debugf("User already signed up: %s", res.Error)
		return
	}

	// check the referral code from registration form matches the one from environment variable/config
	referralCode, _ := c.Get("referralcode")
	if user.ReferralCode != referralCode {
		c.Redirect(http.StatusFound, "/auth/register")
		flashMessage(c, "The referral code you entered is incorrect.")
		log.Debugf("Invalid referral code: %s", user.ReferralCode)
		return
	}

	// hash the passowrd
	hahsedPWD, err := hashPassword(user.Password)
	if err != nil {
		// wrong password
		c.Redirect(http.StatusFound, "/auth/register")
		flashMessage(c, "The username or password you entered is incorrect.")
		log.Debugf("Wrong password entered: %s", err)
		return
	}

	// use the hashed password
	user.Password = hahsedPWD
	// create the db record
	res = database.DB.Create(&user)
	if res.Error != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Debugf("Error creating new user record: %s", res.Error.Error())
		return
	}

	// render response
	// TODO respond with JSON if selected, or HTML if selected
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "signup successful",
	// })
	flashMessage(c, fmt.Sprintf("New user '%s' registered successfully.", user.Username))
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
