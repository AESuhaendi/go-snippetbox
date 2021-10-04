package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func Test_ping(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	ping(rr, r)

	rs := rr.Result()
	if rs.StatusCode != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, rs.StatusCode)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "OK" {
		t.Errorf("want body to equal %q", "OK")
	}
}

func Test_ping_E2E(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")

	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}

	if string(body) != "OK" {
		t.Errorf("want body to equal %q", "OK")
	}
}

func Test_showSnippet(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody []byte
	}{
		{"Valid ID", "/snippet/1", http.StatusOK, []byte("An old silent pond...")},
		{"Non-existent ID", "/snippet/2", http.StatusNotFound, nil},
		{"Negative ID", "/snippet/-1", http.StatusNotFound, nil},
		{"Decimal ID", "/snippet/1.23", http.StatusNotFound, nil},
		{"String ID", "/snippet/foo", http.StatusNotFound, nil},
		{"Empty ID", "/snippet/", http.StatusNotFound, nil},
		{"Trailing slash", "/snippet/1/", http.StatusNotFound, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q", tt.wantBody)
			}
		})
	}
}

func Test_signupUser(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, _, body := ts.get(t, "/user/signup")
	csrfToken := extractCSRFToken(t, body)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     []byte
	}{
		{"Valid submission", "admin", "admin@localhost", "passwordadmin", csrfToken, http.StatusSeeOther, nil},
		{"Empty name", "", "admin@localhost", "passwordadmin", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Empty email", "admin", "", "passwordadmin", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Empty password", "admin", "admin@localhost", "", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Invalid email (incomplete domain)", "admin", "admin@example.", "passwordadmin", csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Invalid email (missing @)", "admin", "adminexample.com", "passwordadmin", csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Invalid email (missing local part)", "admin", "@example.com", "passwordadmin", csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Short password", "admin", "admin@localhost", "pass", csrfToken, http.StatusOK, []byte("This field is too short (minimum is 10 characters)")},
		{"Duplicate email", "dupe", "dupe@example.com", "passwordadmin", csrfToken, http.StatusOK, []byte("Address is already in use")},
		{"Invalid CSRF Token", "", "", "", "wrongToken", http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contains %q", body, tt.wantBody)
			}
		})
	}
}

func Test_createSnippetForm(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name               string
		isAuth             bool
		wantCode           int
		wantBody           []byte
		wantHeaderLocation string
	}{
		{
			name:               "Unauthenticated",
			isAuth:             false,
			wantCode:           http.StatusSeeOther,
			wantBody:           []byte(""),
			wantHeaderLocation: "/user/login",
		},
		{
			name:               "Authenticated",
			isAuth:             true,
			wantCode:           http.StatusOK,
			wantBody:           []byte(`<form action="/snippet/create" method="POST">`),
			wantHeaderLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isAuth {
				_, _, body := ts.get(t, "/user/login")
				csrfToken := extractCSRFToken(t, body)

				form := url.Values{}
				form.Add("email", "alice@example.com")
				form.Add("password", "")
				form.Add("csrf_token", csrfToken)
				ts.postForm(t, "/user/login", form)
			}

			code, header, body := ts.get(t, "/snippet/create")

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			loc := header.Get("Location")
			if loc != tt.wantHeaderLocation {
				t.Errorf("want %q; got %q", tt.wantHeaderLocation, loc)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q", tt.wantBody)
			}
		})
	}

}
