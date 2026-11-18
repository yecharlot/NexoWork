package domain

import "time"

// Lifecycle states for NEXO organisms (domain semantics on top of Core).
const (
	StatusCreated   = "created"
	StatusActive    = "active"
	StatusInProcess = "in_process"
	StatusMatched   = "matched"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)

// Organism kinds (logical keys: nexo.<type>.<id>).
const (
	KindRequest  = "request"
	KindProvider = "provider"
	KindMatch    = "match"
	KindClient   = "client"
)

// Service types offered in Cuba home services.
var ServiceTypes = []string{
	"Limpieza de casa",
	"Limpieza de baños",
	"Cuidado de ancianos",
	"Cuidado de niños",
	"Cocina / comida",
	"Jardinería",
	"Recogida de escombros",
	"Lavado de ropa",
	"Reparaciones menores",
	"Otro",
}

// Cuban municipalities (common set; free text also allowed).
var Municipalities = []string{
	"Plaza de la Revolución", "Centro Habana", "Habana Vieja", "Habana del Este",
	"Regla", "Guanabacoa", "San Miguel del Padrón", "Diez de Octubre",
	"Cerro", "Marianao", "La Lisa", "Boyeros", "Arroyo Naranjo", "Cotorro",
	"Playa", "Otro municipio / provincia",
}

type ServiceRequest struct {
	ID              string    `json:"id"`
	RootCID         string    `json:"root_cid,omitempty"`
	Status          string    `json:"status"`
	ServiceType     string    `json:"service_type"`
	Description     string    `json:"description"`
	Municipality    string    `json:"municipality"`
	PreferredDate   string    `json:"preferred_date,omitempty"`
	ContactName     string    `json:"contact_name"`
	ContactPhone    string    `json:"contact_phone"`
	FromAbroad      bool      `json:"from_abroad"`
	AbroadName      string    `json:"abroad_name,omitempty"`
	AbroadContact   string    `json:"abroad_contact,omitempty"`
	LocalContactName string   `json:"local_contact_name,omitempty"`
	LocalContactPhone string  `json:"local_contact_phone,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ServiceProvider struct {
	ID           string    `json:"id"`
	RootCID      string    `json:"root_cid,omitempty"`
	Status       string    `json:"status"`
	Services     []string  `json:"services"`
	Zone         string    `json:"zone"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Experience   string    `json:"experience,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Match struct {
	ID          string    `json:"id"`
	RootCID     string    `json:"root_cid,omitempty"`
	Status      string    `json:"status"`
	RequestID   string    `json:"request_id"`
	ProviderID  string    `json:"provider_id"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Client struct {
	ID        string    `json:"id"`
	RootCID   string    `json:"root_cid,omitempty"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone,omitempty"`
	Email     string    `json:"email,omitempty"`
	FromAbroad bool     `json:"from_abroad"`
	CreatedAt time.Time `json:"created_at"`
}
