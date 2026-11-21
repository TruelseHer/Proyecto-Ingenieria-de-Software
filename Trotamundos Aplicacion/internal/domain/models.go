package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type TripStatus string

const (
	TripStatusActive    TripStatus = "ACTIVE"
	TripStatusCompleted TripStatus = "COMPLETED"
	TripStatusArchived  TripStatus = "ARCHIVED"
)

type ReservationType string

const (
	ReservationFlight ReservationType = "FLIGHT"
	ReservationHotel  ReservationType = "HOTEL"
	ReservationRental ReservationType = "RENTAL"
	ReservationTrain  ReservationType = "TRAIN"
	ReservationBus    ReservationType = "BUS"
)

type ReservationStatus string

const (
	StatusConfirmed   ReservationStatus = "CONFIRMED"
	StatusDelayed     ReservationStatus = "DELAYED"
	StatusRescheduled ReservationStatus = "RESCHEDULED"
	StatusCancelled   ReservationStatus = "CANCELLED"
)

type ReservationSource string

const (
	SourceManual      ReservationSource = "MANUAL"
	SourceCatalog     ReservationSource = "CATALOG"
	SourceEmailScan   ReservationSource = "EMAIL_SCAN"
	SourceProviderAPI ReservationSource = "PROVIDER_API"
)

type User struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"-"`
	Role              string    `json:"role"`
	HomeCurrency      string    `json:"homeCurrency"`
	PreferredAgencyID string    `json:"preferredAgencyId"`
	MailboxAuthorized bool      `json:"mailboxAuthorized"`
	ConnectedEmail    string    `json:"connectedEmail"`
	AvatarInitials    string    `json:"avatarInitials"`
	CreatedAt         time.Time `json:"createdAt"`
}

type TravelAgency struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	SupportEmail string `json:"supportEmail"`
	Phone        string `json:"phone"`
	SLAHours     int    `json:"slaHours"`
	Specialty    string `json:"specialty"`
}

type SupportTicket struct {
	ID            string    `json:"id"`
	ReferenceCode string    `json:"referenceCode"`
	ReservationID string    `json:"reservationId"`
	TripID        string    `json:"tripId"`
	UserID        string    `json:"userId"`
	AgencyID      string    `json:"agencyId"`
	AgencyName    string    `json:"agencyName"`
	IssueSummary  string    `json:"issueSummary"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ReservationDetails struct {
	// Flight
	FlightNumber     string `json:"flightNumber,omitempty"`
	DepartureAirport string `json:"departureAirport,omitempty"`
	ArrivalAirport   string `json:"arrivalAirport,omitempty"`
	Terminal         string `json:"terminal,omitempty"`
	Gate             string `json:"gate,omitempty"`
	CabinClass       string `json:"cabinClass,omitempty"`

	// Hotel
	HotelName    string  `json:"hotelName,omitempty"`
	Address      string  `json:"address,omitempty"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
	RoomType     string  `json:"roomType,omitempty"`
	CheckInTime  string  `json:"checkInTime,omitempty"`
	CheckOutTime string  `json:"checkOutTime,omitempty"`

	// Vehicle Rental
	VehicleCategory   string `json:"vehicleCategory,omitempty"`
	PickupLocation    string `json:"pickupLocation,omitempty"`
	DropoffLocation   string `json:"dropoffLocation,omitempty"`
	InsuranceIncluded bool   `json:"insuranceIncluded,omitempty"`

	// Train
	TrainLine        string `json:"trainLine,omitempty"`
	TrainNumber      string `json:"trainNumber,omitempty"`
	DepartureStation string `json:"departureStation,omitempty"`
	ArrivalStation   string `json:"arrivalStation,omitempty"`
	Wagon            string `json:"wagon,omitempty"`

	// Bus
	BusCompany        string `json:"busCompany,omitempty"`
	LineNumber        string `json:"lineNumber,omitempty"`
	DepartureTerminal string `json:"departureTerminal,omitempty"`
	ArrivalTerminal   string `json:"arrivalTerminal,omitempty"`
	Platform          string `json:"platform,omitempty"`

	// Shared seat info (Flight / Train / Bus)
	Seat string `json:"seat,omitempty"`
}

type Reservation struct {
	ID               string             `json:"id"`
	TripID           string             `json:"tripId"`
	Type             ReservationType    `json:"type"`
	Status           ReservationStatus  `json:"status"`
	ProviderName     string             `json:"providerName"`
	ConfirmationCode string             `json:"confirmationCode"`
	Origin           string             `json:"origin"`
	Destination      string             `json:"destination"`
	StartDate        time.Time          `json:"startDate"`
	EndDate          time.Time          `json:"endDate"`
	Price            float64            `json:"price"`
	Currency         string             `json:"currency"`
	Notes            string             `json:"notes"`
	Source           ReservationSource  `json:"source"`
	ManuallyModified bool               `json:"manuallyModified"`
	ModifiedFields   []string           `json:"modifiedFields"`
	Details          ReservationDetails `json:"details"`
	SupportTicket    *SupportTicket     `json:"supportTicket,omitempty"`
	CreatedAt        time.Time          `json:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt"`
}

type ReservationUpdateInput struct {
	ProviderName     *string
	ConfirmationCode *string
	Origin           *string
	Destination      *string
	StartDate        *time.Time
	EndDate          *time.Time
	Status           *ReservationStatus
	Price            *float64
	Currency         *string
	Notes            *string
	Details          *ReservationDetails
}

func (r *Reservation) HasModifiedField(field string) bool {
	for _, f := range r.ModifiedFields {
		if f == field {
			return true
		}
	}
	return false
}

func (r *Reservation) addModifiedField(field string) {
	if !r.HasModifiedField(field) {
		r.ModifiedFields = append(r.ModifiedFields, field)
	}
}

func (r *Reservation) ApplyManualEdit(in ReservationUpdateInput) {
	if in.ProviderName != nil && strings.TrimSpace(*in.ProviderName) != "" && *in.ProviderName != r.ProviderName {
		r.ProviderName = strings.TrimSpace(*in.ProviderName)
		r.addModifiedField("providerName")
	}
	if in.ConfirmationCode != nil && strings.TrimSpace(*in.ConfirmationCode) != "" && *in.ConfirmationCode != r.ConfirmationCode {
		r.ConfirmationCode = strings.TrimSpace(*in.ConfirmationCode)
		r.addModifiedField("confirmationCode")
	}
	if in.Origin != nil && strings.TrimSpace(*in.Origin) != "" && *in.Origin != r.Origin {
		r.Origin = strings.TrimSpace(*in.Origin)
		r.addModifiedField("origin")
	}
	if in.Destination != nil && strings.TrimSpace(*in.Destination) != "" && *in.Destination != r.Destination {
		r.Destination = strings.TrimSpace(*in.Destination)
		r.addModifiedField("destination")
	}
	if in.StartDate != nil && !in.StartDate.IsZero() && !in.StartDate.Equal(r.StartDate) {
		r.StartDate = *in.StartDate
		r.addModifiedField("startDate")
	}
	if in.EndDate != nil && !in.EndDate.IsZero() && !in.EndDate.Equal(r.EndDate) {
		r.EndDate = *in.EndDate
		r.addModifiedField("endDate")
	}
	if in.Status != nil && *in.Status != "" && *in.Status != r.Status {
		r.Status = *in.Status
		r.addModifiedField("status")
	}
	if in.Price != nil && *in.Price >= 0 && *in.Price != r.Price {
		r.Price = *in.Price
		r.addModifiedField("price")
	}
	if in.Currency != nil && strings.TrimSpace(*in.Currency) != "" && *in.Currency != r.Currency {
		r.Currency = strings.ToUpper(strings.TrimSpace(*in.Currency))
		r.addModifiedField("currency")
	}
	if in.Notes != nil && *in.Notes != r.Notes {
		r.Notes = strings.TrimSpace(*in.Notes)
		r.addModifiedField("notes")
	}
	if in.Details != nil {
		r.Details = mergeDetails(r.Details, *in.Details)
		r.addModifiedField("details")
	}
	r.ManuallyModified = true
	r.UpdatedAt = time.Now().UTC()
}

func (r *Reservation) ApplyExternalUpdate(status ReservationStatus, startDate, endDate time.Time, notes string, details ReservationDetails) (applied []string, skipped []string) {
	if status != "" && status != r.Status {
		if r.ManuallyModified && r.HasModifiedField("status") {
			skipped = append(skipped, "status")
		} else {
			r.Status = status
			applied = append(applied, "status")
		}
	}
	if !startDate.IsZero() && !startDate.Equal(r.StartDate) {
		if r.ManuallyModified && r.HasModifiedField("startDate") {
			skipped = append(skipped, "startDate")
		} else {
			r.StartDate = startDate
			applied = append(applied, "startDate")
		}
	}
	if !endDate.IsZero() && !endDate.Equal(r.EndDate) {
		if r.ManuallyModified && r.HasModifiedField("endDate") {
			skipped = append(skipped, "endDate")
		} else {
			r.EndDate = endDate
			applied = append(applied, "endDate")
		}
	}
	if notes != "" && notes != r.Notes {
		if r.ManuallyModified && r.HasModifiedField("notes") {
			skipped = append(skipped, "notes")
		} else {
			r.Notes = notes
			applied = append(applied, "notes")
		}
	}
	if details != (ReservationDetails{}) {
		if r.ManuallyModified && r.HasModifiedField("details") {
			skipped = append(skipped, "details")
		} else {
			r.Details = mergeDetails(r.Details, details)
			applied = append(applied, "details")
		}
	}
	r.UpdatedAt = time.Now().UTC()
	return applied, skipped
}

func mergeDetails(base, patch ReservationDetails) ReservationDetails {
	if patch.FlightNumber != "" {
		base.FlightNumber = patch.FlightNumber
	}
	if patch.DepartureAirport != "" {
		base.DepartureAirport = patch.DepartureAirport
	}
	if patch.ArrivalAirport != "" {
		base.ArrivalAirport = patch.ArrivalAirport
	}
	if patch.Terminal != "" {
		base.Terminal = patch.Terminal
	}
	if patch.Gate != "" {
		base.Gate = patch.Gate
	}
	if patch.CabinClass != "" {
		base.CabinClass = patch.CabinClass
	}
	if patch.HotelName != "" {
		base.HotelName = patch.HotelName
	}
	if patch.Address != "" {
		base.Address = patch.Address
	}
	if patch.Latitude != 0 {
		base.Latitude = patch.Latitude
	}
	if patch.Longitude != 0 {
		base.Longitude = patch.Longitude
	}
	if patch.RoomType != "" {
		base.RoomType = patch.RoomType
	}
	if patch.CheckInTime != "" {
		base.CheckInTime = patch.CheckInTime
	}
	if patch.CheckOutTime != "" {
		base.CheckOutTime = patch.CheckOutTime
	}
	if patch.VehicleCategory != "" {
		base.VehicleCategory = patch.VehicleCategory
	}
	if patch.PickupLocation != "" {
		base.PickupLocation = patch.PickupLocation
	}
	if patch.DropoffLocation != "" {
		base.DropoffLocation = patch.DropoffLocation
	}
	if patch.InsuranceIncluded {
		base.InsuranceIncluded = true
	}
	if patch.TrainLine != "" {
		base.TrainLine = patch.TrainLine
	}
	if patch.TrainNumber != "" {
		base.TrainNumber = patch.TrainNumber
	}
	if patch.DepartureStation != "" {
		base.DepartureStation = patch.DepartureStation
	}
	if patch.ArrivalStation != "" {
		base.ArrivalStation = patch.ArrivalStation
	}
	if patch.Wagon != "" {
		base.Wagon = patch.Wagon
	}
	if patch.BusCompany != "" {
		base.BusCompany = patch.BusCompany
	}
	if patch.LineNumber != "" {
		base.LineNumber = patch.LineNumber
	}
	if patch.DepartureTerminal != "" {
		base.DepartureTerminal = patch.DepartureTerminal
	}
	if patch.ArrivalTerminal != "" {
		base.ArrivalTerminal = patch.ArrivalTerminal
	}
	if patch.Platform != "" {
		base.Platform = patch.Platform
	}
	if patch.Seat != "" {
		base.Seat = patch.Seat
	}
	return base
}

func (r *Reservation) TypeLabel() string {
	switch r.Type {
	case ReservationFlight:
		return "Vuelo"
	case ReservationHotel:
		return "Alojamiento"
	case ReservationRental:
		return "Alquiler de auto"
	case ReservationTrain:
		return "Tren"
	case ReservationBus:
		return "Autobús"
	default:
		return "Reserva"
	}
}

func (r *Reservation) StatusLabel() string {
	switch r.Status {
	case StatusConfirmed:
		return "Confirmada"
	case StatusDelayed:
		return "Demorada"
	case StatusRescheduled:
		return "Reprogramada"
	case StatusCancelled:
		return "Cancelada"
	default:
		return string(r.Status)
	}
}

func (r *Reservation) SourceLabel() string {
	switch r.Source {
	case SourceCatalog:
		return "Catálogo Trotamundos"
	case SourceEmailScan:
		return "Sincronizada por correo"
	case SourceProviderAPI:
		return "Importada por código PNR"
	case SourceManual:
		return "Carga manual"
	default:
		return "Reserva"
	}
}

func (r *Reservation) IsDisrupted() bool {
	return r.Status == StatusDelayed || r.Status == StatusRescheduled || r.Status == StatusCancelled
}

func (r *Reservation) DurationLabel() string {
	if r.EndDate.IsZero() || r.EndDate.Before(r.StartDate) {
		return "Por confirmar"
	}
	d := r.EndDate.Sub(r.StartDate)
	if r.Type == ReservationHotel || r.Type == ReservationRental {
		days := int(d.Hours() / 24)
		if days < 1 {
			days = 1
		}
		if days == 1 {
			return "1 noche"
		}
		if r.Type == ReservationRental {
			return fmt.Sprintf("%d días", days)
		}
		return fmt.Sprintf("%d noches", days)
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	if hours == 0 {
		return fmt.Sprintf("%dm", mins)
	}
	return fmt.Sprintf("%dh %02dm", hours, mins)
}

func (r *Reservation) LocationSummary() string {
	switch r.Type {
	case ReservationFlight:
		dep := r.Details.DepartureAirport
		if dep == "" {
			dep = r.Origin
		}
		arr := r.Details.ArrivalAirport
		if arr == "" {
			arr = r.Destination
		}
		return fmt.Sprintf("%s → %s", dep, arr)
	case ReservationHotel:
		if r.Details.HotelName != "" {
			return r.Details.HotelName
		}
		return r.Destination
	case ReservationRental:
		pick := r.Details.PickupLocation
		if pick == "" {
			pick = r.Origin
		}
		drop := r.Details.DropoffLocation
		if drop == "" {
			drop = r.Destination
		}
		return fmt.Sprintf("%s → %s", pick, drop)
	case ReservationTrain:
		dep := r.Details.DepartureStation
		if dep == "" {
			dep = r.Origin
		}
		arr := r.Details.ArrivalStation
		if arr == "" {
			arr = r.Destination
		}
		return fmt.Sprintf("%s → %s", dep, arr)
	case ReservationBus:
		dep := r.Details.DepartureTerminal
		if dep == "" {
			dep = r.Origin
		}
		arr := r.Details.ArrivalTerminal
		if arr == "" {
			arr = r.Destination
		}
		return fmt.Sprintf("%s → %s", dep, arr)
	default:
		return fmt.Sprintf("%s → %s", r.Origin, r.Destination)
	}
}

type Trip struct {
	ID                string         `json:"id"`
	OwnerID           string         `json:"ownerId"`
	SharedWithUserIDs []string       `json:"sharedWithUserIds"`
	SharedWithEmails  []string       `json:"sharedWithEmails"`
	Title             string         `json:"title"`
	DestinationCity   string         `json:"destinationCity"`
	CountryCode       string         `json:"countryCode"`
	CoverImageURL     string         `json:"coverImageUrl"`
	Status            TripStatus     `json:"status"`
	StartDate         time.Time      `json:"startDate"`
	EndDate           time.Time      `json:"endDate"`
	Reservations      []*Reservation `json:"reservations"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
}

func (t *Trip) CanAccess(userID string) bool {
	if t.OwnerID == userID {
		return true
	}
	for _, id := range t.SharedWithUserIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func (t *Trip) CanAccessByUser(u User) bool {
	if t.CanAccess(u.ID) {
		return true
	}
	email := strings.ToLower(strings.TrimSpace(u.Email))
	if email == "" {
		return false
	}
	for _, e := range t.SharedWithEmails {
		if strings.ToLower(strings.TrimSpace(e)) == email {
			return true
		}
	}
	return false
}

func (t *Trip) ShareWith(userID string) bool {
	if userID == "" || userID == t.OwnerID {
		return false
	}
	for _, id := range t.SharedWithUserIDs {
		if id == userID {
			return false
		}
	}
	t.SharedWithUserIDs = append(t.SharedWithUserIDs, userID)
	t.UpdatedAt = time.Now().UTC()
	return true
}

func (t *Trip) ShareWithEmail(email string, resolvedUserID string) bool {
	clean := strings.ToLower(strings.TrimSpace(email))
	if clean == "" {
		return false
	}
	added := false
	hasEmail := false
	for _, e := range t.SharedWithEmails {
		if strings.ToLower(strings.TrimSpace(e)) == clean {
			hasEmail = true
			break
		}
	}
	if !hasEmail {
		t.SharedWithEmails = append(t.SharedWithEmails, clean)
		added = true
	}
	if resolvedUserID != "" && resolvedUserID != t.OwnerID {
		if t.ShareWith(resolvedUserID) {
			added = true
		}
	}
	if added {
		t.UpdatedAt = time.Now().UTC()
	}
	return added
}

func (t *Trip) RecalculateBounds() {
	sort.Slice(t.Reservations, func(i, j int) bool {
		return t.Reservations[i].StartDate.Before(t.Reservations[j].StartDate)
	})
	if len(t.Reservations) == 0 {
		return
	}
	minStart := t.Reservations[0].StartDate
	maxEnd := t.Reservations[0].EndDate
	for _, res := range t.Reservations {
		if res.StartDate.Before(minStart) {
			minStart = res.StartDate
		}
		if res.EndDate.After(maxEnd) {
			maxEnd = res.EndDate
		}
	}
	t.StartDate = minStart
	t.EndDate = maxEnd
	t.UpdatedAt = time.Now().UTC()
}

func (t *Trip) EvaluateLifecycle(now time.Time) bool {
	if t.Status != TripStatusActive || len(t.Reservations) == 0 {
		return false
	}
	if !t.EndDate.IsZero() && t.EndDate.Before(now) {
		t.Status = TripStatusArchived
		t.UpdatedAt = now.UTC()
		return true
	}
	return false
}

func (t *Trip) Archive() {
	t.Status = TripStatusArchived
	t.UpdatedAt = time.Now().UTC()
}

func (t *Trip) Restore() {
	t.Status = TripStatusActive
	t.UpdatedAt = time.Now().UTC()
}

func (t *Trip) HasDisruption() bool {
	for _, r := range t.Reservations {
		if r.IsDisrupted() {
			return true
		}
	}
	return false
}

func (t *Trip) TransportTypes() []ReservationType {
	seen := make(map[ReservationType]bool)
	var out []ReservationType
	for _, r := range t.Reservations {
		if !seen[r.Type] {
			seen[r.Type] = true
			out = append(out, r.Type)
		}
	}
	return out
}

func (t *Trip) DurationDays() int {
	if t.EndDate.IsZero() || t.StartDate.IsZero() || t.EndDate.Before(t.StartDate) {
		return 1
	}
	days := int(t.EndDate.Sub(t.StartDate).Hours()/24) + 1
	if days < 1 {
		return 1
	}
	return days
}

func (t *Trip) TotalCost() float64 {
	var sum float64
	for _, r := range t.Reservations {
		sum += r.Price
	}
	return sum
}

func (t *Trip) PrimaryHotel() *Reservation {
	for _, r := range t.Reservations {
		if r.Type == ReservationHotel {
			return r
		}
	}
	return nil
}

// ---- Preloaded Travel Catalog Entities ----

type CatalogPackageDate struct {
	ID             string    `json:"id"`
	Label          string    `json:"label"`
	StartDate      time.Time `json:"startDate"`
	EndDate        time.Time `json:"endDate"`
	SpotsAvailable int       `json:"spotsAvailable"`
	PriceModifier  float64   `json:"priceModifier"`
}

type CatalogSegmentTemplate struct {
	Type              ReservationType    `json:"type"`
	ProviderName      string             `json:"providerName"`
	Origin            string             `json:"origin"`
	Destination       string             `json:"destination"`
	DayOffsetStart    int                `json:"dayOffsetStart"`
	HourStart         int                `json:"hourStart"`
	DurationHours     int                `json:"durationHours"`
	PriceShare        float64            `json:"priceShare"`
	Notes             string             `json:"notes"`
	ConfirmationPrefx string             `json:"confirmationPrefix"`
	Details           ReservationDetails `json:"details"`
}

type CatalogPackage struct {
	ID               string                   `json:"id"`
	Slug             string                   `json:"slug"`
	Title            string                   `json:"title"`
	Subtitle         string                   `json:"subtitle"`
	DestinationCity  string                   `json:"destinationCity"`
	CountryCode      string                   `json:"countryCode"`
	Region           string                   `json:"region"`
	Summary          string                   `json:"summary"`
	ImageURL         string                   `json:"imageUrl"`
	Badge            string                   `json:"badge"`
	DurationDays     int                      `json:"durationDays"`
	PricePerPerson   float64                  `json:"pricePerPerson"`
	Currency         string                   `json:"currency"`
	Rating           float64                  `json:"rating"`
	ReviewsCount     int                      `json:"reviewsCount"`
	IncludesSummary  []string                 `json:"includesSummary"`
	AvailableDates   []CatalogPackageDate     `json:"availableDates"`
	IncludedSegments []CatalogSegmentTemplate `json:"includedSegments"`
}

type CatalogOffer struct {
	ID                 string             `json:"id"`
	Type               ReservationType    `json:"type"`
	ProviderName       string             `json:"providerName"`
	ProviderEmail      string             `json:"providerEmail"`
	Title              string             `json:"title"`
	Subtitle           string             `json:"subtitle"`
	Origin             string             `json:"origin"`
	Destination        string             `json:"destination"`
	DestinationCity    string             `json:"destinationCity"`
	CountryCode        string             `json:"countryCode"`
	StartDate          time.Time          `json:"startDate"`
	EndDate            time.Time          `json:"endDate"`
	Price              float64            `json:"price"`
	Currency           string             `json:"currency"`
	SeatsAvailable     int                `json:"seatsAvailable"`
	Rating             float64            `json:"rating"`
	ImageURL           string             `json:"imageUrl"`
	HighlightTag       string             `json:"highlightTag"`
	ConfirmationPrefix string             `json:"confirmationPrefix"`
	SamplePNR          string             `json:"samplePnr"`
	Details            ReservationDetails `json:"details"`
}

func (o *CatalogOffer) TypeLabel() string {
	r := Reservation{Type: o.Type}
	return r.TypeLabel()
}

// ---- User Inbox & Sync Rules ----

type EmailConfirmation struct {
	ID             string            `json:"id"`
	UserID         string            `json:"userId"`
	FromEmail      string            `json:"fromEmail"`
	ProviderName   string            `json:"providerName"`
	Subject        string            `json:"subject"`
	Body           string            `json:"body"`
	PNR            string            `json:"pnr"`
	Type           ReservationType   `json:"type"`
	Origin         string            `json:"origin"`
	Destination    string            `json:"destination"`
	DestinationCity string           `json:"destinationCity"`
	CountryCode    string            `json:"countryCode"`
	StartDate      time.Time         `json:"startDate"`
	EndDate        time.Time         `json:"endDate"`
	Price          float64           `json:"price"`
	Currency       string            `json:"currency"`
	Details        ReservationDetails `json:"details"`
	Status         string            `json:"status"` // PENDING, IMPORTED, BLOCKED
	StatusReason   string            `json:"statusReason"`
	ImportedTripID string            `json:"importedTripId"`
	ReceivedAt     time.Time         `json:"receivedAt"`
}

type SenderRule struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	Pattern       string    `json:"pattern"`
	ProviderLabel string    `json:"providerLabel"`
	Category      string    `json:"category"`
	Enabled       bool      `json:"enabled"`
	IsDefault     bool      `json:"isDefault"`
	CreatedAt     time.Time `json:"createdAt"`
}

type PendingSyncItem struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	Source       string    `json:"source"`
	ProviderName string    `json:"providerName"`
	Summary      string    `json:"summary"`
	RawPayload   string    `json:"rawPayload"`
	ErrorReason  string    `json:"errorReason"`
	Attempts     int       `json:"attempts"`
	Status       string    `json:"status"`
	NextRetryAt  time.Time `json:"nextRetryAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Recommendation struct {
	ID             string  `json:"id"`
	Strategy       string  `json:"strategy"`
	Title          string  `json:"title"`
	Destination    string  `json:"destination"`
	Category       string  `json:"category"`
	Description    string  `json:"description"`
	EstimatedPrice float64 `json:"estimatedPrice"`
	Currency       string  `json:"currency"`
	Score          int     `json:"score"`
	HighlightTag   string  `json:"highlightTag"`
	ImageURL       string  `json:"imageUrl"`
}

type AnonymizedTripFact struct {
	ID              string            `json:"id"`
	PseudonymID     string            `json:"pseudonymId"`
	EventType       string            `json:"eventType"`
	DestinationCity string            `json:"destinationCity"`
	CountryCode     string            `json:"countryCode"`
	TransportTypes  []ReservationType `json:"transportTypes"`
	Disrupted       bool              `json:"disrupted"`
	ProviderName    string            `json:"providerName"`
	RecordedAt      time.Time         `json:"recordedAt"`
}

type DestinationTrend struct {
	City            string  `json:"city"`
	CountryCode     string  `json:"countryCode"`
	TripVolume      int     `json:"tripVolume"`
	GrowthPercent   float64 `json:"growthPercent"`
	TopTransport    string  `json:"topTransport"`
	SeasonHighlight string  `json:"seasonHighlight"`
}

type ProviderDisruption struct {
	ProviderName    string  `json:"providerName"`
	TransportType   string  `json:"transportType"`
	TotalSegments   int     `json:"totalSegments"`
	DisruptedCount  int     `json:"disruptedCount"`
	OnTimeRate      float64 `json:"onTimeRate"`
	AvgDelayMinutes int     `json:"avgDelayMinutes"`
}

type MonetizationSummary struct {
	TotalCompletedTrips int     `json:"totalCompletedTrips"`
	UniqueTravelersHash int     `json:"uniqueTravelers"`
	PartnerBookings     int     `json:"partnerBookings"`
	EstimatedGrossValue float64 `json:"estimatedGrossValue"`
	TopCorridor         string  `json:"topCorridor"`
	Currency            string  `json:"currency"`
}

type AuditLogEntry struct {
	ID           string    `json:"id"`
	ActorUserID  string    `json:"actorUserId"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resourceType"`
	ResourceID   string    `json:"resourceId"`
	Outcome      string    `json:"outcome"`
	Reason       string    `json:"reason"`
	IPAddress    string    `json:"ipAddress"`
	CreatedAt    time.Time `json:"createdAt"`
}

type WeatherInfo struct {
	TemperatureC  float64 `json:"temperatureC"`
	ConditionText string  `json:"conditionText"`
	WindKmh       float64 `json:"windKmh"`
	Humidity      int     `json:"humidity"`
	Source        string  `json:"source"`
}

type RouteCoord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type NavigationRoute struct {
	TripID              string       `json:"tripId"`
	TripTitle           string       `json:"tripTitle"`
	DestinationCity     string       `json:"destinationCity"`
	CountryCode         string       `json:"countryCode"`
	HotelName           string       `json:"hotelName"`
	HotelAddress        string       `json:"hotelAddress"`
	HotelLat            float64      `json:"hotelLat"`
	HotelLon            float64      `json:"hotelLon"`
	OriginLabel         string       `json:"originLabel"`
	OriginLat           float64      `json:"originLat"`
	OriginLon           float64      `json:"originLon"`
	DistanceKm          float64      `json:"distanceKm"`
	TransitMinutes      int          `json:"transitMinutes"`
	TaxiMinutes         int          `json:"taxiMinutes"`
	DestinationTimezone string       `json:"destinationTimeZone"`
	LocalTimeFormatted  string       `json:"localTimeFormatted"`
	LocalCurrency       string       `json:"localCurrency"`
	TargetCurrency      string       `json:"targetCurrency"`
	ExchangeRate        float64      `json:"exchangeRate"`
	RateSource          string       `json:"rateSource"`
	SampleFareLocal     float64      `json:"sampleFareLocal"`
	SampleFareTarget    float64      `json:"sampleFareTarget"`
	RouteSVGPoints      string       `json:"routeSvgPoints"`
	PolylineCoords      []RouteCoord `json:"polylineCoords"`
	Steps               []string     `json:"steps"`
	Weather             WeatherInfo  `json:"weather"`
	RoutingToken        string       `json:"routingToken"`
	GeneratedInMs       int64        `json:"generatedInMs"`
}
