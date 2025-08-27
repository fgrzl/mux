package mux

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldBindParseQueryParamAsInteger(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?limit=10", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	// Attach RouteOptions that declare the type of "limit" as integer
	rb := Route("GET", "/").WithQueryParam("limit", 10)
	ctx.options = rb.Options

	var model struct {
		Limit int `json:"limit"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10, model.Limit)
}

func TestShouldBindParseHeaderParamAsInteger(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Count", "42")
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	// Declare header param X-Count with integer example
	rb := Route("GET", "/").WithHeaderParam("X-Count", 42, false)
	ctx.options = rb.Options

	var model struct {
		XCount int `json:"X-Count"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 42, model.XCount)
}

func TestShouldBindParsePathParamAsUUID(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/items/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)

	// Set the params map (what the router would populate)
	ctx.params = RouteParams{"id": "550e8400-e29b-41d4-a716-446655440000"}

	// Declare path parameter "id" as a UUID via example
	rb := Route("GET", "/items/{id}").WithPathParam("id", uuid.Nil)
	ctx.options = rb.Options

	var model struct {
		ID uuid.UUID `json:"id"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), model.ID)
}

// Numeric type tests
func TestShouldBindParseQueryInt(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?i=1", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("i", 1)
	ctx.options = rb.Options

	var model struct {
		I int `json:"i"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, model.I)
}

func TestShouldBindParseQueryInt16(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?i16=16", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("i16", int16(16))
	ctx.options = rb.Options

	var model struct {
		I16 int16 `json:"i16"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int16(16), model.I16)
}

func TestShouldBindParseQueryInt32(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?i32=32000", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("i32", int32(32000))
	ctx.options = rb.Options

	var model struct {
		I32 int32 `json:"i32"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int32(32000), model.I32)
}

func TestShouldBindParseQueryInt64(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?i64=64000000000", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("i64", int64(64000000000))
	ctx.options = rb.Options

	var model struct {
		I64 int64 `json:"i64"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(64000000000), model.I64)
}

func TestShouldBindParseQueryUint(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?u=7", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("u", uint(7))
	ctx.options = rb.Options

	var model struct {
		U uint `json:"u"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, uint(7), model.U)
}

func TestShouldBindParseQueryFloat32(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?f32=1.5", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("f32", float32(1.5))
	ctx.options = rb.Options

	var model struct {
		F32 float32 `json:"f32"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.InDelta(t, float32(1.5), model.F32, 0.0001)
}

func TestShouldBindParseQueryFloat64(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?f64=2.5", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("f64", float64(2.5))
	ctx.options = rb.Options

	var model struct {
		F64 float64 `json:"f64"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.InDelta(t, float64(2.5), model.F64, 0.0000001)
}

func TestShouldBindParseHeaderInt64(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Int64", "9223372036854775807")
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithHeaderParam("X-Int64", int64(0), false)
	ctx.options = rb.Options

	var model struct {
		HInt64 int64 `json:"X-Int64"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(9223372036854775807), model.HInt64)
}

func TestShouldBindParseHeaderUint(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Uint", "123")
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithHeaderParam("X-Uint", uint(0), false)
	ctx.options = rb.Options

	var model struct {
		HUint uint `json:"X-Uint"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, uint(123), model.HUint)
}

func TestShouldBindParsePathInt64(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/items/12345", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	ctx.params = RouteParams{"pid": "12345"}
	rb := Route("GET", "/items/{pid}").WithPathParam("pid", int64(12345))
	ctx.options = rb.Options

	var model struct {
		PID int64 `json:"pid"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(12345), model.PID)
}

func TestShouldBindParseQueryIntSlice(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?a=1&a=2&a=3", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("a", []int{1})
	ctx.options = rb.Options

	var model struct {
		A []int `json:"a"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, model.A)
}

func TestShouldBindParseQueryStringSlice(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?s=hello&s=world", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("s", []string{"x"})
	ctx.options = rb.Options

	var model struct {
		S []string `json:"s"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"hello", "world"}, model.S)
}

func TestShouldReturnErrorForInvalidIntQueryParam(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?i=notanint", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("i", 1)
	ctx.options = rb.Options

	var model struct {
		I int `json:"i"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.Error(t, err)
}

func TestShouldReturnErrorForIntegerOverflow(t *testing.T) {
	// Arrange: int32 overflow value
	req := httptest.NewRequest("GET", "/?i=99999999999999999999", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("i", int32(0))
	ctx.options = rb.Options

	var model struct {
		I int32 `json:"i"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert
	require.Error(t, err)
}

func TestShouldReturnErrorForHeaderCommaListWhenExpectingSlice(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-List", "a,b,c")
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithHeaderParam("X-List", []string{"x"}, false)
	ctx.options = rb.Options

	var model struct {
		List []string `json:"X-List"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert: CSV splitting is supported for headers when declared as a slice
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, model.List)
}

func TestShouldSplitCommaSeparatedHeaderIntoSlice(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-List", "a,b,c")
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithHeaderParam("X-List", []string{"x"}, false)
	ctx.options = rb.Options

	var model struct {
		List []string `json:"X-List"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert: we expect CSV splitting support (this test will fail until implemented)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, model.List)
}

func TestShouldSplitCommaSeparatedQueryIntoSlice(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?tags=a,b,c", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("tags", []string{"x"})
	ctx.options = rb.Options

	var model struct {
		Tags []string `json:"tags"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert: we expect CSV splitting support (this test will fail until implemented)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, model.Tags)
}

func TestShouldBindDotNotationToNestedStruct(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?user.name=alice&user.age=30", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	// declare the root parameter as an object so the generator/schema knows property types
	rb := Route("GET", "/").WithQueryParam("user", struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{})
	ctx.options = rb.Options

	var model struct {
		User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"user"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert: expected behavior (test will fail until dot/bracket parsing is implemented)
	require.NoError(t, err)
	assert.Equal(t, "alice", model.User.Name)
	assert.Equal(t, 30, model.User.Age)
}

func TestShouldBindBracketNotationToNestedStruct(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/?user[name]=bob&user[age]=25", nil)
	rec := httptest.NewRecorder()
	ctx := NewRouteContext(rec, req)
	rb := Route("GET", "/").WithQueryParam("user", struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{})
	ctx.options = rb.Options

	var model struct {
		User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"user"`
	}

	// Act
	err := ctx.Bind(&model)

	// Assert: expected behavior (test will fail until deep-object parsing is implemented)
	require.NoError(t, err)
	assert.Equal(t, "bob", model.User.Name)
	assert.Equal(t, 25, model.User.Age)
}
