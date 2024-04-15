package httpio_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ggicci/httpin"
	"github.com/pechorka/httpio"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	t.Run("query params only", func(t *testing.T) {
		type fullName struct {
			First string `in:"query=first"`
			Last  string `in:"query=last"`
		}
		type input struct {
			Name   fullName `in:"query=name"`
			Age    int      `in:"query=age"`
			Banned bool     `in:"query=banned"`
			Income uint     `in:"query=income"`
		}

		r := httptest.NewRequest("GET", "/?name.first=John&name.last=Doe&age=30&banned=true&income=100000", nil)

		var v input
		err := httpio.Unmarshal(r, &v)
		require.NoError(t, err)

		require.Equal(t, "John", v.Name.First)
		require.Equal(t, "Doe", v.Name.Last)
		require.Equal(t, 30, v.Age)
		require.Equal(t, true, v.Banned)
		require.Equal(t, uint(100000), v.Income)
	})

	t.Run("json and query params", func(t *testing.T) {
		type fullName struct {
			First string `in:"query=first"`
			Last  string `in:"query=last"`
		}
		type input struct {
			Name      fullName `in:"query=name"`
			Age       int      `in:"query=age"`
			Banned    bool     `in:"query=banned"`
			Income    uint     `in:"query=income"`
			AppConfig struct {
				Host string `json:"host"`
				Port int    `json:"port"`
			} `json:"app_config"`
		}

		body := `{"app_config":{"host":"localhost","port":8080}}`

		r := httptest.NewRequest("POST", "/?name.first=John&name.last=Doe&age=30&banned=true&income=100000", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")

		var v input
		err := httpio.Unmarshal(r, &v)
		require.NoError(t, err)

		require.Equal(t, "John", v.Name.First)
		require.Equal(t, "Doe", v.Name.Last)
		require.Equal(t, 30, v.Age)
		require.Equal(t, true, v.Banned)
		require.Equal(t, uint(100000), v.Income)
		require.Equal(t, "localhost", v.AppConfig.Host)
		require.Equal(t, 8080, v.AppConfig.Port)
	})
}

func BenchmarkUnmarshal(b *testing.B) {
	type fullName struct {
		First string `in:"query=first"`
		Last  string `in:"query=last"`
	}
	type input struct {
		Name    fullName `in:"query=name"`
		Age     int      `in:"query=age"`
		Banned  bool     `in:"query=banned"`
		Income  uint     `in:"query=income"`
		Field1  fullName `in:"query=field1"`
		Field2  fullName `in:"query=field2"`
		Field3  fullName `in:"query=field3"`
		Field4  fullName `in:"query=field4"`
		Field5  fullName `in:"query=field5"`
		Field6  fullName `in:"query=field6"`
		Field7  fullName `in:"query=field7"`
		Field8  fullName `in:"query=field8"`
		Field9  fullName `in:"query=field9"`
		Field10 fullName `in:"query=field10"`
		Field11 struct {
			Field1  fullName `in:"query=field1"`
			Field2  fullName `in:"query=field2"`
			Field3  fullName `in:"query=field3"`
			Field4  fullName `in:"query=field4"`
			Field5  fullName `in:"query=field5"`
			Field6  fullName `in:"query=field6"`
			Field7  fullName `in:"query=field7"`
			Field8  fullName `in:"query=field8"`
			Field9  fullName `in:"query=field9"`
			Field10 fullName `in:"query=field10"`
			Field11 struct {
				Field1  fullName `in:"query=field1"`
				Field2  fullName `in:"query=field2"`
				Field3  fullName `in:"query=field3"`
				Field4  fullName `in:"query=field4"`
				Field5  fullName `in:"query=field5"`
				Field6  fullName `in:"query=field6"`
				Field7  fullName `in:"query=field7"`
				Field8  fullName `in:"query=field8"`
				Field9  fullName `in:"query=field9"`
				Field10 fullName `in:"query=field10"`
				Field11 struct {
					Field1  fullName `in:"query=field1"`
					Field2  fullName `in:"query=field2"`
					Field3  fullName `in:"query=field3"`
					Field4  fullName `in:"query=field4"`
					Field5  fullName `in:"query=field5"`
					Field6  fullName `in:"query=field6"`
					Field7  fullName `in:"query=field7"`
					Field8  fullName `in:"query=field8"`
					Field9  fullName `in:"query=field9"`
					Field10 fullName `in:"query=field10"`
					Field11 struct {
						Field1  fullName `in:"query=field1"`
						Field2  fullName `in:"query=field2"`
						Field3  fullName `in:"query=field3"`
						Field4  fullName `in:"query=field4"`
						Field5  fullName `in:"query=field5"`
						Field6  fullName `in:"query=field6"`
						Field7  fullName `in:"query=field7"`
						Field8  fullName `in:"query=field8"`
						Field9  fullName `in:"query=field9"`
						Field10 fullName `in:"query=field10"`
						Field11 struct {
							Field1  fullName `in:"query=field1"`
							Field2  fullName `in:"query=field2"`
							Field3  fullName `in:"query=field3"`
							Field4  fullName `in:"query=field4"`
							Field5  fullName `in:"query=field5"`
							Field6  fullName `in:"query=field6"`
							Field7  fullName `in:"query=field7"`
							Field8  fullName `in:"query=field8"`
							Field9  fullName `in:"query=field9"`
							Field10 fullName `in:"query=field10"`
							Field11 struct {
								Field1 fullName `in:"query=field1"`
							} `in:"query=field11"`
						} `in:"query=field11"`
					} `in:"query=field11"`
				} `in:"query=field11"`
			} `in:"query=field11"`
		} `in:"query=field11"`
	}

	r := httptest.NewRequest("GET", "/?name.first=John&name.last=Doe&age=30&banned=true&income=100000", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var v input
		err := httpio.Unmarshal(r, &v)
		require.NoError(b, err)
	}
}

func BenchmarkHttpin(b *testing.B) {
	type fullName struct {
		First string `in:"query=first"`
		Last  string `in:"query=last"`
	}
	type input struct {
		Name    fullName `in:"query=name"`
		Age     int      `in:"query=age"`
		Banned  bool     `in:"query=banned"`
		Income  uint     `in:"query=income"`
		Field1  fullName `in:"query=field1"`
		Field2  fullName `in:"query=field2"`
		Field3  fullName `in:"query=field3"`
		Field4  fullName `in:"query=field4"`
		Field5  fullName `in:"query=field5"`
		Field6  fullName `in:"query=field6"`
		Field7  fullName `in:"query=field7"`
		Field8  fullName `in:"query=field8"`
		Field9  fullName `in:"query=field9"`
		Field10 fullName `in:"query=field10"`
		Field11 struct {
			Field1  fullName `in:"query=field1"`
			Field2  fullName `in:"query=field2"`
			Field3  fullName `in:"query=field3"`
			Field4  fullName `in:"query=field4"`
			Field5  fullName `in:"query=field5"`
			Field6  fullName `in:"query=field6"`
			Field7  fullName `in:"query=field7"`
			Field8  fullName `in:"query=field8"`
			Field9  fullName `in:"query=field9"`
			Field10 fullName `in:"query=field10"`
			Field11 struct {
				Field1  fullName `in:"query=field1"`
				Field2  fullName `in:"query=field2"`
				Field3  fullName `in:"query=field3"`
				Field4  fullName `in:"query=field4"`
				Field5  fullName `in:"query=field5"`
				Field6  fullName `in:"query=field6"`
				Field7  fullName `in:"query=field7"`
				Field8  fullName `in:"query=field8"`
				Field9  fullName `in:"query=field9"`
				Field10 fullName `in:"query=field10"`
				Field11 struct {
					Field1  fullName `in:"query=field1"`
					Field2  fullName `in:"query=field2"`
					Field3  fullName `in:"query=field3"`
					Field4  fullName `in:"query=field4"`
					Field5  fullName `in:"query=field5"`
					Field6  fullName `in:"query=field6"`
					Field7  fullName `in:"query=field7"`
					Field8  fullName `in:"query=field8"`
					Field9  fullName `in:"query=field9"`
					Field10 fullName `in:"query=field10"`
					Field11 struct {
						Field1  fullName `in:"query=field1"`
						Field2  fullName `in:"query=field2"`
						Field3  fullName `in:"query=field3"`
						Field4  fullName `in:"query=field4"`
						Field5  fullName `in:"query=field5"`
						Field6  fullName `in:"query=field6"`
						Field7  fullName `in:"query=field7"`
						Field8  fullName `in:"query=field8"`
						Field9  fullName `in:"query=field9"`
						Field10 fullName `in:"query=field10"`
						Field11 struct {
							Field1  fullName `in:"query=field1"`
							Field2  fullName `in:"query=field2"`
							Field3  fullName `in:"query=field3"`
							Field4  fullName `in:"query=field4"`
							Field5  fullName `in:"query=field5"`
							Field6  fullName `in:"query=field6"`
							Field7  fullName `in:"query=field7"`
							Field8  fullName `in:"query=field8"`
							Field9  fullName `in:"query=field9"`
							Field10 fullName `in:"query=field10"`
							Field11 struct {
								Field1 fullName `in:"query=field1"`
							} `in:"query=field11"`
						} `in:"query=field11"`
					} `in:"query=field11"`
				} `in:"query=field11"`
			} `in:"query=field11"`
		} `in:"query=field11"`
	}

	r := httptest.NewRequest("GET", "/?name.first=John&name.last=Doe&age=30&banned=true&income=100000", nil)

	decoder, err := httpin.New(input{})
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := decoder.Decode(r)
		require.NoError(b, err)
	}
}
