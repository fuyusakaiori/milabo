package param

import (
	"github.com/fuyusakaiori/gateway/core/constant"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/universal-translator"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
	"strings"
)

// ShouldValidBind 自定义绑定参数并且验证的方法
func ShouldValidBind(c *gin.Context, params interface{}) error {
	if err := c.ShouldBind(params); err != nil {
		return err
	}
	//获取验证器
	valid, err := getValidator(c)
	if err != nil {
		return err
	}
	//获取翻译器
	trans, err := getTranslation(c)
	if err != nil {
		return err
	}
	err = valid.Struct(params)
	if err != nil {
		errs := err.(validator.ValidationErrors)
		var sliceErrs []string
		for _, e := range errs {
			sliceErrs = append(sliceErrs, e.Translate(trans))
		}
		return errors.New(strings.Join(sliceErrs, ","))
	}
	return nil
}

// getValidator 获取结构体验证器
func getValidator(c *gin.Context) (*validator.Validate, error) {
	val, ok := c.Get(constant.ValidatorKey)
	if !ok {
		return nil, errors.New("未设置验证器")
	}
	validate, ok := val.(*validator.Validate)
	if !ok {
		return nil, errors.New("获取验证器失败")
	}
	return validate, nil
}

// getTranslation 获取错误信息的翻译器
func getTranslation(c *gin.Context) (ut.Translator, error) {
	trans, ok := c.Get(constant.TranslatorKey)
	if !ok {
		return nil, errors.New("未设置翻译器")
	}
	translate, ok := trans.(ut.Translator)
	if !ok {
		return nil, errors.New("获取翻译器失败")
	}
	return translate, nil
}
