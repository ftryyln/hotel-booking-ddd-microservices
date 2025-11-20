package notificationhttp

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	notificationuc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/notification"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/utils"
)

// Handler exposes notification endpoint.
type Handler struct {
	service *notificationuc.Service
}

func NewHandler(service *notificationuc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/notifications", h.send)
	r.Get("/notifications", h.list)
	r.Get("/notifications/{id}", h.get)
	return r
}

// @Summary Send notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Param request body dto.NotificationRequest true "Notification payload"
// @Success 202 {object} notificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications [post]
func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	var req dto.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	record, err := h.service.Send(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resource := utils.NewResource(record.ID, "notification", "/api/v1/notifications/"+record.ID, record)
	utils.Respond(w, http.StatusAccepted, "notification accepted", resource)
}

// @Summary List notifications (in-memory)
// @Tags Notifications
// @Produce json
// @Success 200 {array} dto.NotificationResponse
// @Router /notifications [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items := h.service.List(r.Context())
	var resources []utils.Resource
	for _, n := range items {
		resources = append(resources, utils.NewResource(n.ID, "notification", "/api/v1/notifications/"+n.ID, n))
	}
	utils.RespondWithCount(w, http.StatusOK, "notifications listed", resources, len(resources))
}

// @Summary Get notification by ID
// @Tags Notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} dto.NotificationResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /notifications/{id} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	notification, ok := h.service.Get(r.Context(), id)
	if !ok {
		writeError(w, pkgErrors.New("not_found", "notification not found"))
		return
	}
	resource := utils.NewResource(notification.ID, "notification", "/api/v1/notifications/"+notification.ID, notification)
	utils.Respond(w, http.StatusOK, "notification retrieved", resource)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	utils.Respond(w, pkgErrors.StatusCode(err), err.Message, err)
}
