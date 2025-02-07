package model

// Server represents a single server's status as returned by the API.
type Server struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Congestion string `json:"congestion"`
	Creation   string `json:"creation"`
}
