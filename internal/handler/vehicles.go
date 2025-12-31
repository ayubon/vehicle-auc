package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VehicleHandler struct {
	db       *pgxpool.Pool
	logger   *slog.Logger
	validate *validator.Validate
}

func NewVehicleHandler(db *pgxpool.Pool, logger *slog.Logger) *VehicleHandler {
	return &VehicleHandler{
		db:       db,
		logger:   logger,
		validate: validator.New(),
	}
}

type VehicleResponse struct {
	ID            int64   `json:"id"`
	SellerID      int64   `json:"seller_id"`
	VIN           string  `json:"vin"`
	Year          int     `json:"year"`
	Make          string  `json:"make"`
	Model         string  `json:"model"`
	Trim          *string `json:"trim,omitempty"`
	Mileage       *int    `json:"mileage,omitempty"`
	ExteriorColor *string `json:"exterior_color,omitempty"`
	StartingPrice string  `json:"starting_price"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
}

// ListVehicles returns paginated vehicles
func (h *VehicleHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse query params
	limit := 20
	offset := 0
	
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	
	// Optional filters
	makeFilter := r.URL.Query().Get("make")
	modelFilter := r.URL.Query().Get("model")
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active"
	}
	
	// Query vehicles
	query := `
		SELECT id, seller_id, vin, year, make, model, trim, mileage, 
		       exterior_color, starting_price, status, created_at
		FROM vehicles
		WHERE status = $1
		  AND ($2 = '' OR make ILIKE $2)
		  AND ($3 = '' OR model ILIKE $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`
	
	rows, err := h.db.Query(ctx, query, status, makeFilter, modelFilter, limit, offset)
	if err != nil {
		h.logger.Error("failed to query vehicles", slog.String("error", err.Error()))
		h.jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	vehicles := make([]VehicleResponse, 0)
	for rows.Next() {
		var v VehicleResponse
		var startingPrice float64
		var createdAt interface{}
		
		err := rows.Scan(
			&v.ID, &v.SellerID, &v.VIN, &v.Year, &v.Make, &v.Model,
			&v.Trim, &v.Mileage, &v.ExteriorColor, &startingPrice,
			&v.Status, &createdAt,
		)
		if err != nil {
			h.logger.Error("failed to scan vehicle", slog.String("error", err.Error()))
			continue
		}
		v.StartingPrice = strconv.FormatFloat(startingPrice, 'f', 2, 64)
		vehicles = append(vehicles, v)
	}
	
	// Get total count
	var total int64
	countQuery := `
		SELECT COUNT(*) FROM vehicles
		WHERE status = $1
		  AND ($2 = '' OR make ILIKE $2)
		  AND ($3 = '' OR model ILIKE $3)
	`
	h.db.QueryRow(ctx, countQuery, status, makeFilter, modelFilter).Scan(&total)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vehicles": vehicles,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"has_more": int64(offset+len(vehicles)) < total,
	})
}

// GetVehicle returns a single vehicle
func (h *VehicleHandler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}
	
	query := `
		SELECT v.id, v.seller_id, v.vin, v.year, v.make, v.model, v.trim,
		       v.body_type, v.exterior_color, v.interior_color, v.mileage,
		       v.engine, v.transmission, v.drivetrain, v.fuel_type,
		       v.title_status, v.condition_grade, v.description,
		       v.starting_price, v.reserve_price, v.buy_now_price,
		       v.location_city, v.location_state, v.location_zip,
		       v.status, v.created_at,
		       u.first_name as seller_first_name, u.last_name as seller_last_name
		FROM vehicles v
		JOIN users u ON v.seller_id = u.id
		WHERE v.id = $1
	`
	
	var vehicle struct {
		VehicleResponse
		BodyType        *string `json:"body_type,omitempty"`
		InteriorColor   *string `json:"interior_color,omitempty"`
		Engine          *string `json:"engine,omitempty"`
		Transmission    *string `json:"transmission,omitempty"`
		Drivetrain      *string `json:"drivetrain,omitempty"`
		FuelType        *string `json:"fuel_type,omitempty"`
		TitleStatus     *string `json:"title_status,omitempty"`
		ConditionGrade  *string `json:"condition_grade,omitempty"`
		Description     *string `json:"description,omitempty"`
		ReservePrice    *string `json:"reserve_price,omitempty"`
		BuyNowPrice     *string `json:"buy_now_price,omitempty"`
		LocationCity    *string `json:"location_city,omitempty"`
		LocationState   *string `json:"location_state,omitempty"`
		LocationZip     *string `json:"location_zip,omitempty"`
		SellerFirstName *string `json:"seller_first_name,omitempty"`
		SellerLastName  *string `json:"seller_last_name,omitempty"`
	}
	
	var startingPrice, reservePrice, buyNowPrice *float64
	var createdAt interface{}
	
	err = h.db.QueryRow(ctx, query, id).Scan(
		&vehicle.ID, &vehicle.SellerID, &vehicle.VIN, &vehicle.Year,
		&vehicle.Make, &vehicle.Model, &vehicle.Trim,
		&vehicle.BodyType, &vehicle.ExteriorColor, &vehicle.InteriorColor,
		&vehicle.Mileage, &vehicle.Engine, &vehicle.Transmission,
		&vehicle.Drivetrain, &vehicle.FuelType, &vehicle.TitleStatus,
		&vehicle.ConditionGrade, &vehicle.Description,
		&startingPrice, &reservePrice, &buyNowPrice,
		&vehicle.LocationCity, &vehicle.LocationState, &vehicle.LocationZip,
		&vehicle.Status, &createdAt,
		&vehicle.SellerFirstName, &vehicle.SellerLastName,
	)
	
	if err != nil {
		h.jsonError(w, "vehicle not found", http.StatusNotFound)
		return
	}
	
	if startingPrice != nil {
		vehicle.StartingPrice = strconv.FormatFloat(*startingPrice, 'f', 2, 64)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vehicle": vehicle,
	})
}

// CreateVehicle creates a new vehicle listing
func (h *VehicleHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	
	var req struct {
		VIN           string  `json:"vin" validate:"required,len=17"`
		Year          int     `json:"year" validate:"required,min=1900,max=2030"`
		Make          string  `json:"make" validate:"required"`
		Model         string  `json:"model" validate:"required"`
		Trim          string  `json:"trim"`
		Mileage       int     `json:"mileage"`
		StartingPrice float64 `json:"starting_price" validate:"required,gt=0"`
		Description   string  `json:"description"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.validate.Struct(req); err != nil {
		h.jsonError(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	query := `
		INSERT INTO vehicles (seller_id, vin, year, make, model, trim, mileage, starting_price, description, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'draft')
		RETURNING id, created_at
	`
	
	var vehicleID int64
	var createdAt interface{}
	err := h.db.QueryRow(ctx, query,
		userID, req.VIN, req.Year, req.Make, req.Model,
		nilIfEmpty(req.Trim), nilIfZero(req.Mileage),
		req.StartingPrice, nilIfEmpty(req.Description),
	).Scan(&vehicleID, &createdAt)
	
	if err != nil {
		h.logger.Error("failed to create vehicle", slog.String("error", err.Error()))
		h.jsonError(w, "failed to create vehicle", http.StatusInternalServerError)
		return
	}
	
	h.logger.Info("vehicle_created",
		slog.Int64("vehicle_id", vehicleID),
		slog.Int64("seller_id", userID),
		slog.String("vin", req.VIN),
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vehicle_id": vehicleID,
		"message":    "Vehicle created successfully",
	})
}

// UpdateVehicle updates a vehicle listing
func (h *VehicleHandler) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	vehicleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}

	// Check ownership and status
	var sellerID int64
	var status string
	err = h.db.QueryRow(ctx, `SELECT seller_id, status FROM vehicles WHERE id = $1`, vehicleID).Scan(&sellerID, &status)
	if err != nil {
		h.jsonError(w, "vehicle not found", http.StatusNotFound)
		return
	}
	if sellerID != userID {
		h.jsonError(w, "not authorized to edit this vehicle", http.StatusForbidden)
		return
	}
	if status == "sold" {
		h.jsonError(w, "cannot edit sold vehicles", http.StatusBadRequest)
		return
	}

	var req struct {
		Year          *int     `json:"year"`
		Make          *string  `json:"make"`
		Model         *string  `json:"model"`
		Trim          *string  `json:"trim"`
		BodyType      *string  `json:"body_type"`
		Engine        *string  `json:"engine"`
		Transmission  *string  `json:"transmission"`
		Drivetrain    *string  `json:"drivetrain"`
		ExteriorColor *string  `json:"exterior_color"`
		InteriorColor *string  `json:"interior_color"`
		Mileage       *int     `json:"mileage"`
		ConditionGrade *string `json:"condition_grade"`
		TitleStatus   *string  `json:"title_status"`
		Description   *string  `json:"description"`
		StartingPrice *float64 `json:"starting_price"`
		ReservePrice  *float64 `json:"reserve_price"`
		BuyNowPrice   *float64 `json:"buy_now_price"`
		LocationCity  *string  `json:"location_city"`
		LocationState *string  `json:"location_state"`
		LocationZip   *string  `json:"location_zip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE vehicles SET
			year = COALESCE($2, year),
			make = COALESCE($3, make),
			model = COALESCE($4, model),
			trim = COALESCE($5, trim),
			body_type = COALESCE($6, body_type),
			engine = COALESCE($7, engine),
			transmission = COALESCE($8, transmission),
			drivetrain = COALESCE($9, drivetrain),
			exterior_color = COALESCE($10, exterior_color),
			interior_color = COALESCE($11, interior_color),
			mileage = COALESCE($12, mileage),
			condition_grade = COALESCE($13, condition_grade),
			title_status = COALESCE($14, title_status),
			description = COALESCE($15, description),
			starting_price = COALESCE($16, starting_price),
			reserve_price = COALESCE($17, reserve_price),
			buy_now_price = COALESCE($18, buy_now_price),
			location_city = COALESCE($19, location_city),
			location_state = COALESCE($20, location_state),
			location_zip = COALESCE($21, location_zip)
		WHERE id = $1
	`

	_, err = h.db.Exec(ctx, query, vehicleID,
		req.Year, req.Make, req.Model, req.Trim, req.BodyType,
		req.Engine, req.Transmission, req.Drivetrain,
		req.ExteriorColor, req.InteriorColor, req.Mileage,
		req.ConditionGrade, req.TitleStatus, req.Description,
		req.StartingPrice, req.ReservePrice, req.BuyNowPrice,
		req.LocationCity, req.LocationState, req.LocationZip,
	)
	if err != nil {
		h.logger.Error("failed to update vehicle", slog.String("error", err.Error()))
		h.jsonError(w, "failed to update vehicle", http.StatusInternalServerError)
		return
	}

	h.logger.Info("vehicle_updated", slog.Int64("vehicle_id", vehicleID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Vehicle updated",
		"vehicle_id": vehicleID,
	})
}

// DeleteVehicle deletes a vehicle listing
func (h *VehicleHandler) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	vehicleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}

	// Check ownership, status, and active auction
	var sellerID int64
	var status string
	var hasActiveAuction bool
	err = h.db.QueryRow(ctx, `
		SELECT v.seller_id, v.status, 
		       EXISTS(SELECT 1 FROM auctions a WHERE a.vehicle_id = v.id AND a.status = 'active')
		FROM vehicles v WHERE v.id = $1
	`, vehicleID).Scan(&sellerID, &status, &hasActiveAuction)
	if err != nil {
		h.jsonError(w, "vehicle not found", http.StatusNotFound)
		return
	}
	if sellerID != userID {
		h.jsonError(w, "not authorized to delete this vehicle", http.StatusForbidden)
		return
	}
	if status == "sold" {
		h.jsonError(w, "cannot delete sold vehicles", http.StatusBadRequest)
		return
	}
	if hasActiveAuction {
		h.jsonError(w, "cannot delete vehicle with active auction", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec(ctx, `DELETE FROM vehicles WHERE id = $1`, vehicleID)
	if err != nil {
		h.logger.Error("failed to delete vehicle", slog.String("error", err.Error()))
		h.jsonError(w, "failed to delete vehicle", http.StatusInternalServerError)
		return
	}

	h.logger.Info("vehicle_deleted", slog.Int64("vehicle_id", vehicleID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Vehicle deleted"})
}

// SubmitVehicle submits a draft vehicle for listing
func (h *VehicleHandler) SubmitVehicle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	vehicleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}

	// Check ownership and required fields
	var sellerID int64
	var status string
	var year, mileage *int
	var vinMake, model *string
	var startingPrice *float64
	err = h.db.QueryRow(ctx, `
		SELECT seller_id, status, year, make, model, starting_price, mileage
		FROM vehicles WHERE id = $1
	`, vehicleID).Scan(&sellerID, &status, &year, &vinMake, &model, &startingPrice, &mileage)
	if err != nil {
		h.jsonError(w, "vehicle not found", http.StatusNotFound)
		return
	}
	if sellerID != userID {
		h.jsonError(w, "not authorized", http.StatusForbidden)
		return
	}
	if status != "draft" {
		h.jsonError(w, "only draft vehicles can be submitted", http.StatusBadRequest)
		return
	}
	if year == nil || vinMake == nil || model == nil || startingPrice == nil {
		h.jsonError(w, "missing required fields (year, make, model, starting_price)", http.StatusBadRequest)
		return
	}

	// Update to active
	_, err = h.db.Exec(ctx, `UPDATE vehicles SET status = 'active' WHERE id = $1`, vehicleID)
	if err != nil {
		h.logger.Error("failed to submit vehicle", slog.String("error", err.Error()))
		h.jsonError(w, "failed to submit vehicle", http.StatusInternalServerError)
		return
	}

	h.logger.Info("vehicle_submitted", slog.Int64("vehicle_id", vehicleID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Vehicle is now active",
		"status":  "active",
	})
}

// GetVehicleImages returns images for a vehicle
func (h *VehicleHandler) GetVehicleImages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	vehicleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}

	rows, err := h.db.Query(ctx, `
		SELECT id, s3_key, url, is_primary, display_order
		FROM vehicle_images WHERE vehicle_id = $1 ORDER BY display_order
	`, vehicleID)
	if err != nil {
		h.jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	images := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var s3Key, url string
		var isPrimary bool
		var displayOrder int
		rows.Scan(&id, &s3Key, &url, &isPrimary, &displayOrder)
		images = append(images, map[string]interface{}{
			"id":            id,
			"s3_key":        s3Key,
			"url":           url,
			"is_primary":    isPrimary,
			"display_order": displayOrder,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"images": images})
}

func (h *VehicleHandler) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nilIfZero(i int) interface{} {
	if i == 0 {
		return nil
	}
	return i
}

