package authboss

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func loadClientStateP(ab *Authboss, w http.ResponseWriter, r *http.Request) *http.Request {
	r, err := ab.LoadClientState(w, r)
	if err != nil {
		panic(err)
	}

	return r
}

func testSetupContext() (*Authboss, *http.Request) {
	ab := New()
	ab.Storage.SessionState = newMockClientStateRW(SessionKey, "george-pid")
	ab.Storage.Server = &mockServerStorer{
		Users: map[string]*mockUser{
			"george-pid": {Email: "george-pid", Password: "unreadable"},
		},
	}
	r := httptest.NewRequest("GET", "/", nil)
	w := ab.NewResponse(httptest.NewRecorder())
	r = loadClientStateP(ab, w, r)

	return ab, r
}

func testSetupContextCached() (*Authboss, *mockUser, *http.Request) {
	ab := New()
	wantUser := &mockUser{Email: "george-pid", Password: "unreadable"}
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), CTXKeyPID, "george-pid")
	ctx = context.WithValue(ctx, CTXKeyUser, wantUser)
	req = req.WithContext(ctx)

	return ab, wantUser, req
}

func testSetupContextPanic() *Authboss {
	ab := New()
	ab.Storage.SessionState = newMockClientStateRW(SessionKey, "george-pid")
	ab.Storage.Server = &mockServerStorer{}

	return ab
}

func TestCurrentUserID(t *testing.T) {
	t.Parallel()

	ab, r := testSetupContext()

	id, err := ab.CurrentUserID(r)
	if err != nil {
		t.Error(err)
	}

	if id != "george-pid" {
		t.Error("got:", id)
	}
}

func TestCurrentUserIDContext(t *testing.T) {
	t.Parallel()

	ab, r := testSetupContext()

	id, err := ab.CurrentUserID(r)
	if err != nil {
		t.Error(err)
	}

	if id != "george-pid" {
		t.Error("got:", id)
	}
}

func TestCurrentUserIDP(t *testing.T) {
	t.Parallel()

	ab := testSetupContextPanic()
	// Overwrite the setup functions state storer
	ab.Storage.SessionState = newMockClientStateRW()

	defer func() {
		if recover().(error) != ErrUserNotFound {
			t.Failed()
		}
	}()

	_ = ab.CurrentUserIDP(httptest.NewRequest("GET", "/", nil))
}

func TestCurrentUser(t *testing.T) {
	t.Parallel()

	ab, r := testSetupContext()

	user, err := ab.CurrentUser(r)
	if err != nil {
		t.Error(err)
	}

	if got := user.GetPID(); got != "george-pid" {
		t.Error("got:", got)
	}
}

func TestCurrentUserContext(t *testing.T) {
	t.Parallel()

	ab, _, r := testSetupContextCached()

	user, err := ab.CurrentUser(r)
	if err != nil {
		t.Error(err)
	}

	if got := user.GetPID(); got != "george-pid" {
		t.Error("got:", got)
	}
}

func TestCurrentUserP(t *testing.T) {
	t.Parallel()

	ab := testSetupContextPanic()

	defer func() {
		if recover().(error) != ErrUserNotFound {
			t.Failed()
		}
	}()

	_ = ab.CurrentUserP(httptest.NewRequest("GET", "/", nil))
}

func TestLoadCurrentUserID(t *testing.T) {
	t.Parallel()

	ab, r := testSetupContext()

	id, err := ab.LoadCurrentUserID(&r)
	if err != nil {
		t.Error(err)
	}

	if id != "george-pid" {
		t.Error("got:", id)
	}

	if r.Context().Value(CTXKeyPID).(string) != "george-pid" {
		t.Error("context was not updated in local request")
	}
}

func TestLoadCurrentUserIDContext(t *testing.T) {
	t.Parallel()

	ab, _, r := testSetupContextCached()

	pid, err := ab.LoadCurrentUserID(&r)
	if err != nil {
		t.Error(err)
	}

	if pid != "george-pid" {
		t.Error("got:", pid)
	}
}

func TestLoadCurrentUserIDP(t *testing.T) {
	t.Parallel()

	ab := testSetupContextPanic()

	defer func() {
		if recover().(error) != ErrUserNotFound {
			t.Failed()
		}
	}()

	r := httptest.NewRequest("GET", "/", nil)
	_ = ab.LoadCurrentUserIDP(&r)
}

func TestLoadCurrentUser(t *testing.T) {
	t.Parallel()

	ab, r := testSetupContext()

	user, err := ab.LoadCurrentUser(&r)
	if err != nil {
		t.Error(err)
	}

	if got := user.GetPID(); got != "george-pid" {
		t.Error("got:", got)
	}

	want := user.(*mockUser)
	got := r.Context().Value(CTXKeyUser).(*mockUser)
	if got != want {
		t.Errorf("users mismatched:\nwant: %#v\ngot: %#v", want, got)
	}
}

func TestLoadCurrentUserContext(t *testing.T) {
	t.Parallel()

	ab, wantUser, r := testSetupContextCached()

	user, err := ab.LoadCurrentUser(&r)
	if err != nil {
		t.Error(err)
	}

	got := user.(*mockUser)
	if got != wantUser {
		t.Errorf("users mismatched:\nwant: %#v\ngot: %#v", wantUser, got)
	}
}

func TestLoadCurrentUserP(t *testing.T) {
	t.Parallel()

	ab := testSetupContextPanic()

	defer func() {
		if recover().(error) != ErrUserNotFound {
			t.Failed()
		}
	}()

	r := httptest.NewRequest("GET", "/", nil)
	_ = ab.LoadCurrentUserP(&r)
}

func TestCTXKeyString(t *testing.T) {
	t.Parallel()

	if got := CTXKeyPID.String(); got != "authboss ctx key pid" {
		t.Error(got)
	}
}
