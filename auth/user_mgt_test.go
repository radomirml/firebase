package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"firebase.google.com/go/internal"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

type testReq struct {
	Method string
	Path   string
	Header http.Header
	Body   []byte
	Query  map[string]string
}

func newTestReq(r *http.Request) (*testReq, error) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(r.RequestURI)
	if err != nil {
		return nil, err
	}

	query := make(map[string]string)
	for k, v := range u.Query() {
		query[k] = v[0]
	}
	return &testReq{
		Method: r.Method,
		Path:   u.Path,
		Header: r.Header,
		Body:   b,
		Query:  query,
	}, nil
}

type mockAuthServer struct {
	Resp   interface{}
	Header map[string]string
	Status int
	Req    *http.Request
	srv    *httptest.Server
	client *Client
}

func echoServer(resp interface{}) *mockAuthServer {
	s := mockAuthServer{Resp: resp}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		s.Req = r

		for k, v := range s.Header {
			w.Header().Set(k, v)
		}
		if s.Status != 0 {
			w.WriteHeader(s.Status)
		}
		b, _ := json.Marshal(s.Resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})
	s.srv = httptest.NewServer(handler)
	authClient, err := NewClient(context.Background(),
		&internal.AuthConfig{
			Opts: []option.ClientOption{option.WithHTTPClient(s.srv.Client())},
		})
	_ = err
	authClient.url = s.srv.URL
	s.client = authClient
	return &s
}
func (s *mockAuthServer) Client() *Client {
	return s.client
}

func TestExportPayload(t *testing.T) {
	uf := NewUserFields()
	_ = uf
}

/*
users := []map[string]interface{}{
{
        "localId" : "testuser0",
        "email" : "testuser@example.com",
        "phoneNumber" : "+1234567890",
        "emailVerified" : true,
        "displayName" : "Test User",
        "providerUserInfo" : [ {
            "providerId" : "password",
            "displayName" : "Test User",
            "photoUrl" : "http://www.example.com/testuser/photo.png",
            "federatedId" : "testuser@example.com",
            "email" : "testuser@example.com",
            "rawId" : "testuser@example.com"
        }, {
            "providerId" : "phone",
            "phoneNumber" : "+1234567890",
            "rawId" : "+1234567890"
        } ],
        "photoUrl" : "http://www.example.com/testuser/photo.png",
        "passwordHash" : "passwordHash",
        "salt": "passwordSalt",
        "passwordUpdatedAt" : 1.494364393E+12,
        "validSince" : "1494364393",
        "disabled" : false,
        "createdAt" : "1234567890",
        "customAttributes" : "{\"admin\": true, \"package\": \"gold\"}"
    }, {
        "localId" : "testuser1",
        "email" : "testuser@example.com",
        "phoneNumber" : "+1234567890",
        "emailVerified" : true,
        "displayName" : "Test User",
        "providerUserInfo" : [ {
            "providerId" : "password",
            "displayName" : "Test User",
            "photoUrl" : "http://www.example.com/testuser/photo.png",
            "federatedId" : "testuser@example.com",
            "email" : "testuser@example.com",
            "rawId" : "testuser@example.com"
        }, {
            "providerId" : "phone",
            "phoneNumber" : "+1234567890",
            "rawId" : "+1234567890"
        } ],
        "photoUrl" : "http://www.example.com/testuser/photo.png",
        "passwordHash" : "passwordHash",
        "salt": "passwordSalt",
        "passwordUpdatedAt" : 1.494364393E+12,
        "validSince" : "1494364393",
        "disabled" : false,
        "createdAt" : "1234567890",
        "customAttributes" : "{\"admin\": true, \"package\": \"gold\"}"
    } ]
*/
func TestGetUser(t *testing.T) {

	//	b, err := ioutil.ReadFile(internal.Resource("get_user_data.json"))
	//	if err != nil {
	//		log.Fatalln(err)
	//	}
	//	t.Fatalf("%#v \n\n%s\n", b, string(b))
	s := echoServer(map[string]interface{}{
		"kind": "identitytoolkit#GetAccountInfoResponse",
		"users": []map[string]interface{}{
			{
				"localId":       "ZY1rJK0...",
				"email":         "user@example.com",
				"emailVerified": true,
				"displayName":   "John Doe",
				"providerUserInfo": []map[string]interface{}{
					{
						"providerId":  "password",
						"displayName": "John Doe",
						"photoUrl":    "http://localhost:8080/img1234567890/photo.png",
						"email":       "user@example.com",
					},
				},
				"photoUrl":     "https://lh5.googleusercontent.com/.../photo.jpg",
				"passwordHash": "...",
				"disabled":     false,
				"lastLoginAt":  "1484628946000",
				"createdAt":    "1484124142000",
			},
		},
	})

	defer s.srv.Close()
	user, err := s.Client().GetUser(context.Background(), "ignored_id")
	if err != nil {
		t.Error(err)

	}
	tests := []struct {
		got  interface{}
		want interface{}
	}{
		{user.UID, "ZY1rJK0..."},
		{user.UserMetadata, &UserMetadata{CreationTimestamp: 1484124142000, LastLogInTimestamp: 1484628946000}},
		{user.Email, "user@example.com"},
		{user.EmailVerified, true},
		{user.PhotoURL, "https://lh5.googleusercontent.com/.../photo.jpg"},
		{user.Disabled, false},
		/*	{user.ProviderUserInfo, []map[string]interface{}{
			{
				"providerId":  "password",
				"displayName": "John Doe",
				"photoUrl":    "http://localhost:8080/img1234567890/photo.png",
				"email":       "user@example.com",
			},
		}},*/
	}
	for _, test := range tests {
		if !reflect.DeepEqual(test.want, test.got) {
			t.Errorf("got %#v wanted %#v", test.got, test.want)
		}
	}
	//	t.Errorf("%#v %#v", user.UserMetadata, user.CustomClaims)
}
