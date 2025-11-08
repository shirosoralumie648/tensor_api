package auth

import (
	"chat/globals"
	"chat/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

func getScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	return "http"
}

func getCallbackURL(c *gin.Context, provider string) string {
	base := fmt.Sprintf("%s://%s", getScheme(c), c.Request.Host)
	path := "/oauth/" + provider + "/callback"
	if viper.GetBool("serve_static") {
		path = "/api" + path
	}
	return base + path
}

func getFrontendOAuthCallback(c *gin.Context) string {
	base := fmt.Sprintf("%s://%s", getScheme(c), c.Request.Host)
	return base + "/oauth/callback"
}

func setOAuthState(c *gin.Context, cache *redis.Client, state string, provider string) {
	cache.Set(c, fmt.Sprintf("nio:oauth:%s", state), provider, 10*time.Minute)
}

func getOAuthState(c *gin.Context, cache *redis.Client, state string) string {
	v, _ := cache.Get(c, fmt.Sprintf("nio:oauth:%s", state)).Result()
	return v
}

func delOAuthState(c *gin.Context, cache *redis.Client, state string) {
	cache.Del(c, fmt.Sprintf("nio:oauth:%s", state))
}

func OAuthStart(c *gin.Context) {
	provider := strings.ToLower(c.Param("provider"))
	state := utils.GenerateChar(24)
	cache := utils.GetCacheFromContext(c)
	// support binding mode: append "|bind|<token>" into state storage value
	bindMode := c.Query("bind") == "1" || strings.ToLower(c.Query("mode")) == "bind"
	token := strings.TrimSpace(c.Query("token"))
	code := strings.TrimSpace(c.Query("code"))
	value := provider
	if bindMode {
		if token == "" || code == "" {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "bind requires token and code"})
			return
		}
		value = provider + "|bind|" + token + "|" + code
	}
	setOAuthState(c, cache, state, value)

	redirectURL := getCallbackURL(c, provider)

	switch provider {
	case "github":
		clientID := viper.GetString("oauth.github.client_id")
		if clientID == "" {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "github not configured"})
			return
		}
		u, _ := url.Parse("https://github.com/login/oauth/authorize")
		q := u.Query()
		q.Set("client_id", clientID)
		q.Set("redirect_uri", redirectURL)
		q.Set("scope", "read:user user:email")
		q.Set("state", state)
		u.RawQuery = q.Encode()
		c.Redirect(http.StatusFound, u.String())
		return
	case "google":
		clientID := viper.GetString("oauth.google.client_id")
		if clientID == "" {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "google not configured"})
			return
		}
		u, _ := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
		q := u.Query()
		q.Set("client_id", clientID)
		q.Set("redirect_uri", redirectURL)
		q.Set("response_type", "code")
		q.Set("scope", "openid email profile")
		q.Set("access_type", "online")
		q.Set("include_granted_scopes", "true")
		q.Set("state", state)
		u.RawQuery = q.Encode()
		c.Redirect(http.StatusFound, u.String())
		return
	case "wechat":
		appID := viper.GetString("oauth.wechat.app_id")
		if appID == "" {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "wechat not configured"})
			return
		}
		u, _ := url.Parse("https://open.weixin.qq.com/connect/qrconnect")
		q := u.Query()
		q.Set("appid", appID)
		q.Set("redirect_uri", redirectURL)
		q.Set("response_type", "code")
		q.Set("scope", "snsapi_login")
		q.Set("state", state)
		u.RawQuery = q.Encode()
		u.Fragment = "wechat_redirect"
		c.Redirect(http.StatusFound, u.String())
		return
	case "qq":
		appID := viper.GetString("oauth.qq.app_id")
		if appID == "" {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "qq not configured"})
			return
		}
		u, _ := url.Parse("https://graph.qq.com/oauth2.0/authorize")
		q := u.Query()
		q.Set("response_type", "code")
		q.Set("client_id", appID)
		q.Set("redirect_uri", redirectURL)
		q.Set("state", state)
		q.Set("scope", "get_user_info")
		u.RawQuery = q.Encode()
		c.Redirect(http.StatusFound, u.String())
		return
	default:
		c.JSON(http.StatusOK, gin.H{"status": false, "error": "unsupported provider"})
		return
	}
}

type githubUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Login string `json:"login"`
	Email string `json:"email"`
}

type googleUser struct {
	Sub   string `json:"sub"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type wechatToken struct {
	AccessToken string `json:"access_token"`
	OpenID      string `json:"openid"`
	UnionID     string `json:"unionid"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type qqOpenID struct {
	ClientID string `json:"client_id"`
	OpenID   string `json:"openid"`
}

func httpGetJSON(url string, headers map[string]string, out any) error {
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
	}
	// QQ openid api may return "callback( ... );"
	text := strings.TrimSpace(string(b))
	if strings.HasPrefix(text, "callback(") {
		text = strings.TrimPrefix(text, "callback(")
		text = strings.TrimSuffix(text, ");")
		b = []byte(text)
	}
	return json.Unmarshal(b, out)
}

func httpPostFormJSON(u string, data url.Values, headers map[string]string, out any) error {
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest(http.MethodPost, u, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
	}
	return json.Unmarshal(b, out)
}

func getOrCreateOAuthUser(c *gin.Context, provider, openID, display, email string, unionID string) (*User, error) {
	db := utils.GetDBFromContext(c)
	var userID int64
	err := globals.QueryRowDb(db, "SELECT user_id FROM oauth WHERE provider = ? AND open_id = ?", provider, openID).Scan(&userID)
	if err == nil && userID > 0 {
		u := &User{ID: userID}
		return u, nil
	}
	if globals.CloseRegistration {
		return nil, errors.New("registration closed")
	}
	name := strings.TrimSpace(display)
	if name == "" {
		name = fmt.Sprintf("%s_%s", provider, utils.GenerateChar(8))
	}
	if len(name) < 2 {
		name = name + utils.GenerateChar(2)
	}
	if len(name) > 24 {
		name = name[:24]
	}
	// ensure unique username
	for i := 0; IsUserExist(db, name); i++ {
		name = fmt.Sprintf("%s_%d", name, i+1)
		if len(name) > 24 {
			name = name[:24]
		}
	}
	password := utils.GenerateChar(32)
	user := &User{Username: name, Password: utils.Sha2Encrypt(password), Email: email, BindID: getMaxBindId(db) + 1, Token: utils.Sha2Encrypt(name+email)}
	if _, err := globals.ExecDb(db, `INSERT INTO auth (username, password, email, bind_id, token) VALUES (?, ?, ?, ?, ?)`, user.Username, user.Password, user.Email, user.BindID, user.Token); err != nil {
		return nil, err
	}
	user.CreateInitialQuota(db)
	if _, err := globals.ExecDb(db, `INSERT INTO oauth (provider, open_id, union_id, user_id) VALUES (?, ?, ?, ?)`, provider, openID, unionID, user.GetID(db)); err != nil {
		return nil, err
	}
	return user, nil
}

func OAuthCallback(c *gin.Context) {
	provider := strings.ToLower(c.Param("provider"))
	state := c.Query("state")
	code := c.Query("code")
	cache := utils.GetCacheFromContext(c)
	raw := getOAuthState(c, cache, state)
	parts := strings.Split(raw, "|")
	storedProvider := ""
	if len(parts) > 0 {
		storedProvider = parts[0]
	}
	bindMode := len(parts) >= 2 && parts[1] == "bind"
	bindToken := ""
	if bindMode && len(parts) >= 3 {
		bindToken = parts[2]
	}
	bindCode := ""
	if bindMode && len(parts) >= 4 {
		bindCode = parts[3]
	}
	if code == "" || state == "" || storedProvider != provider {
		c.JSON(http.StatusOK, gin.H{"status": false, "error": "invalid state or code"})
		return
	}
	delOAuthState(c, cache, state)
	redirectFront := getFrontendOAuthCallback(c)

	switch provider {
	case "github":
		clientID := viper.GetString("oauth.github.client_id")
		clientSecret := viper.GetString("oauth.github.client_secret")
		if clientID == "" || clientSecret == "" {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "github not configured"})
			return
		}
		data := url.Values{}
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
		data.Set("code", code)
		data.Set("redirect_uri", getCallbackURL(c, provider))
		var tok map[string]any
		if err := httpPostFormJSON("https://github.com/login/oauth/access_token", data, map[string]string{"Accept": "application/json"}, &tok); err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		accessToken, _ := tok["access_token"].(string)
		var info githubUser
		if err := httpGetJSON("https://api.github.com/user", map[string]string{"Authorization": "token " + accessToken, "Accept": "application/json"}, &info); err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		if bindMode {
			db := utils.GetDBFromContext(c)
			target := ParseToken(c, bindToken)
			if target == nil {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "bind requires login"})
				return
			}
			email := target.GetEmail(db)
			if strings.TrimSpace(email) == "" {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "email required"})
				return
			}
			if !checkCode(c, cache, email, bindCode) {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "invalid or expired code"})
				return
			}
			if err := bindOAuthToUser(c, target, provider, fmt.Sprintf("%d", info.ID), ""); err != nil {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
				return
			}
			c.Redirect(http.StatusFound, fmt.Sprintf("%s?bind=%s&ok=1", redirectFront, provider))
			return
		}
		u, err := getOrCreateOAuthUser(c, provider, fmt.Sprintf("%d", info.ID), firstNonEmpty(info.Name, info.Login), info.Email, "")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		token, err := u.GenerateTokenSafe(utils.GetDBFromContext(c))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("%s?jwt=%s", redirectFront, url.QueryEscape(token)))
		return
	case "google":
		clientID := viper.GetString("oauth.google.client_id")
		clientSecret := viper.GetString("oauth.google.client_secret")
		if clientID == "" || clientSecret == "" {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "google not configured"})
			return
		}
		data := url.Values{}
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
		data.Set("code", code)
		data.Set("grant_type", "authorization_code")
		data.Set("redirect_uri", getCallbackURL(c, provider))
		var tok struct{ AccessToken string `json:"access_token"` }
		if err := httpPostFormJSON("https://oauth2.googleapis.com/token", data, nil, &tok); err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		var info googleUser
		if err := httpGetJSON("https://openidconnect.googleapis.com/v1/userinfo", map[string]string{"Authorization": "Bearer " + tok.AccessToken}, &info); err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		if bindMode {
			db := utils.GetDBFromContext(c)
			target := ParseToken(c, bindToken)
			if target == nil {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "bind requires login"})
				return
			}
			email := target.GetEmail(db)
			if strings.TrimSpace(email) == "" {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "email required"})
				return
			}
			if !checkCode(c, cache, email, bindCode) {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "invalid or expired code"})
				return
			}
			if err := bindOAuthToUser(c, target, provider, info.Sub, ""); err != nil {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
				return
			}
			c.Redirect(http.StatusFound, fmt.Sprintf("%s?bind=%s&ok=1", redirectFront, provider))
			return
		}
		u, err := getOrCreateOAuthUser(c, provider, info.Sub, info.Name, info.Email, "")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		token, err := u.GenerateTokenSafe(utils.GetDBFromContext(c))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("%s?jwt=%s", redirectFront, url.QueryEscape(token)))
		return
	case "wechat":
		appID := viper.GetString("oauth.wechat.app_id")
		secret := viper.GetString("oauth.wechat.app_secret")
		if appID == "" || secret == "" {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "wechat not configured"})
			return
		}
		var tok wechatToken
		u := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code", url.QueryEscape(appID), url.QueryEscape(secret), url.QueryEscape(code))
		if err := httpGetJSON(u, nil, &tok); err != nil || tok.ErrCode != 0 {
			if err == nil {
				err = fmt.Errorf("wechat error: %d %s", tok.ErrCode, tok.ErrMsg)
			}
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		if bindMode {
			db := utils.GetDBFromContext(c)
			target := ParseToken(c, bindToken)
			if target == nil {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "bind requires login"})
				return
			}
			email := target.GetEmail(db)
			if strings.TrimSpace(email) == "" {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "email required"})
				return
			}
			if !checkCode(c, cache, email, bindCode) {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "invalid or expired code"})
				return
			}
			if err := bindOAuthToUser(c, target, provider, tok.OpenID, tok.UnionID); err != nil {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
				return
			}
			c.Redirect(http.StatusFound, fmt.Sprintf("%s?bind=%s&ok=1", redirectFront, provider))
			return
		}
		u2, err := getOrCreateOAuthUser(c, provider, tok.OpenID, "wechat_user", "", tok.UnionID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		token, err := u2.GenerateTokenSafe(utils.GetDBFromContext(c))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("%s?jwt=%s", redirectFront, url.QueryEscape(token)))
		return
	case "qq":
		appID := viper.GetString("oauth.qq.app_id")
		secret := viper.GetString("oauth.qq.app_secret")
		if appID == "" || secret == "" {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "qq not configured"})
			return
		}
		// get access token
		uTok := fmt.Sprintf("https://graph.qq.com/oauth2.0/token?grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s", url.QueryEscape(appID), url.QueryEscape(secret), url.QueryEscape(code), url.QueryEscape(getCallbackURL(c, provider)))
		resp, err := http.Get(uTok)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		vals, _ := url.ParseQuery(string(b))
		accessToken := vals.Get("access_token")
		if accessToken == "" {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": "qq empty token"})
			return
		}
		var open qqOpenID
		if err := httpGetJSON("https://graph.qq.com/oauth2.0/me?access_token="+url.QueryEscape(accessToken), nil, &open); err != nil || open.OpenID == "" {
			if err == nil {
				err = errors.New("qq openid empty")
			}
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		if bindMode {
			db := utils.GetDBFromContext(c)
			target := ParseToken(c, bindToken)
			if target == nil {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "bind requires login"})
				return
			}
			email := target.GetEmail(db)
			if strings.TrimSpace(email) == "" {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "email required"})
				return
			}
			if !checkCode(c, cache, email, bindCode) {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": "invalid or expired code"})
				return
			}
			if err := bindOAuthToUser(c, target, provider, open.OpenID, ""); err != nil {
				c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
				return
			}
			c.Redirect(http.StatusFound, fmt.Sprintf("%s?bind=%s&ok=1", redirectFront, provider))
			return
		}
		u2, err := getOrCreateOAuthUser(c, provider, open.OpenID, "qq_user", "", "")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		token, err := u2.GenerateTokenSafe(utils.GetDBFromContext(c))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
			return
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("%s?jwt=%s", redirectFront, url.QueryEscape(token)))
		return
	default:
		c.JSON(http.StatusOK, gin.H{"status": false, "error": "unsupported provider"})
		return
	}
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return strings.TrimSpace(b)
}

func bindOAuthToUser(c *gin.Context, user *User, provider, openID, unionID string) error {
	db := utils.GetDBFromContext(c)
	var uid int64
	err := globals.QueryRowDb(db, "SELECT user_id FROM oauth WHERE provider = ? AND open_id = ?", provider, openID).Scan(&uid)
	if err == nil && uid > 0 {
		if uid == user.GetID(db) {
			return nil
		}
		return fmt.Errorf("this account has been bound with another user")
	}
	_, err = globals.ExecDb(db, `INSERT INTO oauth (provider, open_id, union_id, user_id) VALUES (?, ?, ?, ?)`, provider, openID, unionID, user.GetID(db))
	return err
}

type oauthBinding struct {
	Provider  string `json:"provider"`
	OpenID    string `json:"open_id"`
	UnionID   string `json:"union_id"`
	CreatedAt string `json:"created_at"`
}

func OAuthBindingsAPI(c *gin.Context) {
	user := GetUserByCtx(c)
	if user == nil {
		return
	}
	db := utils.GetDBFromContext(c)
	rows, err := globals.QueryDb(db, "SELECT provider, open_id, union_id, created_at FROM oauth WHERE user_id = ?", user.GetID(db))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
		return
	}
	defer rows.Close()
	var list []oauthBinding
	for rows.Next() {
		var b oauthBinding
		if err := rows.Scan(&b.Provider, &b.OpenID, &b.UnionID, &b.CreatedAt); err == nil {
			list = append(list, b)
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "data": list})
}

func OAuthUnbindAPI(c *gin.Context) {
	user := GetUserByCtx(c)
	if user == nil {
		return
	}
	provider := strings.ToLower(c.Param("provider"))
	db := utils.GetDBFromContext(c)
	var body struct{ Code string `json:"code"` }
	_ = c.ShouldBindJSON(&body)
	email := user.GetEmail(db)
	if strings.TrimSpace(email) == "" {
		c.JSON(http.StatusOK, gin.H{"status": false, "error": "email required"})
		return
	}
	if !checkCode(c, utils.GetCacheFromContext(c), email, strings.TrimSpace(body.Code)) {
		c.JSON(http.StatusOK, gin.H{"status": false, "error": "invalid or expired code"})
		return
	}
	if _, err := globals.ExecDb(db, "DELETE FROM oauth WHERE provider = ? AND user_id = ?", provider, user.GetID(db)); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true})
}
