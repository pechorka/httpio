package httpio_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pechorka/gostdlib/pkg/testing/require"
	"github.com/pechorka/httpio"
)

func TestUnmarshal(t *testing.T) {
	t.Run("query params only", func(t *testing.T) {
		type fullName struct {
			First  string  `query:"first"`
			Last   string  `query:"last"`
			Middle *string `query:"middle"`
		}
		type input struct {
			Name   fullName `query:"name"`
			Age    int      `query:"age"`
			Banned bool     `query:"banned"`
			Income uint     `query:"income"`
		}

		r := httptest.NewRequest("GET", "/?name.first=John&name.last=Doe&age=30&banned=true&income=100000&name.middle=Middle", nil)

		unmarshaler, err := httpio.NewUnmarshaler[input]()
		require.NoError(t, err)

		var v input
		err = unmarshaler.Unmarshal(r, &v)
		require.NoError(t, err)

		require.Equal(t, "John", v.Name.First)
		require.Equal(t, "Doe", v.Name.Last)
		require.Equal(t, "Middle", *v.Name.Middle)
		require.Equal(t, 30, v.Age)
		require.Equal(t, true, v.Banned)
		require.Equal(t, uint(100000), v.Income)
	})

	t.Run("json and query params", func(t *testing.T) {
		type fullName struct {
			First string `query:"first"`
			Last  string `query:"last"`
		}
		type input struct {
			Name      fullName `query:"name"`
			Age       int      `query:"age"`
			Banned    bool     `query:"banned"`
			Income    uint     `query:"income"`
			AppConfig struct {
				Host string `json:"host"`
				Port int    `json:"port"`
			} `json:"app_config"`
		}

		body := `{"app_config":{"host":"localhost","port":8080}}`

		r := httptest.NewRequest("POST", "/?name.first=John&name.last=Doe&age=30&banned=true&income=100000", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")

		unmarshaler, err := httpio.NewUnmarshaler[input]()
		require.NoError(t, err)

		var v input
		err = unmarshaler.Unmarshal(r, &v)
		require.NoError(t, err)

		require.Equal(t, "John", v.Name.First)
		require.Equal(t, "Doe", v.Name.Last)
		require.Equal(t, 30, v.Age)
		require.Equal(t, true, v.Banned)
		require.Equal(t, uint(100000), v.Income)
		require.Equal(t, "localhost", v.AppConfig.Host)
		require.Equal(t, 8080, v.AppConfig.Port)
	})

	t.Run("path params", func(t *testing.T) {
		type input struct {
			UserID string `path:"user_id"`
			OrgID  string `path:"org_id"`
		}

		r := httptest.NewRequest("GET", "/users/123/orgs/456", nil)
		r.SetPathValue("user_id", "123")
		r.SetPathValue("org_id", "456")

		unmarshaler, err := httpio.NewUnmarshaler[input]()
		require.NoError(t, err)

		var v input
		err = unmarshaler.Unmarshal(r, &v)
		require.NoError(t, err)

		require.Equal(t, "123", v.UserID)
		require.Equal(t, "456", v.OrgID)
	})

	t.Run("header params", func(t *testing.T) {
		type input struct {
			Authorization string `header:"Authorization"`
			UserAgent     string `header:"User-Agent"`
		}

		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", "Bearer token123")
		r.Header.Set("User-Agent", "test-client/1.0")

		unmarshaler, err := httpio.NewUnmarshaler[input]()
		require.NoError(t, err)

		var v input
		err = unmarshaler.Unmarshal(r, &v)
		require.NoError(t, err)

		require.Equal(t, "Bearer token123", v.Authorization)
		require.Equal(t, "test-client/1.0", v.UserAgent)
	})

	t.Run("cookie params", func(t *testing.T) {
		type input struct {
			SessionID string `cookie:"session_id"`
			Theme     string `cookie:"theme"`
		}

		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
		r.AddCookie(&http.Cookie{Name: "theme", Value: "dark"})

		unmarshaler, err := httpio.NewUnmarshaler[input]()
		require.NoError(t, err)

		var v input
		err = unmarshaler.Unmarshal(r, &v)
		require.NoError(t, err)

		require.Equal(t, "abc123", v.SessionID)
		require.Equal(t, "dark", v.Theme)
	})
}

func BenchmarkUnmarshal(b *testing.B) {
	type fullName struct {
		First string `query:"first"`
		Last  string `query:"last"`
	}
	type input struct {
		Name    fullName `query:"name"`
		Age     int      `query:"age"`
		Banned  bool     `query:"banned"`
		Income  uint     `query:"income"`
		Field1  fullName `query:"field1"`
		Field2  fullName `query:"field2"`
		Field3  fullName `query:"field3"`
		Field4  fullName `query:"field4"`
		Field5  fullName `query:"field5"`
		Field6  fullName `query:"field6"`
		Field7  fullName `query:"field7"`
		Field8  fullName `query:"field8"`
		Field9  fullName `query:"field9"`
		Field10 fullName `query:"field10"`
		Field11 struct {
			Field1  fullName `query:"field1"`
			Field2  fullName `query:"field2"`
			Field3  fullName `query:"field3"`
			Field4  fullName `query:"field4"`
			Field5  fullName `query:"field5"`
			Field6  fullName `query:"field6"`
			Field7  fullName `query:"field7"`
			Field8  fullName `query:"field8"`
			Field9  fullName `query:"field9"`
			Field10 fullName `query:"field10"`
			Field11 struct {
				Field1  fullName `query:"field1"`
				Field2  fullName `query:"field2"`
				Field3  fullName `query:"field3"`
				Field4  fullName `query:"field4"`
				Field5  fullName `query:"field5"`
				Field6  fullName `query:"field6"`
				Field7  fullName `query:"field7"`
				Field8  fullName `query:"field8"`
				Field9  fullName `query:"field9"`
				Field10 fullName `query:"field10"`
				Field11 struct {
					Field1  fullName `query:"field1"`
					Field2  fullName `query:"field2"`
					Field3  fullName `query:"field3"`
					Field4  fullName `query:"field4"`
					Field5  fullName `query:"field5"`
					Field6  fullName `query:"field6"`
					Field7  fullName `query:"field7"`
					Field8  fullName `query:"field8"`
					Field9  fullName `query:"field9"`
					Field10 fullName `query:"field10"`
					Field11 struct {
						Field1  fullName `query:"field1"`
						Field2  fullName `query:"field2"`
						Field3  fullName `query:"field3"`
						Field4  fullName `query:"field4"`
						Field5  fullName `query:"field5"`
						Field6  fullName `query:"field6"`
						Field7  fullName `query:"field7"`
						Field8  fullName `query:"field8"`
						Field9  fullName `query:"field9"`
						Field10 fullName `query:"field10"`
						Field11 struct {
							Field1  fullName `query:"field1"`
							Field2  fullName `query:"field2"`
							Field3  fullName `query:"field3"`
							Field4  fullName `query:"field4"`
							Field5  fullName `query:"field5"`
							Field6  fullName `query:"field6"`
							Field7  fullName `query:"field7"`
							Field8  fullName `query:"field8"`
							Field9  fullName `query:"field9"`
							Field10 fullName `query:"field10"`
							Field11 struct {
								Field1 fullName `query:"field1"`
							} `query:"field11"`
						} `query:"field11"`
					} `query:"field11"`
				} `query:"field11"`
			} `query:"field11"`
		} `query:"field11"`
	}

	r := httptest.NewRequest("GET", "/?name.first=John&name.last=Doe&age=30&banned=true&income=100000", nil)

	unmarshaler, err := httpio.NewUnmarshaler[input]()
	require.NoError(b, err)

	b.ReportAllocs()

	for b.Loop() {
		var v input
		err := unmarshaler.Unmarshal(r, &v)
		if err != nil {
			b.Fatal(err)
		}
	}
}
