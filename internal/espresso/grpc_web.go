package espresso

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gregorychen3/espresso-controller/internal/espresso/power_manager"
	"github.com/gregorychen3/espresso-controller/internal/log"
	"github.com/gregorychen3/espresso-controller/internal/metrics"
	"github.com/hako/durafmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

// WebMiddleware to handle gRPC calls from browser. It is also able to isolate
// REST traffic and handle it accordingly
type WebMiddleware struct {
	*grpcweb.WrappedGrpcServer
}

var debugLevelURIs = map[string]struct{}{}

// Handler to isolate gRPC vs. non-gRPC requests.
//
// If the incoming traffic is gRPC, use grpc web to unmarshal the incoming grpc
// request.
// If the call is not gRPC web, then it is treated like regular REST.
func (m *WebMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.IsAcceptableGrpcCorsRequest(r) || m.IsGrpcWebRequest(r) {
			m.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewGrpcWebMiddleware(grpcWeb *grpcweb.WrappedGrpcServer) *WebMiddleware {
	return &WebMiddleware{grpcWeb}
}

type GRPCWebServer struct {
	grpcServer *grpc.Server
	fs         embed.FS
}

func NewGRPCWebServer(server *grpc.Server, uiFS embed.FS) *GRPCWebServer {
	return &GRPCWebServer{
		grpcServer: server,
		fs:         uiFS,
	}
}

func (s *GRPCWebServer) Listen(listener net.Listener, enableDevLogger bool, powerManager *power_manager.PowerManager) error {
	loggerMiddleware := NewProdLoggerMiddleware
	if enableDevLogger {
		loggerMiddleware = middleware.Logger
	}

	router := chi.NewRouter()

	router.Use(
		loggerMiddleware,
		middleware.Recoverer,
		cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}).Handler,
	)

	router.Post("/scheduling/on", func(writer http.ResponseWriter, req *http.Request) {
		powerManager.PowerOn()
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
	})
	router.Post("/scheduling/off", func(writer http.ResponseWriter, req *http.Request) {
		powerManager.PowerOff()
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
	})

	router.Post("/power/on", func(writer http.ResponseWriter, req *http.Request) {
		powerManager.PowerOn()
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
	})
	router.Post("/power/off", func(writer http.ResponseWriter, req *http.Request) {
		powerManager.PowerOff()
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
	})
	router.Post("/power/toggle", func(writer http.ResponseWriter, req *http.Request) {
		powerManager.PowerToggle()
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
	})
	router.Post("/power/total-off", func(writer http.ResponseWriter, req *http.Request) {
		powerManager.TotalPowerOff()
		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
	})
	router.Get("/power/status", func(writer http.ResponseWriter, req *http.Request) {

		type PowerManagerStatus struct {
			PowerSchedule        power_manager.PowerSchedule
			AutoOffDuration      string
			OnSince              string
			CurrentlyInASchedule bool
			LastInteraction      string
			PowerOn              bool
			StopScheduling       bool
			TotalOff             bool
		}

		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)

		ps := powerManager.GetStatus()
		var onSince time.Duration
		if ps.OnSince.Equal(time.Time{}) {
			onSince = time.Time{}.Sub(time.Time{})
		} else {
			onSince = time.Now().Sub(ps.OnSince)
		}

		humanPowerStatus := PowerManagerStatus{
			PowerSchedule:        ps.PowerSchedule,
			AutoOffDuration:      durafmt.Parse(ps.AutoOffDuration).String(),
			OnSince:              durafmt.Parse(onSince).LimitFirstN(2).String(),
			CurrentlyInASchedule: ps.CurrentlyInASchedule,
			LastInteraction:      ps.LastInteraction,
			PowerOn:              ps.PowerOn,
			StopScheduling:       ps.StopScheduling,
			TotalOff:             ps.TotalOff,
		}

		j, _ := json.Marshal(humanPowerStatus)
		writer.Write(j)
	})

	router.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		metrics.CollectSystemMetrics()
		promhttp.Handler().ServeHTTP(w, req)
	}))

	faviconBytes, _ := s.fs.ReadFile("ui/build/favicon.ico")

	router.Get("/favicon.ico", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write(faviconBytes)
	})

	router.Group(func(r chi.Router) {
		r.Use(NewGrpcWebMiddleware(grpcweb.WrapServer(s.grpcServer)).Handler)
		sub, _ := fs.Sub(s.fs, "ui/build/static")

		indexBytes, err := s.fs.ReadFile("ui/build/index.html")

		r.Get("/static/*", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Cache-Control", "max-age=31536000")
			writer.Header().Set("Content-Encoding", "gzip")
			request.URL.Path = strings.Replace(request.RequestURI, ".js", ".js.gz", 1)
			http.StripPrefix("/static/", http.FileServer(http.FS(sub))).ServeHTTP(writer, request)
		})

		// respond with index.html for all other routes (react router routes)
		r.Get("/*", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(200)
			if err != nil {
				log.Error("error serving index.html", zap.Error(err))
			}
			writer.Write(indexBytes)
		})

		// matches grpc requests to trigger grpc-web middleware
		r.Post("/*", func(http.ResponseWriter, *http.Request) {})
	})

	return http.Serve(listener, router)
}

type logEntry struct {
	req *http.Request
}

func (e *logEntry) Write(status, bytes int, elapsed time.Duration) {
	logFunc := log.Info
	if _, ok := debugLevelURIs[e.req.RequestURI]; ok {
		logFunc = log.Debug
	}
	logFunc(
		"Finished http request",
		zap.String("method", e.req.Method),
		zap.String("host", e.req.Host),
		zap.String("requestURI", e.req.RequestURI),
		zap.String("remoteAddr", e.req.RemoteAddr),
		zap.Int("status", status),
		zap.Int("responseSizeBytes", bytes),
		zap.Duration("requestDurationMs", elapsed),
	)
}
func NewProdLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := logEntry{req: r}
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t0 := time.Now()
		defer func() {
			entry.Write(ww.Status(), ww.BytesWritten(), time.Since(t0))
		}()
		r = r.WithContext(context.WithValue(r.Context(), middleware.LogEntryCtxKey, entry))
		next.ServeHTTP(ww, r)
	})
}
