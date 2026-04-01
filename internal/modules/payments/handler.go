package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ashparshp/mentormatch-backend/internal/config"
	"github.com/ashparshp/mentormatch-backend/internal/middleware"
	"github.com/ashparshp/mentormatch-backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	repo     Repository
	validate *validator.Validate
	cfg      *config.Config
}

func NewHandler(repo Repository, cfg *config.Config) *Handler {
	return &Handler{
		repo:     repo,
		validate: validator.New(),
		cfg:      cfg,
	}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR")
		return
	}

	// Real Razorpay Order Logic
	orderID, err := h.createRazorpayOrder(req.Amount, req.Currency)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create Razorpay order", "PAYMENT_API_ERROR")
		return
	}

	payment := &Payment{
		BookingID:       req.BookingID,
		StudentID:       userID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          StatusPending,
		Provider:        "razorpay",
		RazorpayOrderID: orderID,
	}

	if err := h.repo.CreatePayment(r.Context(), payment); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to store payment record", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusCreated, payment, "Razorpay order created")
}

func (h *Handler) VerifyPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "id")
	if paymentID == "" {
		response.Error(w, http.StatusBadRequest, "Missing payment internal ID", "BAD_REQUEST")
		return
	}

	var req VerifyPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR")
		return
	}

	// HMAC-SHA256 Signature Verification
	if !h.verifySignature(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature) {
		response.Error(w, http.StatusUnauthorized, "Invalid payment signature", "PAYMENT_SIGNATURE_ERROR")
		return
	}

	if err := h.repo.UpdatePaymentStatus(r.Context(), paymentID, StatusSucceeded, req.RazorpayPaymentID, req.RazorpaySignature); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update payment status", "INTERNAL_ERROR")
		return
	}

	response.Success(w, http.StatusOK, nil, "Payment verified successfully")
}

func (h *Handler) createRazorpayOrder(amount float64, currency string) (string, error) {
	if h.cfg.RazorpayKeyID == "" || h.cfg.RazorpayKeySecret == "" {
		return "", fmt.Errorf("razorpay credentials missing")
	}

	// Amount in cents/paise (e.g. 500.00 -> 50000)
	amountInPaise := int(amount * 100)
	
	payload := map[string]interface{}{
		"amount":   amountInPaise,
		"currency": currency,
		"receipt":  fmt.Sprintf("rcpt_%d", amountInPaise),
	}
	
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.razorpay.com/v1/orders", strings.NewReader(string(body)))
	req.SetBasicAuth(h.cfg.RazorpayKeyID, h.cfg.RazorpayKeySecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("razorpay error: %s", string(resBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	
	return result.ID, nil
}

func (h *Handler) verifySignature(orderID, paymentID, signature string) bool {
	message := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(h.cfg.RazorpayKeySecret))
	mac.Write([]byte(message))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
