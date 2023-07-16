module github.com/fuyusakaiori/gateway/core

go 1.14

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/e421083458/go_gateway v0.0.0-20230623173026-d2d21f7c93e8
	github.com/e421083458/golang_common v1.2.1
	github.com/e421083458/gorm v1.0.1
	github.com/gin-gonic/contrib v0.0.0-20191209060500-d6e26eeaa607
	github.com/gin-gonic/gin v1.7.7
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/pkg/errors v0.8.1
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.2.0
	github.com/swaggo/swag v1.6.5
	gopkg.in/go-playground/validator.v9 v9.29.0
)

replace github.com/gin-contrib/sse v0.1.0 => github.com/e421083458/sse v0.1.1
