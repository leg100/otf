// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
)

func Users(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/users", organization))
}

func CreateUser(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/users/create", organization))
}

func NewUser(organization string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/users/new", organization))
}

func User(user string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/users/%s", user))
}

func EditUser(user string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/users/%s/edit", user))
}

func UpdateUser(user string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/users/%s/update", user))
}

func DeleteUser(user string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/users/%s/delete", user))
}
