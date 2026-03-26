package netex

// ServiceJourney represents a parsed NeTEx ServiceJourney.
type ServiceJourney struct {
	ID                string // e.g., "it:apb:ServiceJourney:2292401-Pfeld-10-1-44100:T3:"
	DepartureTime     string // HH:MM:SS
	DayTypeRef        string // e.g., "it:apb:DayType:T3_1:"
	JourneyPatternRef string // e.g., "it:apb:ServiceJourneyPattern:222401.26a100206:"
	OperatorRef       string
	TransportMode     string
}

// JourneyPattern represents a parsed NeTEx ServiceJourneyPattern.
type JourneyPattern struct {
	ID       string // e.g., "it:apb:ServiceJourneyPattern:222401.26a100206:"
	RouteRef string // e.g., "it:apb:Route:22-240-1-26a-1-2/H:"
}

// Route represents a parsed NeTEx Route.
type Route struct {
	ID           string // e.g., "it:apb:Route:22-240-1-26a-1-2/H:"
	LineRef      string // e.g., "it:apb:Line:222401.26a:"
	DirectionRef string // e.g., "it:apb:Direction:H:" or "it:apb:Direction:R:"
}

// Line represents a parsed NeTEx Line.
type Line struct {
	ID         string // e.g., "it:apb:Line:222401.26a:"
	PublicCode string // e.g., "240"
}

// DayType represents a parsed NeTEx DayType.
type DayType struct {
	ID   string // e.g., "it:apb:DayType:T3_1:"
	Name string // e.g., "domenica e festivi"
}

// LineInfo holds resolved line information for a ServiceJourney.
type LineInfo struct {
	PublicCode    string // GTFS route_short_name equivalent
	Direction     string // "H" or "R"
	DepartureTime string // HH:MM:SS (first stop departure)
	DayTypeRef    string
}
