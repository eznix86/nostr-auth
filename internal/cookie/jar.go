package cookie

import "net/http"

type Jar struct {
	Domain   string
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
}

func NewJar(domain string, secure bool) *Jar {
	return &Jar{
		Domain:   domain,
		Secure:   secure,
		HTTPOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func (j *Jar) Set(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, j.Cookie(name, value, maxAge))
}

func (j *Jar) Clear(w http.ResponseWriter, name string) {
	j.Set(w, name, "", -1)
}

func (j *Jar) Value(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func (j *Jar) Cookie(name, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   j.Domain,
		HttpOnly: j.HTTPOnly,
		SameSite: j.SameSite,
		Secure:   j.Secure,
		MaxAge:   maxAge,
	}
}
