// Code generated by "go generate"; DO NOT EDIT.

package paths

import "fmt"

func Users(organization fmt.Stringer) string {
	return fmt.Sprintf("/app/organizations/%s/users", organization)
}

func CreateUser(organization fmt.Stringer) string {
	return fmt.Sprintf("/app/organizations/%s/users/create", organization)
}

func NewUser(organization fmt.Stringer) string {
	return fmt.Sprintf("/app/organizations/%s/users/new", organization)
}

func User(user fmt.Stringer) string {
	return fmt.Sprintf("/app/users/%s", user)
}

func EditUser(user fmt.Stringer) string {
	return fmt.Sprintf("/app/users/%s/edit", user)
}

func UpdateUser(user fmt.Stringer) string {
	return fmt.Sprintf("/app/users/%s/update", user)
}

func DeleteUser(user fmt.Stringer) string {
	return fmt.Sprintf("/app/users/%s/delete", user)
}
