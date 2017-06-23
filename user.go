package main

type User struct {

	ID string // this field will be indexed

	Owner 		bool
	Admin 		bool
	SModerator	bool
	JModerator	bool
	Editor		bool
	Agora		bool
	Streamer	bool
	Recruiter	bool

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

	}
}


