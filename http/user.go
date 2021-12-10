package http

import (
	"html/template"
	"net/http"

	"github.com/leg100/otf/http/html/assets"
)

// User represents a Terraform Enterprise user.
type User struct {
	ID               string     `jsonapi:"primary,users"`
	AvatarURL        string     `jsonapi:"attr,avatar-url"`
	Email            string     `jsonapi:"attr,email"`
	IsServiceAccount bool       `jsonapi:"attr,is-service-account"`
	TwoFactor        *TwoFactor `jsonapi:"attr,two-factor"`
	UnconfirmedEmail string     `jsonapi:"attr,unconfirmed-email"`
	Username         string     `jsonapi:"attr,username"`
	V2Only           bool       `jsonapi:"attr,v2-only"`

	// Relations
	// AuthenticationTokens *AuthenticationTokens `jsonapi:"relation,authentication-tokens"`
}

type GetProfileTemplateOptions struct {
	assets.LayoutTemplateOptions
}

// TwoFactor represents the organization permissions.
type TwoFactor struct {
	Enabled  bool `jsonapi:"attr,enabled"`
	Verified bool `jsonapi:"attr,verified"`
}

func (s *Server) GetProfile(w http.ResponseWriter, r *http.Request) {
	opts := GetProfileTemplateOptions{
		LayoutTemplateOptions: s.NewLayoutTemplateOptions("Tokens", w, r),
	}

	if err := s.GetTemplate("tokens_list.tmpl").Execute(w, opts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) NewLayoutTemplateOptions(title string, w http.ResponseWriter, r *http.Request) assets.LayoutTemplateOptions {
	return assets.LayoutTemplateOptions{
		Title:       title,
		Stylesheets: s.Links(),
	}
}

func interfaceSliceToStringSlice(is []interface{}) (ss []string) {
	for _, i := range is {
		ss = append(ss, i.(string))
	}
	return ss
}

func strSliceToHTMLTemplateSlice(s []string) (ht []template.HTML) {
	for _, i := range s {
		ht = append(ht, template.HTML(i))
	}
	return ht
}
