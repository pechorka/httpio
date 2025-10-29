package httpio_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		assertNoError(t, err)

		var v input
		err = unmarshaler.Unmarshal(r, &v)
		assertNoError(t, err)

		assertEqual(t, "John", v.Name.First)
		assertEqual(t, "Doe", v.Name.Last)
		assertEqual(t, "Middle", *v.Name.Middle)
		assertEqual(t, 30, v.Age)
		assertEqual(t, true, v.Banned)
		assertEqual(t, uint(100000), v.Income)
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
		assertNoError(t, err)

		var v input
		err = unmarshaler.Unmarshal(r, &v)
		assertNoError(t, err)

		assertEqual(t, "John", v.Name.First)
		assertEqual(t, "Doe", v.Name.Last)
		assertEqual(t, 30, v.Age)
		assertEqual(t, true, v.Banned)
		assertEqual(t, uint(100000), v.Income)
		assertEqual(t, "localhost", v.AppConfig.Host)
		assertEqual(t, 8080, v.AppConfig.Port)
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
		assertNoError(t, err)

		var v input
		err = unmarshaler.Unmarshal(r, &v)
		assertNoError(t, err)

		assertEqual(t, "123", v.UserID)
		assertEqual(t, "456", v.OrgID)
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
		assertNoError(t, err)

		var v input
		err = unmarshaler.Unmarshal(r, &v)
		assertNoError(t, err)

		assertEqual(t, "Bearer token123", v.Authorization)
		assertEqual(t, "test-client/1.0", v.UserAgent)
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
		assertNoError(t, err)

		var v input
		err = unmarshaler.Unmarshal(r, &v)
		assertNoError(t, err)

		assertEqual(t, "abc123", v.SessionID)
		assertEqual(t, "dark", v.Theme)
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
	assertNoError(b, err)

	b.ReportAllocs()

	for b.Loop() {
		var v input
		err := unmarshaler.Unmarshal(r, &v)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func assertEqual[T comparable](tb testing.TB, expected, got T) {
	tb.Helper()
	if expected != got {
		tb.Fatalf("expected %v, got %v", expected, got)
	}
}

func assertNoError(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}

func assertError(tb testing.TB, err error) {
	tb.Helper()
	if err == nil {
		tb.Fatalf("expected an error, got nil")
	}
}
