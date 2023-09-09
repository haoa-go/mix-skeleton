package middleware

import (
	httpHeper "app/common/httpHelper"
	"app/common/log"
	"app/exception"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RecoverMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			switch err.(type) {
			case exception.MsgException:
				var ex = err.(exception.MsgException)
				c.JSON(http.StatusOK, gin.H{
					"code": ex.Code,
					"msg":  ex.Msg,
					"data": ex.Data,
				})
			case exception.ErrorException:
				var ex = err.(exception.ErrorException)
				log.ErrHandle(err)
				c.JSON(http.StatusOK, gin.H{
					"code": ex.Code,
					"msg":  ex.Msg,
					"data": ex.Data,
				})
			default:
				log.ErrHandle(err)
				c.JSON(http.StatusOK, gin.H{
					"code": httpHeper.CodeServerError,
					"msg":  httpHeper.GetMsgByCode(httpHeper.CodeServerError),
					"data": "",
				})
			}

		}()

		c.Next()
	}
}
