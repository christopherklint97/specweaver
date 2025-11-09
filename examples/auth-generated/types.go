package api

type Error struct {
	// Error type
	Error string `json:"error"`
	// Detailed error message
	Message string `json:"message,omitempty"`
}

type Resource struct {
	// Resource ID
	Id int64 `json:"id"`
	// Resource name
	Name string `json:"name"`
	// Owner user ID
	OwnerId int64 `json:"ownerId,omitempty"`
}

type User struct {
	// Email address
	Email string `json:"email"`
	// User ID
	Id int64 `json:"id"`
	// User role
	Role string `json:"role,omitempty"`
	// Username
	Username string `json:"username"`
}

