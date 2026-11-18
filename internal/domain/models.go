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

// Cuban provinces and municipalities (nationwide coverage).
var Provinces = map[string][]string{
	"Pinar del Río": {"Consolación del Sur", "Guane", "La Palma", "Los Palacios", "Mantua", "Minas de Matahambre", "Pinar del Río", "San Juan y Martínez", "San Luis", "Sandino", "Viñales"},
	"Artemisa": {"Alquízar", "Artemisa", "Bahía Honda", "Bauta", "Caimito", "Candelaria", "Guanajay", "Güira de Melena", "Mariel", "San Antonio de los Baños", "San Cristóbal"},
	"La Habana": {"Arroyo Naranjo", "Boyeros", "Centro Habana", "Cerro", "Cotorro", "Diez de Octubre", "Guanabacoa", "Habana del Este", "Habana Vieja", "La Lisa", "Marianao", "Playa", "Plaza de la Revolución", "Regla", "San Miguel del Padrón"},
	"Mayabeque": {"Batabanó", "Bejucal", "Güines", "Jaruco", "Madruga", "Melena del Sur", "Nueva Paz", "Quivicán", "San José de las Lajas", "San Nicolás de Bari", "Santa Cruz del Norte"},
	"Matanzas": {"Calimete", "Cárdenas", "Ciénaga de Zapata", "Colón", "Jagüey Grande", "Jovellanos", "Limonar", "Los Arabos", "Martí", "Matanzas", "Pedro Betancourt", "Perico", "Unión de Reyes"},
	"Cienfuegos": {"Abreus", "Aguada de Pasajeros", "Cienfuegos", "Cruces", "Cumanayagua", "Lajas", "Palmira", "Rodas"},
	"Villa Clara": {"Caibarién", "Camajuaní", "Cifuentes", "Corralillo", "Encrucijada", "Manicaragua", "Placetas", "Quemado de Güines", "Ranchuelo", "Remedios", "Sagua la Grande", "Santa Clara", "Santo Domingo"},
	"Sancti Spíritus": {"Cabaiguán", "Fomento", "Jatibonico", "La Sierpe", "Sancti Spíritus", "Taguasco", "Trinidad", "Yaguajay"},
	"Ciego de Ávila": {"Baraguá", "Bolivia", "Chambas", "Ciego de Ávila", "Ciro Redondo", "Florencia", "Majagua", "Morón", "Primero de Enero", "Venezuela"},
	"Camagüey": {"Camagüey", "Carlos Manuel de Céspedes", "Esmeralda", "Florida", "Guáimaro", "Jimaguayú", "Minas", "Najasa", "Nuevitas", "Santa Cruz del Sur", "Sibanicú", "Sierra de Cubitas", "Vertientes"},
	"Las Tunas": {"Amancio", "Colombia", "Jesús Menéndez", "Jobabo", "Las Tunas", "Majibacoa", "Manatí", "Puerto Padre"},
	"Holguín": {"Antilla", "Báguanos", "Banes", "Cacocum", "Calixto García", "Cueto", "Frank País", "Gibara", "Holguín", "Mayarí", "Moa", "Rafael Freyre", "Sagua de Tánamo", "Urbano Noris"},
	"Granma": {"Bartolomé Masó", "Bayamo", "Buey Arriba", "Campechuela", "Cauto Cristo", "Guisa", "Jiguaní", "Manzanillo", "Media Luna", "Niquero", "Pilón", "Río Cauto", "Yara"},
	"Santiago de Cuba": {"Contramaestre", "Guamá", "Julio Antonio Mella", "Palma Soriano", "San Luis", "Santiago de Cuba", "Segundo Frente", "Songo-La Maya", "Tercer Frente"},
	"Guantánamo": {"Baracoa", "Caimanera", "El Salvador", "Guantánamo", "Imías", "Maisí", "Manuel Tames", "Niceto Pérez", "San Antonio del Sur", "Yateras"},
	"Isla de la Juventud": {"Isla de la Juventud"},
}

// Municipalities is a flat "Provincia — Municipio" list for simple selects.
var Municipalities = []string{
	"Pinar del Río — Consolación del Sur",
	"Pinar del Río — Guane",
	"Pinar del Río — La Palma",
	"Pinar del Río — Los Palacios",
	"Pinar del Río — Mantua",
	"Pinar del Río — Minas de Matahambre",
	"Pinar del Río — Pinar del Río",
	"Pinar del Río — San Juan y Martínez",
	"Pinar del Río — San Luis",
	"Pinar del Río — Sandino",
	"Pinar del Río — Viñales",
	"Artemisa — Alquízar",
	"Artemisa — Artemisa",
	"Artemisa — Bahía Honda",
	"Artemisa — Bauta",
	"Artemisa — Caimito",
	"Artemisa — Candelaria",
	"Artemisa — Guanajay",
	"Artemisa — Güira de Melena",
	"Artemisa — Mariel",
	"Artemisa — San Antonio de los Baños",
	"Artemisa — San Cristóbal",
	"La Habana — Arroyo Naranjo",
	"La Habana — Boyeros",
	"La Habana — Centro Habana",
	"La Habana — Cerro",
	"La Habana — Cotorro",
	"La Habana — Diez de Octubre",
	"La Habana — Guanabacoa",
	"La Habana — Habana del Este",
	"La Habana — Habana Vieja",
	"La Habana — La Lisa",
	"La Habana — Marianao",
	"La Habana — Playa",
	"La Habana — Plaza de la Revolución",
	"La Habana — Regla",
	"La Habana — San Miguel del Padrón",
	"Mayabeque — Batabanó",
	"Mayabeque — Bejucal",
	"Mayabeque — Güines",
	"Mayabeque — Jaruco",
	"Mayabeque — Madruga",
	"Mayabeque — Melena del Sur",
	"Mayabeque — Nueva Paz",
	"Mayabeque — Quivicán",
	"Mayabeque — San José de las Lajas",
	"Mayabeque — San Nicolás de Bari",
	"Mayabeque — Santa Cruz del Norte",
	"Matanzas — Calimete",
	"Matanzas — Cárdenas",
	"Matanzas — Ciénaga de Zapata",
	"Matanzas — Colón",
	"Matanzas — Jagüey Grande",
	"Matanzas — Jovellanos",
	"Matanzas — Limonar",
	"Matanzas — Los Arabos",
	"Matanzas — Martí",
	"Matanzas — Matanzas",
	"Matanzas — Pedro Betancourt",
	"Matanzas — Perico",
	"Matanzas — Unión de Reyes",
	"Cienfuegos — Abreus",
	"Cienfuegos — Aguada de Pasajeros",
	"Cienfuegos — Cienfuegos",
	"Cienfuegos — Cruces",
	"Cienfuegos — Cumanayagua",
	"Cienfuegos — Lajas",
	"Cienfuegos — Palmira",
	"Cienfuegos — Rodas",
	"Villa Clara — Caibarién",
	"Villa Clara — Camajuaní",
	"Villa Clara — Cifuentes",
	"Villa Clara — Corralillo",
	"Villa Clara — Encrucijada",
	"Villa Clara — Manicaragua",
	"Villa Clara — Placetas",
	"Villa Clara — Quemado de Güines",
	"Villa Clara — Ranchuelo",
	"Villa Clara — Remedios",
	"Villa Clara — Sagua la Grande",
	"Villa Clara — Santa Clara",
	"Villa Clara — Santo Domingo",
	"Sancti Spíritus — Cabaiguán",
	"Sancti Spíritus — Fomento",
	"Sancti Spíritus — Jatibonico",
	"Sancti Spíritus — La Sierpe",
	"Sancti Spíritus — Sancti Spíritus",
	"Sancti Spíritus — Taguasco",
	"Sancti Spíritus — Trinidad",
	"Sancti Spíritus — Yaguajay",
	"Ciego de Ávila — Baraguá",
	"Ciego de Ávila — Bolivia",
	"Ciego de Ávila — Chambas",
	"Ciego de Ávila — Ciego de Ávila",
	"Ciego de Ávila — Ciro Redondo",
	"Ciego de Ávila — Florencia",
	"Ciego de Ávila — Majagua",
	"Ciego de Ávila — Morón",
	"Ciego de Ávila — Primero de Enero",
	"Ciego de Ávila — Venezuela",
	"Camagüey — Camagüey",
	"Camagüey — Carlos Manuel de Céspedes",
	"Camagüey — Esmeralda",
	"Camagüey — Florida",
	"Camagüey — Guáimaro",
	"Camagüey — Jimaguayú",
	"Camagüey — Minas",
	"Camagüey — Najasa",
	"Camagüey — Nuevitas",
	"Camagüey — Santa Cruz del Sur",
	"Camagüey — Sibanicú",
	"Camagüey — Sierra de Cubitas",
	"Camagüey — Vertientes",
	"Las Tunas — Amancio",
	"Las Tunas — Colombia",
	"Las Tunas — Jesús Menéndez",
	"Las Tunas — Jobabo",
	"Las Tunas — Las Tunas",
	"Las Tunas — Majibacoa",
	"Las Tunas — Manatí",
	"Las Tunas — Puerto Padre",
	"Holguín — Antilla",
	"Holguín — Báguanos",
	"Holguín — Banes",
	"Holguín — Cacocum",
	"Holguín — Calixto García",
	"Holguín — Cueto",
	"Holguín — Frank País",
	"Holguín — Gibara",
	"Holguín — Holguín",
	"Holguín — Mayarí",
	"Holguín — Moa",
	"Holguín — Rafael Freyre",
	"Holguín — Sagua de Tánamo",
	"Holguín — Urbano Noris",
	"Granma — Bartolomé Masó",
	"Granma — Bayamo",
	"Granma — Buey Arriba",
	"Granma — Campechuela",
	"Granma — Cauto Cristo",
	"Granma — Guisa",
	"Granma — Jiguaní",
	"Granma — Manzanillo",
	"Granma — Media Luna",
	"Granma — Niquero",
	"Granma — Pilón",
	"Granma — Río Cauto",
	"Granma — Yara",
	"Santiago de Cuba — Contramaestre",
	"Santiago de Cuba — Guamá",
	"Santiago de Cuba — Julio Antonio Mella",
	"Santiago de Cuba — Palma Soriano",
	"Santiago de Cuba — San Luis",
	"Santiago de Cuba — Santiago de Cuba",
	"Santiago de Cuba — Segundo Frente",
	"Santiago de Cuba — Songo-La Maya",
	"Santiago de Cuba — Tercer Frente",
	"Guantánamo — Baracoa",
	"Guantánamo — Caimanera",
	"Guantánamo — El Salvador",
	"Guantánamo — Guantánamo",
	"Guantánamo — Imías",
	"Guantánamo — Maisí",
	"Guantánamo — Manuel Tames",
	"Guantánamo — Niceto Pérez",
	"Guantánamo — San Antonio del Sur",
	"Guantánamo — Yateras",
	"Isla de la Juventud — Isla de la Juventud",
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
