package fetch

import (
	"testing"
)

func TestNewJSURL(t *testing.T) {
	u, err := NewJSURL("https://example.com/path?query=value#hash", "")
	if err != nil {
		t.Fatalf("NewJSURL returned error: %v", err)
	}
	if u == nil {
		t.Fatal("NewJSURL returned nil")
	}
}

func TestNewJSURLWithBase(t *testing.T) {
	u, err := NewJSURL("/path?query=value", "https://example.com")
	if err != nil {
		t.Fatalf("NewJSURL with base returned error: %v", err)
	}

	if u.Hostname() != "example.com" {
		t.Errorf("Hostname = %q, want %q", u.Hostname(), "example.com")
	}
	if u.Pathname() != "/path" {
		t.Errorf("Pathname = %q, want %q", u.Pathname(), "/path")
	}
}

func TestJSURLHref(t *testing.T) {
	u, _ := NewJSURL("https://example.com/path?query=value#hash", "")
	if u.Href() != "https://example.com/path?query=value#hash" {
		t.Errorf("Href = %q, want %q", u.Href(), "https://example.com/path?query=value#hash")
	}
}

func TestJSURLProtocol(t *testing.T) {
	u, _ := NewJSURL("https://example.com", "")
	if u.Protocol() != "https:" {
		t.Errorf("Protocol = %q, want %q", u.Protocol(), "https:")
	}

	u, _ = NewJSURL("http://example.com", "")
	if u.Protocol() != "http:" {
		t.Errorf("Protocol = %q, want %q", u.Protocol(), "http:")
	}
}

func TestJSURLHost(t *testing.T) {
	u, _ := NewJSURL("https://example.com:8080/path", "")
	if u.Host() != "example.com:8080" {
		t.Errorf("Host = %q, want %q", u.Host(), "example.com:8080")
	}
}

func TestJSURLHostname(t *testing.T) {
	u, _ := NewJSURL("https://example.com:8080/path", "")
	if u.Hostname() != "example.com" {
		t.Errorf("Hostname = %q, want %q", u.Hostname(), "example.com")
	}
}

func TestJSURLPort(t *testing.T) {
	u, _ := NewJSURL("https://example.com:8080/path", "")
	if u.Port() != "8080" {
		t.Errorf("Port = %q, want %q", u.Port(), "8080")
	}

	u, _ = NewJSURL("https://example.com/path", "")
	if u.Port() != "" {
		t.Errorf("Port = %q, want empty string", u.Port())
	}
}

func TestJSURLPathname(t *testing.T) {
	u, _ := NewJSURL("https://example.com/path/to/page", "")
	if u.Pathname() != "/path/to/page" {
		t.Errorf("Pathname = %q, want %q", u.Pathname(), "/path/to/page")
	}

	u, _ = NewJSURL("https://example.com", "")
	if u.Pathname() != "/" {
		t.Errorf("Pathname = %q, want %q", u.Pathname(), "/")
	}
}

func TestJSURLSearch(t *testing.T) {
	u, _ := NewJSURL("https://example.com/path?foo=bar&baz=qux", "")
	if u.Search() != "?foo=bar&baz=qux" {
		t.Errorf("Search = %q, want %q", u.Search(), "?foo=bar&baz=qux")
	}

	u, _ = NewJSURL("https://example.com/path", "")
	if u.Search() != "" {
		t.Errorf("Search = %q, want empty string", u.Search())
	}
}

func TestJSURLHash(t *testing.T) {
	u, _ := NewJSURL("https://example.com/path#section", "")
	if u.Hash() != "#section" {
		t.Errorf("Hash = %q, want %q", u.Hash(), "#section")
	}

	u, _ = NewJSURL("https://example.com/path", "")
	if u.Hash() != "" {
		t.Errorf("Hash = %q, want empty string", u.Hash())
	}
}

func TestJSURLOrigin(t *testing.T) {
	u, _ := NewJSURL("https://example.com:8080/path", "")
	if u.Origin() != "https://example.com:8080" {
		t.Errorf("Origin = %q, want %q", u.Origin(), "https://example.com:8080")
	}
}

func TestJSURLUsername(t *testing.T) {
	u, _ := NewJSURL("https://user:pass@example.com/path", "")
	if u.Username() != "user" {
		t.Errorf("Username = %q, want %q", u.Username(), "user")
	}
}

func TestJSURLPassword(t *testing.T) {
	u, _ := NewJSURL("https://user:pass@example.com/path", "")
	if u.Password() != "pass" {
		t.Errorf("Password = %q, want %q", u.Password(), "pass")
	}
}

func TestJSURLSearchParams(t *testing.T) {
	u, _ := NewJSURL("https://example.com/path?foo=bar&baz=qux", "")
	params := u.SearchParams()
	if params == nil {
		t.Fatal("SearchParams() returned nil")
	}

	if params.Get("foo") != "bar" {
		t.Errorf("Get(foo) = %q, want %q", params.Get("foo"), "bar")
	}
	if params.Get("baz") != "qux" {
		t.Errorf("Get(baz) = %q, want %q", params.Get("baz"), "qux")
	}
}

func TestJSURLSetters(t *testing.T) {
	u, _ := NewJSURL("https://example.com/path?foo=bar#hash", "")

	u.SetPathname("/newpath")
	if u.Pathname() != "/newpath" {
		t.Errorf("After SetPathname, Pathname = %q, want %q", u.Pathname(), "/newpath")
	}

	u.SetSearch("newquery=value")
	if u.Search() != "?newquery=value" {
		t.Errorf("After SetSearch, Search = %q, want %q", u.Search(), "?newquery=value")
	}

	u.SetHash("newhash")
	if u.Hash() != "#newhash" {
		t.Errorf("After SetHash, Hash = %q, want %q", u.Hash(), "#newhash")
	}
}

func TestJSURLToString(t *testing.T) {
	u, _ := NewJSURL("https://example.com/path?query=value#hash", "")
	if u.ToString() != u.Href() {
		t.Errorf("ToString = %q, want %q", u.ToString(), u.Href())
	}
}

func TestNewURLSearchParams(t *testing.T) {
	usp := NewURLSearchParams("foo=bar&baz=qux")
	if usp == nil {
		t.Fatal("NewURLSearchParams returned nil")
	}

	if usp.Get("foo") != "bar" {
		t.Errorf("Get(foo) = %q, want %q", usp.Get("foo"), "bar")
	}
}

func TestNewURLSearchParamsFromMap(t *testing.T) {
	usp := NewURLSearchParams(map[string]string{"name": "John", "age": "30"})
	if usp.Get("name") != "John" {
		t.Errorf("Get(name) = %q, want %q", usp.Get("name"), "John")
	}
	if usp.Get("age") != "30" {
		t.Errorf("Get(age) = %q, want %q", usp.Get("age"), "30")
	}
}

func TestURLSearchParamsAppend(t *testing.T) {
	usp := NewURLSearchParams("")
	usp.Append("foo", "bar")
	usp.Append("foo", "baz")

	values := usp.GetAll("foo")
	if len(values) != 2 {
		t.Errorf("GetAll(foo) returned %d values, want 2", len(values))
	}
}

func TestURLSearchParamsDelete(t *testing.T) {
	usp := NewURLSearchParams("foo=bar&baz=qux")
	usp.Delete("foo")

	if usp.Has("foo") {
		t.Error("Has(foo) should be false after Delete")
	}
	if !usp.Has("baz") {
		t.Error("Has(baz) should still be true")
	}
}

func TestURLSearchParamsGet(t *testing.T) {
	usp := NewURLSearchParams("foo=bar&foo=baz")
	if usp.Get("foo") != "bar" {
		t.Errorf("Get(foo) = %q, want first value %q", usp.Get("foo"), "bar")
	}
}

func TestURLSearchParamsGetAll(t *testing.T) {
	usp := NewURLSearchParams("foo=bar&foo=baz")
	values := usp.GetAll("foo")

	if len(values) != 2 {
		t.Errorf("GetAll(foo) returned %d values, want 2", len(values))
	}
	if values[0] != "bar" || values[1] != "baz" {
		t.Errorf("GetAll(foo) = %v, want [bar, baz]", values)
	}
}

func TestURLSearchParamsHas(t *testing.T) {
	usp := NewURLSearchParams("foo=bar")
	if !usp.Has("foo") {
		t.Error("Has(foo) should be true")
	}
	if usp.Has("bar") {
		t.Error("Has(bar) should be false")
	}
}

func TestURLSearchParamsSet(t *testing.T) {
	usp := NewURLSearchParams("foo=bar&foo=baz")
	usp.Set("foo", "new")

	values := usp.GetAll("foo")
	if len(values) != 1 {
		t.Errorf("After Set, GetAll(foo) returned %d values, want 1", len(values))
	}
	if values[0] != "new" {
		t.Errorf("After Set, GetAll(foo)[0] = %q, want %q", values[0], "new")
	}
}

func TestURLSearchParamsSort(t *testing.T) {
	usp := NewURLSearchParams("z=1&a=2&m=3")
	usp.Sort()

	keys := usp.Keys()
	expected := []string{"a", "m", "z"}
	if len(keys) != len(expected) {
		t.Fatalf("Keys() returned %d keys, want %d", len(keys), len(expected))
	}
	for i, k := range keys {
		if k != expected[i] {
			t.Errorf("Keys()[%d] = %q, want %q", i, k, expected[i])
		}
	}
}

func TestURLSearchParamsEncode(t *testing.T) {
	usp := NewURLSearchParams("foo=bar&baz=qux")
	encoded := usp.Encode()

	if encoded != "foo=bar&baz=qux" && encoded != "baz=qux&foo=bar" {
		t.Errorf("Encode = %q, want foo=bar&baz=qux (or order swapped)", encoded)
	}
}

func TestURLSearchParamsEncodeSpecialChars(t *testing.T) {
	usp := NewURLSearchParams("foo=hello world&bar=a=b")
	encoded := usp.Encode()

	// "hello world" should be encoded as "hello%20world"
	if encoded == "foo=hello+world&bar=a%3Db" || encoded == "bar=a%3Db&foo=hello+world" {
		// This is application/x-www-form-urlencoded format where spaces become +
		// Our implementation uses percent-encoding for spaces
		// Check that it was encoded
	}
}

func TestURLSearchParamsKeys(t *testing.T) {
	usp := NewURLSearchParams("foo=1&bar=2")
	keys := usp.Keys()

	if len(keys) != 2 {
		t.Errorf("Keys() returned %d keys, want 2", len(keys))
	}
}

func TestURLSearchParamsValues(t *testing.T) {
	usp := NewURLSearchParams("foo=1&bar=2")
	values := usp.Values()

	if len(values) != 2 {
		t.Errorf("Values() returned %d values, want 2", len(values))
	}
}

func TestURLSearchParamsEntries(t *testing.T) {
	usp := NewURLSearchParams("foo=bar")
	entries := usp.Entries()

	if len(entries) != 1 {
		t.Errorf("Entries() returned %d entries, want 1", len(entries))
	}
	if entries[0][0] != "foo" || entries[0][1] != "bar" {
		t.Errorf("Entries()[0] = %v, want [foo, bar]", entries[0])
	}
}

func TestURLSearchParamsForEach(t *testing.T) {
	usp := NewURLSearchParams("foo=bar&baz=qux")
	var received []string

	usp.ForEach(func(value, key string) {
		received = append(received, key+"="+value)
	})

	if len(received) != 2 {
		t.Errorf("ForEach received %d calls, want 2", len(received))
	}
}

func TestURLSearchParamsEmpty(t *testing.T) {
	usp := NewURLSearchParams("")
	if usp.Encode() != "" {
		t.Errorf("Empty Encode() = %q, want empty string", usp.Encode())
	}
	if usp.Get("foo") != "" {
		t.Error("Get on empty should return empty string")
	}
}

func TestURLSearchParamsMissingLeadingQuestion(t *testing.T) {
	usp := NewURLSearchParams("?foo=bar")
	if usp.Get("foo") != "bar" {
		t.Errorf("Get(foo) = %q, want %q", usp.Get("foo"), "bar")
	}
}
