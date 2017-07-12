package main

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

func AdminRole(u *User) {
	u.Admin = true
	u.SModerator = true
	u.Moderator = true
	u.Editor = true
	u.Agora = true
	u.Streamer = true
	u.Recruiter = true
}

func SModeratorRole(u *User) {
	u.SModerator = true
	u.Moderator = true
	u.Editor = true
	u.Agora = true
	u.Streamer = true
	u.Recruiter = true
}

func ModeratorRole(u *User) {
	u.Moderator = true
	u.Editor = true
	u.Agora = true
	u.Streamer = true
	u.Recruiter = true
}

func EditorRole(u *User) {
	u.Editor = true
}

func AgoraRole(u *User) {
	u.Agora = true
}

func StreamerRole(u *User) {
	u.Streamer = true
}

func RecruiterRole(u *User) {
	u.Recruiter = true
}

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

func CitizenRole(u *User) {
	u.Citizen = true
}
