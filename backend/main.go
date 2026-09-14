package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Rating      float64 `json:"rating"`
	Reviews     int     `json:"reviews"`
	Stock       int     `json:"stock"`
	Color       string  `json:"color"`
	Accent      string  `json:"accent"`
	Description string  `json:"description"`
	Featured    bool    `json:"featured"`
}
type OrderItem struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price,omitempty"`
}
type Order struct {
	ID              string      `json:"id"`
	Email           string      `json:"email"`
	CustomerName    string      `json:"customer_name"`
	ShippingAddress string      `json:"shipping_address"`
	Items           []OrderItem `json:"items"`
	Subtotal        float64     `json:"subtotal"`
	Shipping        float64     `json:"shipping"`
	Total           float64     `json:"total"`
	Status          string      `json:"status"`
	CreatedAt       time.Time   `json:"created_at"`
}
type Store struct {
	mu       sync.RWMutex
	Products []Product `json:"products"`
	Orders   []Order   `json:"orders"`
	path     string
}

func seed() []Product {
	return []Product{
		{1, "Form Carryall", "form-carryall", "Bags", 128, 4.9, 84, 14, "#d8a48f", "#6f3928", "Soft-grain recycled leather with an adaptable interior and magnetic close.", true},
		{2, "Ridge Runner", "ridge-runner", "Footwear", 164, 4.8, 61, 9, "#b8c0aa", "#4f5e49", "A light everyday trainer made with bio-based foam and recycled mesh.", true},
		{3, "Arc Desk Light", "arc-desk-light", "Home", 89, 4.7, 39, 22, "#e4c878", "#76601b", "A dimmable, warm-spectrum task light in powder-coated aluminium.", true},
		{4, "Field Automatic", "field-automatic", "Watches", 240, 4.9, 47, 6, "#9ba7b5", "#394650", "A compact automatic field watch with sapphire glass and woven strap.", true},
		{5, "Sora Bottle", "sora-bottle", "Everyday", 42, 4.8, 112, 35, "#a8c4bd", "#315e58", "Double-wall steel, a ceramic-lined interior, and a leakproof cap.", false},
		{6, "Contour Sling", "contour-sling", "Bags", 74, 4.6, 56, 18, "#c9b9a6", "#665444", "A compact crossbody with a quick-adjust strap and weatherproof shell.", false},
		{7, "Mori Throw", "mori-throw", "Home", 96, 4.9, 28, 11, "#b1a2a0", "#594746", "Soft merino blend woven in a subtle two-tone grid.", false},
		{8, "Daily Notebook", "daily-notebook", "Everyday", 24, 4.7, 93, 48, "#8da1a0", "#314f4e", "Lay-flat binding, fountain-pen paper, and a durable cloth cover.", false},
	}
}
func newStore(path string) *Store {
	s := &Store{Products: seed(), path: path}
	b, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(b, &s.Orders)
	}
	return s
}
func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.Orders, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0644)
}
func (s *Store) product(id int) (Product, bool) {
	for _, p := range s.Products {
		if p.ID == id {
			return p, true
		}
	}
	return Product{}, false
}
func id() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "NS-" + strings.ToUpper(hex.EncodeToString(b))
}
func jsonOut(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func method(m string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != m {
			jsonOut(w, 405, map[string]string{"error": "method not allowed"})
			return
		}
		next(w, r)
	}
}
func (s *Store) listProducts(w http.ResponseWriter, r *http.Request) {
	rows := s.Products
	q := strings.ToLower(r.URL.Query().Get("q"))
	cat := r.URL.Query().Get("category")
	out := make([]Product, 0, len(rows))
	for _, p := range rows {
		if q != "" && !strings.Contains(strings.ToLower(p.Name+" "+p.Category), q) {
			continue
		}
		if cat != "" && !strings.EqualFold(cat, p.Category) {
			continue
		}
		out = append(out, p)
	}
	jsonOut(w, 200, out)
}
func (s *Store) productByID(w http.ResponseWriter, r *http.Request) {
	value := strings.TrimPrefix(r.URL.Path, "/api/products/")
	n, err := strconv.Atoi(value)
	if err != nil {
		jsonOut(w, 400, map[string]string{"error": "invalid product id"})
		return
	}
	p, ok := s.product(n)
	if !ok {
		jsonOut(w, 404, map[string]string{"error": "product not found"})
		return
	}
	jsonOut(w, 200, p)
}
func (s *Store) createOrder(w http.ResponseWriter, r *http.Request) {
	var input Order
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input); err != nil {
		jsonOut(w, 400, map[string]string{"error": "invalid order"})
		return
	}
	if !strings.Contains(input.Email, "@") || input.CustomerName == "" || input.ShippingAddress == "" || len(input.Items) == 0 {
		jsonOut(w, 422, map[string]string{"error": "customer, address, email, and items are required"})
		return
	}
	var subtotal float64
	for i := range input.Items {
		p, ok := s.product(input.Items[i].ProductID)
		if !ok || input.Items[i].Quantity < 1 || input.Items[i].Quantity > p.Stock {
			jsonOut(w, 422, map[string]string{"error": "invalid product or quantity"})
			return
		}
		input.Items[i].UnitPrice = p.Price
		subtotal += p.Price * float64(input.Items[i].Quantity)
	}
	input.ID = id()
	input.Subtotal = subtotal
	if subtotal < 75 {
		input.Shipping = 8
	}
	input.Total = input.Subtotal + input.Shipping
	input.Status = "confirmed"
	input.CreatedAt = time.Now().UTC()
	s.mu.Lock()
	s.Orders = append(s.Orders, input)
	err := s.save()
	s.mu.Unlock()
	if err != nil {
		log.Printf("persist order: %v", err)
	}
	jsonOut(w, 201, input)
}
func (s *Store) listOrders(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Admin-Key") != env("ADMIN_KEY", "northstar-demo-admin") {
		jsonOut(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	jsonOut(w, 200, s.Orders)
}
func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", env("CORS_ORIGIN", "http://localhost:5173"))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func main() {
	store := newStore(env("DATA_FILE", "data/orders.json"))
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", method("GET", func(w http.ResponseWriter, r *http.Request) { jsonOut(w, 200, map[string]string{"status": "ok"}) }))
	mux.HandleFunc("/api/products", method("GET", store.listProducts))
	mux.HandleFunc("/api/products/", method("GET", store.productByID))
	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			store.createOrder(w, r)
			return
		}
		if r.Method == "GET" {
			store.listOrders(w, r)
			return
		}
		jsonOut(w, 405, map[string]string{"error": "method not allowed"})
	})
	dist := env("WEB_DIR", "../frontend/dist")
	if _, err := os.Stat(dist); err == nil {
		mux.Handle("/", http.FileServer(http.Dir(dist)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			jsonOut(w, 404, map[string]string{"error": "web build not found"})
		})
	}
	addr := ":" + env("PORT", "8080")
	server := &http.Server{Addr: addr, Handler: cors(mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("Northstar listening on %s", addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(fmt.Errorf("server: %w", err))
	}
}
