package web

import (
	"fmt"
	"net/http"
	"time"

	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// UserHandler 我准备在它上面定义跟用户有关的路由
type UserHandler struct {
	svc             *service.UserService
	emailExp        *regexp.Regexp
	passwordExp     *regexp.Regexp
	birthdayExp     *regexp.Regexp
	nicknameExp     *regexp.Regexp
	introductionExp *regexp.Regexp
	locationExp     *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern        = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern     = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
		birthdayRegexPattern     = `^(19|20)\d{2}-\d{2}-\d{2}$`
		nicknameRegexPattern     = `^[\u4e00-\u9fa5_a-zA-Z0-9]{2,10}$`
		introductionRegexPattern = `^[\u4e00-\u9fa5，。]{10,300}$`
		locationRegexPattern     = `^[\u4e00-\u9fa5]{3,60}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	birthdayExp := regexp.MustCompile(birthdayRegexPattern, regexp.None)
	nicknameExp := regexp.MustCompile(nicknameRegexPattern, regexp.None)
	introductionExp := regexp.MustCompile(introductionRegexPattern, regexp.None)
	locationExp := regexp.MustCompile(locationRegexPattern, regexp.None)
	return &UserHandler{
		svc:             svc,
		emailExp:        emailExp,
		passwordExp:     passwordExp,
		birthdayExp:     birthdayExp,
		nicknameExp:     nicknameExp,
		introductionExp: introductionExp,
		locationExp:     locationExp,
	}
}

func (u *UserHandler) RegisterRoutesV1(ug *gin.RouterGroup) {
	ug.GET("/profile", u.Profile)
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.POST("/edit", u.Edit)
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.GET("/profile", u.Profile)
	ug.POST("/signup", u.SignUp)
	//ug.POST("/login", u.Login)
	ug.POST("/login", u.Login)
	ug.POST("/edit", u.Edit)
	ug.GET("/status", u.Status)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	// Bind 方法会根据 Content-Type 来解析你的数据到 req 里面
	// 解析错了，就会直接写回一个 400 的错误
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "你的邮箱格式不对")
		return
	}
	if req.ConfirmPassword != req.Password {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		// 记录日志
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码至少8位,包含数字、特殊字符")
		return
	}

	// 调用一下 svc 的方法
	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 步骤2
	// 在这里用 JWT 设置登录态
	// 生成一个 JWT token

	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid:       user.Id,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
	fmt.Println(user)
	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 步骤2
	// 在这里登录成功了
	// 设置 session
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 session 里面的值
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		Secure:   true,
		HttpOnly: true,
		// 一分钟过期
		MaxAge: 600,
	})
	sess.Save()
	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 session 里面的值
	sess.Options(sessions.Options{
		//Secure: true,
		//HttpOnly: true,
		MaxAge: -1,
	})
	sess.Save()
	ctx.String(http.StatusOK, "退出登录成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type editReq struct {
		NickName     string `json:"nickname"`
		Birthday     string `json:"birthday"`
		Introduction string `json:"introduction"`
		Location     string `json:"location"`
		Avatar       string `json:"avatar"`
	}
	var req editReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if req.Birthday != "" {
		ok, err := u.birthdayExp.MatchString(req.Birthday)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		if !ok {
			ctx.String(http.StatusOK, "生日格式不对,必须以19或者20开头,正确例子1999-1-28")
			return
		}
	}
	if req.NickName != "" {
		ok, err := u.nicknameExp.MatchString(req.NickName)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		if !ok {
			ctx.String(http.StatusOK, "昵称格式不对,至少2个中文字符,并且不能超过10个中文字符")
			return
		}
	}
	if req.Introduction != "" {
		ok, err := u.introductionExp.MatchString(req.Introduction)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		if !ok {
			ctx.String(http.StatusOK, "个人简介格式不对,至少10个中文字符,并且不能超过300个中文字符")
			return
		}
	}
	if req.Location != "" {
		ok, err := u.locationExp.MatchString(req.Location)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		if !ok {
			ctx.String(http.StatusOK, "地址格式不对,至少3个中文字符,并且不能超过60个中文字符")
			return
		}
	}
	sess := sessions.Default(ctx)
	id := sess.Get("userId").(int64)
	err := u.svc.Edit(ctx, domain.User{
		Id:           id,
		Birthday:     req.Birthday,
		NickName:     req.NickName,
		Introduction: req.Introduction,
		Location:     req.Location,
	})
	if err != nil {
		ctx.String(http.StatusOK, "编辑失败")
		return
	}
	ctx.String(http.StatusOK, "编辑成功")

}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, _ := ctx.Get("claims")
	// 你可以断定，必然有 claims
	//if !ok {
	//	// 你可以考虑监控住这里
	//	ctx.String(http.StatusOK, "系统错误")
	//	return
	//}
	// ok 代表是不是 *UserClaims
	claims, ok := c.(*UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	println(claims.Uid)
	ctx.String(http.StatusOK, "你的 profile")
	// 这边就是你补充 profile 的其它代码
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	type ProfileResp struct {
		Email        string `json:"email"`
		Birthday     string `json:"birthday"`
		Nickname     string `json:"nickname"`
		Introduction string `json:"introduction"`
		Location     string `json:"location"`
	}
	sess := sessions.Default(ctx)
	id := sess.Get("userId").(int64)
	user, err := u.svc.GetProfile(ctx, id)
	if err != nil {
		ctx.String(http.StatusOK, "获取简介失败")
	}
	ctx.JSON(http.StatusOK, ProfileResp{
		Email:        user.Email,
		Birthday:     user.Birthday,
		Nickname:     user.NickName,
		Introduction: user.Introduction,
		Location:     user.Location,
	})
}
func (u *UserHandler) Status(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "状态没问题")
}

type UserClaims struct {
	jwt.RegisteredClaims
	// 声明你自己的要放进去 token 里面的数据
	Uid int64
	// 自己随便加
	UserAgent string
}
