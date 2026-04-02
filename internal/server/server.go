package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ashparshp/mentormatch-backend/internal/config"
	"github.com/ashparshp/mentormatch-backend/internal/middleware"
	"github.com/ashparshp/mentormatch-backend/internal/modules/admin"
	"github.com/ashparshp/mentormatch-backend/internal/modules/analytics"
	"github.com/ashparshp/mentormatch-backend/internal/modules/bookings"
	"github.com/ashparshp/mentormatch-backend/internal/modules/mentors"
	"github.com/ashparshp/mentormatch-backend/internal/modules/notifications"
	"github.com/ashparshp/mentormatch-backend/internal/modules/payments"
	"github.com/ashparshp/mentormatch-backend/internal/modules/recommendations"
	"github.com/ashparshp/mentormatch-backend/internal/modules/resources"
	"github.com/ashparshp/mentormatch-backend/internal/modules/reviews"
	"github.com/ashparshp/mentormatch-backend/internal/modules/users"
	"github.com/ashparshp/mentormatch-backend/internal/modules/chat"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
)

type Server struct {
	Router *chi.Mux
	DB     *pgxpool.Pool
	Config *config.Config
}

func NewServer(cfg *config.Config, db *pgxpool.Pool) *Server {
	r := chi.NewRouter()

	// Default middlewares
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// Security Headers Middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			next.ServeHTTP(w, r)
		})
	})

	// CORS Setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(c.Handler)

	s := &Server{
		Router: r,
		DB:     db,
		Config: cfg,
	}

	s.SetupRoutes()

	return s
}

func (s *Server) SetupRoutes() {
	// Repositories
	userRepo := users.NewRepository(s.DB)
	mentorRepo := mentors.NewRepository(s.DB)
	bookingRepo := bookings.NewRepository(s.DB)
	resourceRepo := resources.NewRepository(s.DB)
	reviewRepo := reviews.NewRepository(s.DB)
	paymentRepo := payments.NewRepository(s.DB)
	adminRepo := admin.NewRepository(s.DB)
	recRepo := recommendations.NewRepository(s.DB)
	analyticsRepo := analytics.NewRepository(s.DB)
	chatRepo := chat.NewRepository(s.DB)

	// Services
	userService := users.NewService(userRepo)
	mentorService := mentors.NewService(mentorRepo)
	reviewService := reviews.NewService(reviewRepo, bookingRepo)
	bookingService := bookings.NewService(bookingRepo, mentorRepo)
	chatService := chat.NewService(chatRepo, bookingRepo)

	// Handlers
	userHandler := users.NewHandler(userService)
	mentorHandler := mentors.NewHandler(mentorService, analyticsRepo)
	bookingHandler := bookings.NewHandler(bookingService)
	resourceHandler := resources.NewHandler(resourceRepo)
	reviewHandler := reviews.NewHandler(reviewService)
	paymentHandler := payments.NewHandler(paymentRepo, s.Config)
	adminHandler := admin.NewHandler(adminRepo)
	recHandler := recommendations.NewHandler(recRepo)
	analyticsHandler := analytics.NewHandler(analyticsRepo)
	chatHandler := chat.NewHandler(chatService)

	// Public Routes
	s.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	s.Router.Route("/api/v1", func(r chi.Router) {
		// Public Mentor Routes
		r.Get("/mentors", mentorHandler.DiscoverMentors)
		r.Get("/mentors/suggested-rate", mentorHandler.GetSuggestedRate)
		r.Get("/mentors/{id}", mentorHandler.GetMentor)
		r.Get("/mentors/{id}/reviews", reviewHandler.GetMentorReviews)
		r.Get("/resources", resourceHandler.ListResources)
		r.Get("/search/autocomplete", analyticsHandler.GetAutocomplete)

		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(s.Config))

			// User Routes
			r.Get("/me", userHandler.GetProfile)
			r.Put("/me", userHandler.UpdateProfile)
			r.Post("/onboard/student", userHandler.OnboardStudent)

			// Mentor Routes (Authenticated)
			r.Post("/resources/{id}/download", resourceHandler.DownloadResource)
			r.Post("/onboard/mentor", mentorHandler.OnboardMentor)
			r.Put("/mentors/availability", mentorHandler.UpdateAvailability)
			r.Patch("/mentors/me", mentorHandler.UpdateProfile)
			r.Put("/mentors/templates", mentorHandler.SaveProtocolTemplates)
			// Mentors & Recommendations
			r.Get("/mentors/me", mentorHandler.GetMyProfile)
			r.Get("/recommendations/mentors", recHandler.ListRecommendedMentors)

			// Payment Routes
			r.Post("/payments/create-order", paymentHandler.CreateOrder)
			r.Post("/payments/{id}/verify", paymentHandler.VerifyPayment)

			// Booking Routes
			r.Post("/bookings", bookingHandler.CreateBooking)
			r.Get("/bookings/{id}", bookingHandler.GetBooking)
			r.Get("/bookings/student", bookingHandler.ListStudentBookings)
			r.Get("/bookings/mentor", bookingHandler.ListMentorBookings)
			r.Patch("/bookings/{id}/status", bookingHandler.UpdateBookingStatus)
			r.Post("/bookings/{id}/review", reviewHandler.CreateReview)
			
			// Chat Routes
			r.Post("/bookings/{id}/chat", chatHandler.SendMessage)
			r.Get("/bookings/{id}/chat", chatHandler.GetMessages)

			// Admin Routes (Role Restricted)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RoleGuard("admin"))
				r.Get("/admin/stats", adminHandler.GetDashboardStats)
				r.Get("/admin/mentors/pending", adminHandler.ListPendingMentors)
				r.Patch("/admin/mentors/{id}/verify", adminHandler.VerifyMentor)
				r.Get("/admin/payments", adminHandler.ListPayments)
				r.Get("/admin/analytics/searches", analyticsHandler.GetTopSearches)
				r.Get("/admin/analytics/gaps", analyticsHandler.GetZeroResultSearches)
				r.Get("/admin/analytics/trends", analyticsHandler.GetTrends)
				r.Get("/admin/alerts/gap-alerts", analyticsHandler.GetAlerts)
			})
		})
	})
}

func (s *Server) Start() error {
	server := &http.Server{
		Addr:         ":" + s.Config.Port,
		Handler:      s.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		slog.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Server forced to shutdown", "error", err)
		}
	}()

	slog.Info(fmt.Sprintf("Server starting on port %s", s.Config.Port))
	return server.ListenAndServe()
}

func SetupNotificationService(db *pgxpool.Pool) *notifications.Service {
	bookingRepo := bookings.NewRepository(db)
	return notifications.NewService(bookingRepo)
}
