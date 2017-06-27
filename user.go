package main

type User struct {

	ID string `storm:"id"`// primary key

	Owner      bool `storm:"index"`
	Admin      bool `storm:"index"`
	SModerator bool `storm:"index"`
	Moderator  bool `storm:"index"`
	Editor     bool `storm:"index"`
	Agora      bool `storm:"index"`
	Streamer   bool `storm:"index"`
	Recruiter  bool `storm:"index"`
	Citizen    bool `storm:"index"`

}

func (u *User) Init() {
	ClearRoles(u)
	CitizenRole(u)
}

func (u *User) SetRole(role string) {

	switch role {

	case "owner" :
		OwnerRole(u)

	case "admin" :
		AdminRole(u)

	case "smoderator" :
		SModeratorRole(u)

	case "moderator" :
		ModeratorRole(u)

	case "editor" :
		EditorRole(u)

	case "agora" :
		AgoraRole(u)

	case "streamer" :
		StreamerRole(u)

	case "recruiter" :
		RecruiterRole(u)

	case "citizen" :
		CitizenRole(u)

	case "clear" :
		ClearRoles(u)
	}
}

func (u *User) CheckRole(role string) bool {

	switch role {

	case "owner" :
		return u.Owner

	case "admin" :
		return u.Admin

	case "smoderator" :
		return u.SModerator

	case "moderator" :
		return u.Moderator

	case "editor" :
		return u.Editor

	case "agora" :
		return u.Agora

	case "streamer" :
		return u.Streamer

	case "recruiter" :
		return u.Recruiter

	case "citizen" :
		return u.Citizen
	}

	return false
}
