package main

type User struct {

	ID string `storm:"id"`// primary key

	Owner 		bool `storm:"index"`
	Admin 		bool `storm:"index"`
	SModerator	bool `storm:"index"`
	JModerator	bool `storm:"index"`
	Editor		bool `storm:"index"`
	Agora		bool `storm:"index"`
	Streamer	bool `storm:"index"`
	Recruiter	bool `storm:"index"`
}

func (u *User) Init() {
	ClearRoles(u)
}

func (u *User) SetRole(role string) {

	switch role {

	case "owner" :
		OwnerRole(u)

	case "admin" :
		AdminRole(u)

	case "smoderator" :
		SModeratorRole(u)

	case "jmoderator" :
		JModeratorRole(u)

	case "editor" :
		EditorRole(u)

	case "agora" :
		AgoraRole(u)

	case "streamer" :
		StreamerRole(u)

	case "recruiter" :
		RecruiterRole(u)

	case "clear" :
		ClearRoles(u)
	}
}

