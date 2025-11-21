package paymenthttp

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/payment/assembler"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/payment"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/utils"
)

// Handler exposes payment endpoints.
type Handler struct {
	service *payment.Service
}

// Allow reuse without chi mounting.
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) { h.createPayment(w, r) }
func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request)    { h.getPayment(w, r) }
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) { h.handleWebhook(w, r) }
func (h *Handler) Refund(w http.ResponseWriter, r *http.Request)        { h.refund(w, r) }
func (h *Handler) GetByBooking(w http.ResponseWriter, r *http.Request)  { h.getByBooking(w, r) }

type webhookResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

func NewHandler(service *payment.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/payments", h.createPayment)
	r.Get("/payments/{id}", h.getPayment)
	r.Get("/payments/by-booking/{booking_id}", h.getByBooking)
	r.Post("/payments/webhook", h.handleWebhook)
	r.Post("/payments/refund", h.refund)
	return r
}

// @Summary Initiate payment
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body dto.PaymentRequest true "Payment payload"
// @Success 201 {object} dto.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /payments [post]
func (h *Handler) createPayment(w http.ResponseWriter, r *http.Request) {
	var req dto.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	cmd, err := assembler.FromPaymentRequest(req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	pay, err := h.service.Initiate(r.Context(), cmd)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resp := assembler.ToResponse(pay)
	resource := utils.NewResource(resp.ID, "payment", "/api/v1/payments/"+resp.ID, resp)
	utils.Respond(w, http.StatusCreated, "payment initiated", resource)
}

// @Summary Payment webhook
// @Tags Payments
// @Accept json
// @Produce json
// @Success 200 {object} webhookResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /payments/webhook [post]
func (h *Handler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req dto.WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid webhook"))
		return
	}
	cmd, err := assembler.FromWebhook(req, string(body))
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	if err := h.service.HandleWebhook(r.Context(), cmd); err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resource := utils.NewResource(req.PaymentID, "payment", "/api/v1/payments/"+req.PaymentID, webhookResponse{
		PaymentID: req.PaymentID,
		Status:    req.Status,
		Message:   "webhook processed",
	})
	utils.Respond(w, http.StatusOK, "webhook processed", resource)
}

// @Summary Refund payment
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body dto.RefundRequest true "Refund payload"
// @Success 200 {object} dto.RefundResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /payments/refund [post]
func (h *Handler) refund(w http.ResponseWriter, r *http.Request) {
	var req dto.RefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	cmd, err := assembler.FromRefundRequest(req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	res, err := h.service.Refund(r.Context(), cmd)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	dtoResp := assembler.ToRefundResponse(res)
	resource := utils.NewResource(dtoResp.ID, "refund", "/api/v1/payments/refund/"+dtoResp.ID, dtoResp)
	utils.Respond(w, http.StatusOK, "refund created", resource)
}

// @Summary Get payment
// @Tags Payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} dto.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /payments/{id} [get]
func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request) {
	paymentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid id"))
		return
	}
	resp, err := h.service.GetPayment(r.Context(), paymentID)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	dtoResp := assembler.ToResponse(resp)
	resource := utils.NewResource(dtoResp.ID, "payment", "/api/v1/payments/"+dtoResp.ID, dtoResp)
	utils.Respond(w, http.StatusOK, "payment retrieved", resource)
}

// @Summary Get payment by booking ID
// @Tags Payments
// @Produce json
// @Param booking_id path string true "Booking ID"
// @Success 200 {object} dto.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /payments/by-booking/{booking_id} [get]
func (h *Handler) getByBooking(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "booking_id"))
	if err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid booking id"))
		return
	}
	resp, err := h.service.GetByBooking(r.Context(), bookingID)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	dtoResp := assembler.ToResponse(resp)
	resource := utils.NewResource(dtoResp.ID, "payment", "/api/v1/payments/"+dtoResp.ID, dtoResp)
	utils.Respond(w, http.StatusOK, "payment retrieved", resource)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	utils.Respond(w, pkgErrors.StatusCode(err), err.Message, err)
}
