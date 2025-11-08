package auth

import (
	"chat/globals"
	"chat/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ProfileAPI(c *gin.Context) {
	user := GetUserByCtx(c)
	if user == nil {
		return
	}
	db := utils.GetDBFromContext(c)
	var id int64
	var username, email string
	var admin bool
	if err := globals.QueryRowDb(db, "SELECT id, username, email, is_admin FROM auth WHERE username = ?", user.Username).Scan(&id, &username, &email, &admin); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data": gin.H{
			"id":       id,
			"username": username,
			"email":    email,
			"admin":    admin,
		},
	})
}

func UpdateProfileAPI(c *gin.Context) {
	user := GetUserByCtx(c)
	if user == nil {
		return
	}
	db := utils.GetDBFromContext(c)
	var form UpdateProfileForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "error": "bad request"})
		return
	}
	newUsername := strings.TrimSpace(form.Username)
	newEmail := strings.TrimSpace(form.Email)

	var id int64
	var curUsername, curEmail string
	if err := globals.QueryRowDb(db, "SELECT id, username, email FROM auth WHERE username = ?", user.Username).Scan(&id, &curUsername, &curEmail); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
		return
	}

	if newEmail != "" && newEmail != curEmail {
		if !validateEmail(newEmail) {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "invalid email"})
			return
		}
		if IsEmailExist(db, newEmail) {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "email already exists"})
			return
		}
		if _, err := globals.ExecDb(db, "UPDATE auth SET email = ? WHERE id = ?", newEmail, id); err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		curEmail = newEmail
	}

	token := ""
	if newUsername != "" && newUsername != curUsername {
		if !validateUsername(newUsername) {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "invalid username"})
			return
		}
		if IsUserExist(db, newUsername) {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "username already exists"})
			return
		}
		if _, err := globals.ExecDb(db, "UPDATE auth SET username = ? WHERE id = ?", newUsername, id); err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		u := &User{ID: id}
		if t, err := u.GenerateTokenSafe(db); err == nil {
			token = t
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"token":  token,
	})
}
