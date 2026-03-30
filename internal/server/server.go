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
	"github.com/ashparshp/mentormatch-backend/internal/modules/bookings"
	"github.com/ashparshp/mentormatch-backend/internal/modules/mentors"
	"github.com/ashparshp/mentormatch-backend/internal/modules/resources"
	"github.com/ashparshp/mentormatch-backend/internal/modules/reviews"
	"github.com/ashparshp/mentormatch-backend/internal/modules/users"
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

	// Services
	userService := users.NewService(userRepo)
	mentorService := mentors.NewService(mentorRepo)
	reviewService := reviews.NewService(reviewRepo, bookingRepo)
	bookingService := bookings.NewService(bookingRepo, mentorRepo)

	// Handlers
	userHandler := users.NewHandler(userService)
	mentorHandler := mentors.NewHandler(mentorService)
	bookingHandler := bookings.NewHandler(bookingService)
	resourceHandler := resources.NewHandler(resourceRepo)
	reviewHandler := reviews.NewHandler(reviewService)

	// Public Routes
	s.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	s.Router.Route("/api/v1", func(r chi.Router) {
		// Public Mentor Routes
		r.Get("/mentors", mentorHandler.DiscoverMentors)
		r.Get("/mentors/{id}", mentorHandler.GetMentor)
		r.Get("/mentors/{id}/reviews", reviewHandler.GetMentorReviews)
		r.Get("/resources", resourceHandler.ListResources)

		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(s.Config.JWTSecret))

			// User Routes
			r.Get("/me", userHandler.GetProfile)
			r.Put("/me", userHandler.UpdateProfile)
			r.Post("/onboard/student", userHandler.OnboardStudent)

			// Mentor Routes (Authenticated)
			r.Post("/onboard/mentor", mentorHandler.OnboardMentor)
			r.Put("/mentors/availability", mentorHandler.UpdateAvailability)

			// Booking Routes
			r.Post("/bookings", bookingHandler.CreateBooking)
			r.Get("/bookings/{id}", bookingHandler.GetBooking)
			r.Get("/bookings/student", bookingHandler.ListStudentBookings)
			r.Get("/bookings/mentor", bookingHandler.ListMentorBookings)
			r.Patch("/bookings/{id}/status", bookingHandler.UpdateBookingStatus)
			r.Post("/bookings/{id}/review", reviewHandler.CreateReview)
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
