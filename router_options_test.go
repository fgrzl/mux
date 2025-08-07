package mux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTitleShouldSetTitle(t *testing.T) {
	// Arrange
	title := "Test API"
	options := &RouterOptions{}

	// Act
	opt := WithTitle(title)
	opt(options)

	// Assert
	assert.NotNil(t, options.openapi)
	assert.Equal(t, title, options.openapi.Title)
}

func TestWithSummaryShouldSetSummary(t *testing.T) {
	// Arrange
	summary := "API Summary"
	options := &RouterOptions{}

	// Act
	opt := WithSummary(summary)
	opt(options)

	// Assert
	assert.NotNil(t, options.openapi)
	assert.Equal(t, summary, options.openapi.Summary)
}

func TestWithDescriptionShouldSetDescription(t *testing.T) {
	// Arrange
	description := "This is a test API description"
	options := &RouterOptions{}

	// Act
	opt := WithDescription(description)
	opt(options)

	// Assert
	assert.NotNil(t, options.openapi)
	assert.Equal(t, description, options.openapi.Description)
}

func TestWithTermsOfServiceShouldSetTermsOfService(t *testing.T) {
	// Arrange
	termsURL := "https://example.com/terms"
	options := &RouterOptions{}

	// Act
	opt := WithTermsOfService(termsURL)
	opt(options)

	// Assert
	assert.NotNil(t, options.openapi)
	assert.Equal(t, termsURL, options.openapi.TermsOfService)
}

func TestWithVersionShouldSetVersion(t *testing.T) {
	// Arrange
	version := "1.2.3"
	options := &RouterOptions{}

	// Act
	opt := WithVersion(version)
	opt(options)

	// Assert
	assert.NotNil(t, options.openapi)
	assert.Equal(t, version, options.openapi.Version)
}

func TestWithContactShouldSetContact(t *testing.T) {
	// Arrange
	name := "API Support"
	url := "https://example.com/support"
	email := "support@example.com"
	options := &RouterOptions{}

	// Act
	opt := WithContact(name, url, email)
	opt(options)

	// Assert
	assert.NotNil(t, options.openapi)
	assert.NotNil(t, options.openapi.Contact)
	assert.Equal(t, name, options.openapi.Contact.Name)
	assert.Equal(t, url, options.openapi.Contact.URL)
	assert.Equal(t, email, options.openapi.Contact.Email)
}

func TestWithLicenseShouldSetLicense(t *testing.T) {
	// Arrange
	name := "MIT"
	url := "https://opensource.org/licenses/MIT"
	options := &RouterOptions{}

	// Act
	opt := WithLicense(name, url)
	opt(options)

	// Assert
	assert.NotNil(t, options.openapi)
	assert.NotNil(t, options.openapi.License)
	assert.Equal(t, name, options.openapi.License.Name)
	assert.Equal(t, url, options.openapi.License.URL)
}

func TestShouldChainMultipleRouterOptions(t *testing.T) {
	// Arrange
	title := "Chained API"
	version := "2.0.0"
	description := "A chained API example"
	options := &RouterOptions{}

	// Act
	WithTitle(title)(options)
	WithVersion(version)(options)
	WithDescription(description)(options)

	// Assert
	assert.NotNil(t, options.openapi)
	assert.Equal(t, title, options.openapi.Title)
	assert.Equal(t, version, options.openapi.Version)
	assert.Equal(t, description, options.openapi.Description)
}

func TestInitInfoShouldCreateInfoObjectWhenNil(t *testing.T) {
	// Arrange
	options := &RouterOptions{}

	// Act
	initInfo(options)

	// Assert
	assert.NotNil(t, options.openapi)
}

func TestInitInfoShouldNotOverrideExistingInfo(t *testing.T) {
	// Arrange
	options := &RouterOptions{
		openapi: &InfoObject{Title: "Existing"},
	}
	original := options.openapi

	// Act
	initInfo(options)

	// Assert
	assert.Equal(t, original, options.openapi)
	assert.Equal(t, "Existing", options.openapi.Title)
}

func TestShouldCreateRouterWithMultipleOptions(t *testing.T) {
	// Arrange & Act
	router := NewRouter(
		WithTitle("Multi-Option API"),
		WithVersion("1.0.0"),
		WithDescription("API with multiple options"),
		WithTermsOfService("https://example.com/terms"),
		WithContact("Support", "https://example.com/support", "support@example.com"),
		WithLicense("MIT", "https://opensource.org/licenses/MIT"),
	)

	// Assert
	assert.NotNil(t, router)
	assert.NotNil(t, router.options.openapi)
	assert.Equal(t, "Multi-Option API", router.options.openapi.Title)
	assert.Equal(t, "1.0.0", router.options.openapi.Version)
	assert.Equal(t, "API with multiple options", router.options.openapi.Description)
	assert.Equal(t, "https://example.com/terms", router.options.openapi.TermsOfService)
	assert.NotNil(t, router.options.openapi.Contact)
	assert.NotNil(t, router.options.openapi.License)
}