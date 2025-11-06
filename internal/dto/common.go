package dto

// Response is the standard API response structure
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResult represents paginated results
type PageResult struct {
	Total   int64       `json:"total"`
	Content interface{} `json:"content"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response with token
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}
