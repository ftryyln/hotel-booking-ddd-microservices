package notificationhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	notificationuc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/notification"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/notification/assembler"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
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
// @Success 202 {object} dto.NotificationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications [post]
func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	var req dto.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	cmd, err := assembler.FromRequest(req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	record, err := h.service.Send(r.Context(), cmd)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resp := assembler.ToResponse(record)
	resource := utils.NewResource(resp.ID, "notification", "/api/v1/notifications/"+resp.ID, resp)
	utils.Respond(w, http.StatusAccepted, "notification accepted", resource)
}

// @Summary List notifications (in-memory)
// @Tags Notifications
// @Produce json
// @Param limit query int false "pagination limit (default 50)"
// @Param offset query int false "pagination offset"
// @Success 200 {array} dto.NotificationResponse
// @Router /notifications [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	opts := parseQueryOptions(r)
	items := h.service.List(r.Context(), opts)
	dtoItems := assembler.ToResponses(items)
	var resources []utils.Resource
	for _, n := range dtoItems {
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
	resp := assembler.ToResponse(notification)
	resource := utils.NewResource(resp.ID, "notification", "/api/v1/notifications/"+resp.ID, resp)
	utils.Respond(w, http.StatusOK, "notification retrieved", resource)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	utils.Respond(w, pkgErrors.StatusCode(err), err.Message, err)
}

func parseQueryOptions(r *http.Request) query.Options {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return query.Options{Limit: limit, Offset: offset}
}
