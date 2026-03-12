package server

import (
	"net/http"

	gonertia "github.com/romsar/gonertia/v2"
)

type LogoutPage struct{ H *Context }

func (l *LogoutPage) Index(w http.ResponseWriter, r *http.Request) {
	pubkey := AuthenticatedPubkeyFromContext(r)
	if pubkey == "" {
		l.H.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := l.H.Render(w, r, "Logout", gonertia.Props{
		"title":               l.H.Config.AppName,
		"authenticatedPubkey": pubkey,
		"profile":             l.H.ProfileProp(w, r, "logout.index"),
	})
	if err != nil {
		l.H.Fail(w, err, "failed to render logout page")
	}
}
