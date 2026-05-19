package handler 

import(
	"encoding/json"
	"errors"
    "net/http"

    "github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
    "github.com/vexyruu/LIP/listing-service/model"
    "github.com/vexyruu/LIP/listing-service/store"
    "github.com/vexyruu/LIP/shared/publisher"
)

// Create handler struct that could access main.go's pool and pub
type Handler struct {
	Pool *pgxpool.Pool
	Pub *publisher.Publisher
	Topic string
}

func (h *Handler) CreateListing(w http.ResponseWriter, r *http.Request) {
	var req model.CreateListingRequest

	// Validate request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Create listing in database
	id, err := store.CreateListing(r.Context(), h.Pool, &req)
	if err != nil {
		http.Error(w, "failed to create listing", http.StatusInternalServerError)
		return
	}

	// Publish listing to Pub/Sub
	payload, _ := json.Marshal(map[string]any{
		"listing_id": id,
		"user_id": req.UserID,
		"title": req.Title,
		"description": req.Description,
		"condition": req.Condition,
		"category_id": req.CategoryID,
		"price_ask": req.PriceAsk,
	})
	if err := h.Pub.Publish(r.Context(), h.Topic, payload); err != nil {
		http.Error(w, "failed to publish listing", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(model.CreateListingResponse{
		ListingID: id,
		Status: "processing",
		ETA: 500,
	})

}


// GetListing retrieves a listing from the database
func (h *Handler) GetListing(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")

    listing, err := store.GetListing(r.Context(), h.Pool, id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            http.Error(w, "listing not found", http.StatusNotFound)
            return
        }
        http.Error(w, "failed to get listing", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(listing)
}