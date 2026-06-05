package middleware

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/logger"
	"net/http"
	"runtime/debug"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Recoverer(log logger.Logger) func(http.Handler) http.Handler {
	if log == nil {
		log = logger.NewNop()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				panicValue := recover()
				if panicValue == nil {
					return
				}

				log.Error(
					r.Context(),
					"panic recovered",
					zap.Any("panic", panicValue),
					zap.ByteString("stacktrace", debug.Stack()),
				)

				if ww.Status() != 0 {
					return
				}

				response.WriteError(
					ww,
					http.StatusInternalServerError,
					response.CodeInternalServerError,
					"internal server error",
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
