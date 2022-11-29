package api

import (
	"encoding/json"
	"errors"
	"io"
	"neko-question-box-be/internal/config"
	"neko-question-box-be/internal/database"
	"neko-question-box-be/internal/database/types"
	"neko-question-box-be/internal/logger"
	"neko-question-box-be/internal/middleware"
	"neko-question-box-be/internal/services"
	"neko-question-box-be/pkg/handler"
	"net/http"
	"strings"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

type postQuestionReq struct {
	Id       string `json:"id"`       // captcha id
	Value    string `json:"value"`    // captcha value
	Question string `json:"question"` // 问题
	Username string `json:"username"` // 提问者
}

func (p postQuestionReq) isValid() bool {
	return strings.TrimSpace(p.Id) != "" && strings.TrimSpace(p.Value) != "" && strings.TrimSpace(p.Question) != ""
}

// 获取新的 captcha id
func getCaptcha(ctx *gin.Context) (handler.HandlerResponse, error) {
	id := captcha.New()
	logger.Infof("generate new captcha id: %s", id)
	return id, nil
}

// 通过 captcha id 获取到图片
func getCaptchaImage(ctx *gin.Context) (handler.HandlerResponse, error) {
	// 获取 id
	id := ctx.Query("id")
	if strings.TrimSpace(id) == "" {
		return nil, handler.ErrParams
	}
	// 把图片写进 buffer
	err := captcha.WriteImage(ctx.Writer, id, 200, 100)
	if err != nil {
		logger.Errorf("captcha buffer write error: %s", err.Error())
		return nil, handler.NewHandlerError(http.StatusInternalServerError, 50001, err.Error())
	} else {
		ctx.Abort()
	}
	return nil, nil
}

// 获取 bing 每日壁纸
func getBingWallpaper(ctx *gin.Context) (handler.HandlerResponse, error) {
	// send request
	resp, err := http.Get("https://www.bing.com/HPImageArchive.aspx?format=js&idx=2&n=1")
	if err != nil {
		logger.Errorf("bing wallpaper error, request error: %s", err.Error())
		return nil, handler.NewHandlerError(http.StatusInternalServerError, 50001, err.Error())
	}
	// read body
	p, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("bing wallpaper error, body read error: %s", err.Error())
		return nil, handler.NewHandlerError(http.StatusInternalServerError, 50001, err.Error())
	}
	// 如果直接返回 []byte，会被编码成 base64
	// 如果直接返回 string，会变成 json 里的 json 而不是字段
	// unmarshal body
	b := map[string]any{}
	json.Unmarshal(p, &b)
	return b, nil
}

// 获取已经有的问题和答案
func getQuestion(ctx *gin.Context) (handler.HandlerResponse, error) {
	questions, err := services.GetAllQuestions()
	if err != nil {
		logger.Errorf("get all questions error: %s", err.Error())
		return nil, handler.NewHandlerError(http.StatusInternalServerError, 50001, err.Error())
	}
	return questions, nil
}

// 提交新的问题
func postQuestion(ctx *gin.Context) (handler.HandlerResponse, error) {
	// 参数校验
	body := postQuestionReq{}
	_, MyClaims, err := middleware.ParseToken(ctx)
	body.Username = MyClaims.Username
	if err := ctx.Bind(&body); err != nil {
		return nil, handler.ErrParams
	}
	if !body.isValid() {
		return nil, handler.ErrParams
	}
	// captcha 校验
	if !captcha.VerifyString(body.Id, body.Value) {
		return nil, handler.ErrCaptcha
	}
	// 存入问题库

	err = services.CreateNewQuestion(body.Question, body.Username)
	if err != nil {
		logger.Errorf("save new question error: %s", err.Error())
		// 问题已经存在
		if errors.Is(err, services.ErrQuestionExists) {
			return nil, handler.ErrQuestionExists
		}
		return nil, handler.NewHandlerError(http.StatusInternalServerError, 50001, err.Error())
	}
	// TG Bot 已启用，发送问题到指定 id
	if config.Conf.Telegram.Enabled {
		services.SendToTgChat(body.Question)
	}

	return "问题提交成功", nil
	return nil, nil

}

// 检查服务器状态
func getPing(ctx *gin.Context) (handler.HandlerResponse, error) {
	return nil, nil
}

func AuthLogin(ctx *gin.Context) (handler.HandlerResponse, error) {
	var userProfile types.User
	err := ctx.ShouldBind(&userProfile)
	if err != nil {
		return nil, handler.ErrParams
	}

	//从数据库中提取用户信息。
	resp, err := services.GetUser(userProfile.Username)
	if err != nil {
		if database.IsNoRecordFoundError(err) {
			return nil, handler.ErrUserDoesNotExist
		}
		return nil, err
	}
	if userProfile.Pwd != resp.Pwd {
		return nil, handler.ErrWrongPassWord
	}

	//签发令牌
	tokenString, err := middleware.IssueToken(userProfile)
	if err != nil {
		return nil, handler.NewHandlerError(http.StatusInternalServerError, 50001, err.Error())
	}

	// 最后，我们将客户端cookie token设置为刚刚生成的JWT
	ctx.SetCookie("token", tokenString, 600, "/", "localhost", false, true)
	ctx.String(http.StatusOK, "登入成功。")
	return userProfile, nil
}

func OtherHandlers() handler.HandlerGroup {
	return handler.HandlerGroup{
		Name: "",
		Group: map[string][]handler.Handler{
			"captcha": {handler.NewHandler(http.MethodGet, getCaptcha)},
			"captcha-image": {
				handler.NewHandler(http.MethodGet, getCaptchaImage),
			},
			"bing": {handler.NewHandler(http.MethodGet, getBingWallpaper)},
			"login": {
				handler.NewHandler(http.MethodPost, AuthLogin),
			},
			"ping": {handler.NewHandler(http.MethodGet, getPing)},
		},
	}
}

func QuestionHandlers() handler.HandlerGroup {
	return handler.HandlerGroup{
		Name: "",
		Group: map[string][]handler.Handler{
			"doQuestion": {
				handler.NewHandler(http.MethodGet, getQuestion),
				handler.NewHandler(http.MethodPost, postQuestion),
			},
		},
	}
}
