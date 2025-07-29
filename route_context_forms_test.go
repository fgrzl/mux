package mux

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldParseUrlEncodedFormValues(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("name", "test")
	formData.Set("age", "25")
	formData.Add("tags", "tag1")
	formData.Add("tags", "tag2")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act & Assert
	name, ok := c.FormValue("name")
	assert.True(t, ok)
	assert.Equal(t, "test", name)

	age, ok := c.FormValue("age")
	assert.True(t, ok)
	assert.Equal(t, "25", age)

	tags, ok := c.FormValues("tags")
	assert.True(t, ok)
	assert.Equal(t, []string{"tag1", "tag2"}, tags)
}

func TestShouldParseMultipartFormValues(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	writer.WriteField("name", "test")
	writer.WriteField("age", "25")
	writer.WriteField("tags", "tag1")
	writer.WriteField("tags", "tag2")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act & Assert
	name, ok := c.FormValue("name")
	assert.True(t, ok)
	assert.Equal(t, "test", name)

	age, ok := c.FormValue("age")
	assert.True(t, ok)
	assert.Equal(t, "25", age)

	tags, ok := c.FormValues("tags")
	assert.True(t, ok)
	assert.Equal(t, []string{"tag1", "tag2"}, tags)
}

func TestShouldReturnFalseWhenFormValueNotFound(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("name", "test")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act
	val, ok := c.FormValue("nonexistent")

	// Assert
	assert.False(t, ok)
	assert.Equal(t, "", val)
}

func TestShouldParseFormUUID(t *testing.T) {
	// Arrange
	testUUID := uuid.New()
	formData := url.Values{}
	formData.Set("id", testUUID.String())
	formData.Set("invalid", "not-a-uuid")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act & Assert
	id, ok := c.FormUUID("id")
	assert.True(t, ok)
	assert.Equal(t, testUUID, id)

	invalidID, ok := c.FormUUID("invalid")
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, invalidID)
}

func TestShouldParseFormUUIDs(t *testing.T) {
	// Arrange
	uuid1 := uuid.New()
	uuid2 := uuid.New()
	formData := url.Values{}
	formData.Add("ids", uuid1.String())
	formData.Add("ids", uuid2.String())

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act
	ids, ok := c.FormUUIDs("ids")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, []uuid.UUID{uuid1, uuid2}, ids)
}

func TestShouldParseFormInt(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("age", "25")
	formData.Set("invalid", "not-a-number")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act & Assert
	age, ok := c.FormInt("age")
	assert.True(t, ok)
	assert.Equal(t, 25, age)

	invalid, ok := c.FormInt("invalid")
	assert.False(t, ok)
	assert.Equal(t, 0, invalid)
}

func TestShouldParseFormInts(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Add("numbers", "1")
	formData.Add("numbers", "2")
	formData.Add("numbers", "3")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act
	numbers, ok := c.FormInts("numbers")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, []int{1, 2, 3}, numbers)
}

func TestShouldParseFormInt16(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("value", "1000")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act
	value, ok := c.FormInt16("value")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, int16(1000), value)
}

func TestShouldParseFormInt32(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("value", "100000")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act
	value, ok := c.FormInt32("value")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, int32(100000), value)
}

func TestShouldParseFormInt64(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("value", "10000000000")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act
	value, ok := c.FormInt64("value")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, int64(10000000000), value)
}

func TestShouldParseFormBool(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("enabled", "true")
	formData.Set("disabled", "false")
	formData.Set("invalid", "not-a-bool")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act & Assert
	enabled, ok := c.FormBool("enabled")
	assert.True(t, ok)
	assert.True(t, enabled)

	disabled, ok := c.FormBool("disabled")
	assert.True(t, ok)
	assert.False(t, disabled)

	invalid, ok := c.FormBool("invalid")
	assert.False(t, ok)
	assert.False(t, invalid)
}

func TestShouldParseFormFloat32(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("price", "99.99")
	formData.Set("invalid", "not-a-float")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act & Assert
	price, ok := c.FormFloat32("price")
	assert.True(t, ok)
	assert.Equal(t, float32(99.99), price)

	invalid, ok := c.FormFloat32("invalid")
	assert.False(t, ok)
	assert.Equal(t, float32(0), invalid)
}

func TestShouldParseFormFloat64(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("value", "123.456789")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act
	value, ok := c.FormFloat64("value")

	// Assert
	assert.True(t, ok)
	assert.Equal(t, 123.456789, value)
}

func TestShouldHandleLazyFormParsing(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("name", "test")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act & Assert - first call should parse
	assert.False(t, c.formsParsed)
	
	name1, ok1 := c.FormValue("name")
	assert.True(t, c.formsParsed)
	assert.True(t, ok1)
	assert.Equal(t, "test", name1)

	// Second call should use cached parsed data
	name2, ok2 := c.FormValue("name")
	assert.True(t, ok2)
	assert.Equal(t, "test", name2)
}

func TestShouldReturnFalseForInvalidContentType(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act
	val, ok := c.FormValue("name")

	// Assert
	assert.False(t, ok)
	assert.Equal(t, "", val)
}

func TestShouldHandleMultipartFileField(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// Add a regular field
	writer.WriteField("name", "test")
	
	// Add a file field
	fileWriter, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	fileWriter.Write([]byte("file content"))
	
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c := NewRouteContext(w, req)

	// Act & Assert
	name, ok := c.FormValue("name")
	assert.True(t, ok)
	assert.Equal(t, "test", name)

	// File fields should not appear in form values
	_, ok = c.FormValue("file")
	assert.False(t, ok)
}