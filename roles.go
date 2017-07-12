package main

// OwnerRole function
func OwnerRole(u *User) {
	u.Owner = true
	u.Admin = true
	u.SModerator = true
	u.Moderator = true
	u.Editor = true
	u.Agora = true
	u.Streamer = true
	u.Recruiter = true
}

// AdminRole function
func AdminRole(u *User) {
	u.Admin = true
	u.SModerator = true
	u.Moderator = true
	u.Editor = true
	u.Agora = true
	u.Streamer = true
	u.Recruiter = true
}

// SModeratorRole function
func SModeratorRole(u *User) {
	u.SModerator = true
	u.Moderator = true
	u.Editor = true
	u.Agora = true
	u.Streamer = true
	u.Recruiter = true
}

// ModeratorRole function
func ModeratorRole(u *User) {
	u.Moderator = true
	u.Editor = true
	u.Agora = true
	u.Streamer = true
	u.Recruiter = true
}

// EditorRole function
func EditorRole(u *User) {
	u.Editor = true
}

// AgoraRole function
func AgoraRole(u *User) {
	u.Agora = true
}

// StreamerRole function
func StreamerRole(u *User) {
	u.Streamer = true
}

// RecruiterRole function
func RecruiterRole(u *User) {
	u.Recruiter = true
}

// ClearRoles function
func ClearRoles(u *User) {
	u.Owner = false
	u.Admin = false
	u.SModerator = false
	u.Moderator = false
	u.Editor = false
	u.Agora = false
	u.Streamer = false
	u.Recruiter = false
}

// CitizenRole function
func CitizenRole(u *User) {
	u.Citizen = true
}
