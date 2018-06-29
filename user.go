package main

import "time"

// User struct
type User struct {
	ID string `storm:"id"` // primary key

	Owner      bool `storm:"index"`
	Admin      bool `storm:"index"`
	SModerator bool `storm:"index"`
	Moderator  bool `storm:"index"`
	Editor     bool `storm:"index"`
	Agora      bool `storm:"index"`
	Streamer   bool `storm:"index"`
	Recruiter  bool `storm:"index"`
	Citizen    bool `storm:"index"`

	HistoricalAutoRoles []string
	LatestRoleTimeout   time.Duration
	CurrentAutoRoleID     string
	CurrentAutoRoleName   string
	DisableAutoRole     bool
}

// Init function
func (u *User) Init() {
	ClearRoles(u)
	CitizenRole(u)
}

// SetRole function
func (u *User) SetRole(role string) {

	switch role {

	case "owner":
		OwnerRole(u)

	case "admin":
		AdminRole(u)

	case "smoderator":
		SModeratorRole(u)

	case "moderator":
		ModeratorRole(u)

	case "editor":
		EditorRole(u)

	case "agora":
		AgoraRole(u)

	case "streamer":
		StreamerRole(u)

	case "recruiter":
		RecruiterRole(u)

	case "citizen":
		CitizenRole(u)

	case "clear":
		ClearRoles(u)

	default:
		return
	}
}

// RemoveRole function
func (u *User) RemoveRole(role string) {

	switch role {

	case "owner":
		u.Owner = false

	case "admin":
		u.Admin = false

	case "smoderator":
		u.SModerator = false

	case "moderator":
		u.Moderator = false

	case "editor":
		u.Editor = false

	case "agora":
		u.Agora = false

	case "streamer":
		u.Streamer = false

	case "recruiter":
		u.Recruiter = false

	case "citizen":
		u.Citizen = false

	}
}

// CheckRole function
func (u *User) CheckRole(role string) bool {

	switch role {

	case "owner":
		return u.Owner

	case "admin":
		return u.Admin

	case "smoderator":
		return u.SModerator

	case "moderator":
		return u.Moderator

	case "editor":
		return u.Editor

	case "agora":
		return u.Agora

	case "streamer":
		return u.Streamer

	case "recruiter":
		return u.Recruiter

	case "citizen":
		return u.Citizen
	}

	return false
}
