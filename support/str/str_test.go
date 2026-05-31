package str

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dracory/neat/support/env"
)

// TestAfter tests the After function.
func TestAfter(t *testing.T) {
	// Test After: returns substring after first occurrence of search string
	testCases := []struct {
		input    string
		search   string
		expected string
	}{
		{"GoravelFramework", "Goravel", "Framework"},
		{"parallel", "l", "lel"},
		{"abc123def", "2", "3def"},
		{"abc123def", "4", "abc123def"},
		{"GoravelFramework", "", "GoravelFramework"},
	}

	for _, tc := range testCases {
		result := after(tc.input, tc.search)
		assertEqual(t, tc.expected, result)
	}
}

// TestAfterLast tests the AfterLast function.
func TestAfterLast(t *testing.T) {
	// Test AfterLast: returns substring after last occurrence of search string
	testCases := []struct {
		input    string
		search   string
		expected string
	}{
		{"GoravelFramework", "Goravel", "Framework"},
		{"parallel", "l", ""},
		{"abc123def", "2", "3def"},
		{"abc123def", "4", "abc123def"},
	}

	for _, tc := range testCases {
		result := afterLast(tc.input, tc.search)
		assertEqual(t, tc.expected, result)
	}
}

// after returns the substring after the first occurrence of search
func after(s, search string) string {
	if search == "" {
		return s
	}
	_, after, ok := strings.Cut(s, search)
	if ok {
		return after
	}
	return s
}

// afterLast returns the substring after the last occurrence of search
func afterLast(s, search string) string {
	index := strings.LastIndex(s, search)
	if index != -1 {
		return s[index+len(search):]
	}
	return s
}

// TestAppend tests the Append function.
func TestAppend(t *testing.T) {
	// Test Append: appends strings together
	assertEqual(t, "foobar", appendStr("foo", "bar"))
	assertEqual(t, "foobar", appendStr("foo", "bar", ""))
	assertEqual(t, "foobar", appendStr("foo", "bar"))
}

// TestBasename tests the Basename function.
func TestBasename(t *testing.T) {
	// Test Basename: returns the basename of a file path
	testCases := []struct {
		input    string
		suffix   string
		expected string
	}{
		{"/framework/support/str", "", "str"},
		{"/framework/support/str/", "", "str"},
		{"str", "", "str"},
		{"/str", "", "str"},
		{"/str/", "", "str"},
		{"str/", "", "str"},
	}

	for _, tc := range testCases {
		result := basename(tc.input, tc.suffix)
		assertEqual(t, tc.expected, result)
	}

	// Test root path
	str := basename("/", "")
	if env.IsWindows() {
		assertEqual(t, "\\", str)
	} else {
		assertEqual(t, "/", str)
	}

	// Test empty string
	assertEqual(t, ".", basename("", ""))

	// Test with suffix
	assertEqual(t, "str", basename("/framework/support/str/str.go", ".go"))
}

// TestBefore tests the Before function.
func TestBefore(t *testing.T) {
	// Test Before: returns substring before first occurrence of search string
	testCases := []struct {
		input    string
		search   string
		expected string
	}{
		{"GoravelFramework", "Framework", "Goravel"},
		{"parallel", "l", "para"},
		{"abc123def", "def", "abc123"},
		{"abc123def", "123", "abc"},
	}

	for _, tc := range testCases {
		result := before(tc.input, tc.search)
		assertEqual(t, tc.expected, result)
	}
}

// appendStr joins strings together
func appendStr(base string, values ...string) string {
	return base + strings.Join(values, "")
}

// basename returns the basename of a file path with optional suffix trimming
func basename(path string, suffix string) string {
	result := filepath.Base(path)
	if suffix != "" {
		result = strings.TrimSuffix(result, suffix)
	}
	return result
}

// before returns the substring before the first occurrence of search
func before(s, search string) string {
	index := strings.Index(s, search)
	if index != -1 {
		return s[:index]
	}
	return s
}

// Helper functions for standard Go testing
func assertEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func assertEqualStringSlice(t *testing.T, expected, actual []string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf("Expected slice length %d, got %d", len(expected), len(actual))
		return
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("At index %d: expected %v, got %v", i, expected[i], actual[i])
		}
	}
}

func assertTrue(t *testing.T, condition bool) {
	t.Helper()
	if !condition {
		t.Error("Expected true, got false")
	}
}

func assertFalse(t *testing.T, condition bool) {
	t.Helper()
	if condition {
		t.Error("Expected false, got true")
	}
}

func assertLen(t *testing.T, obj interface{}, expected int) {
	t.Helper()
	var actual int
	switch v := obj.(type) {
	case string:
		actual = len(v)
	case []string:
		actual = len(v)
	case []int:
		actual = len(v)
	default:
		t.Errorf("Unsupported type for assertLen: %T", obj)
		return
	}
	if actual != expected {
		t.Errorf("Expected length %d, got %d", expected, actual)
	}
}

func assertEmpty(t *testing.T, obj interface{}) {
	t.Helper()
	var isEmpty bool
	switch v := obj.(type) {
	case string:
		isEmpty = v == ""
	case []string:
		isEmpty = len(v) == 0
	case []int:
		isEmpty = len(v) == 0
	default:
		t.Errorf("Unsupported type for assertEmpty: %T", obj)
		return
	}
	if !isEmpty {
		t.Error("Expected empty, but was not empty")
	}
}

func assertPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but function did not panic")
		}
	}()
	f()
}

// TestBeforeLast tests the BeforeLast function.
func TestBeforeLast(t *testing.T) {
	assertEqual(t, "Goravel", Of("GoravelFramework").BeforeLast("Framework").String())
	assertEqual(t, "paralle", Of("parallel").BeforeLast("l").String())
	assertEqual(t, "abc123", Of("abc123def").BeforeLast("def").String())
	assertEqual(t, "abc", Of("abc123def").BeforeLast("123").String())
}

// TestBetween tests the Between function.
func TestBetween(t *testing.T) {
	assertEqual(t, "foobarbaz", Of("foobarbaz").Between("", "b").String())
	assertEqual(t, "foobarbaz", Of("foobarbaz").Between("f", "").String())
	assertEqual(t, "foobarbaz", Of("foobarbaz").Between("", "").String())
	assertEqual(t, "obar", Of("foobarbaz").Between("o", "b").String())
	assertEqual(t, "bar", Of("foobarbaz").Between("foo", "baz").String())
	assertEqual(t, "foo][bar][baz", Of("[foo][bar][baz]").Between("[", "]").String())
}

// TestBetweenFirst tests the BetweenFirst function.
func TestBetweenFirst(t *testing.T) {
	assertEqual(t, "foobarbaz", Of("foobarbaz").BetweenFirst("", "b").String())
	assertEqual(t, "foobarbaz", Of("foobarbaz").BetweenFirst("f", "").String())
	assertEqual(t, "foobarbaz", Of("foobarbaz").BetweenFirst("", "").String())
	assertEqual(t, "o", Of("foobarbaz").BetweenFirst("o", "b").String())
	assertEqual(t, "foo", Of("[foo][bar][baz]").BetweenFirst("[", "]").String())
	assertEqual(t, "foobar", Of("foofoobarbaz").BetweenFirst("foo", "baz").String())
}

// TestCamel tests the Camel function.
func TestCamel(t *testing.T) {
	assertEqual(t, "goravelGOFramework", Of("Goravel_g_o_framework").Camel().String())
	assertEqual(t, "goravelGOFramework", Of("Goravel_gO_framework").Camel().String())
	assertEqual(t, "goravelGoFramework", Of("Goravel -_- go -_-  framework  ").Camel().String())

	assertEqual(t, "fooBar", Of("FooBar").Camel().String())
	assertEqual(t, "fooBar", Of("foo_bar").Camel().String())
	assertEqual(t, "fooBar", Of("foo-Bar").Camel().String())
	assertEqual(t, "fooBar", Of("foo bar").Camel().String())
	assertEqual(t, "fooBar", Of("foo.bar").Camel().String())
}

// TestCharAt tests the CharAt function.
func TestCharAt(t *testing.T) {
	assertEqual(t, "好", Of("你好，世界！").CharAt(1))
	assertEqual(t, "त", Of("नमस्ते, दुनिया!").CharAt(4))
	assertEqual(t, "w", Of("Привет, world!").CharAt(8))
	assertEqual(t, "계", Of("안녕하세요, 세계!").CharAt(-2))
	assertEqual(t, "", Of("こんにちは、世界！").CharAt(-200))
}

// TestChopEnd tests the ChopEnd function.
func TestChopEnd(t *testing.T) {
	assertEqual(t, "Goravel", Of("GoravelFramework").ChopEnd("Framework").String())
	assertEqual(t, "https://goravel", Of("https://goravel.dev").ChopEnd(".dev").String())
	assertEqual(t, "https://goravel", Of("https://goravel.dev").ChopEnd(".dev", ".com").String())
	assertEqual(t, "https://goravel", Of("https://goravel.com").ChopEnd(".dev", ".com").String())
	assertEqual(t, "go", Of("golaravel").ChopEnd("laravel").String())
}

// TestChopStart tests the ChopStart function.
func TestChopStart(t *testing.T) {
	assertEqual(t, "Framework", Of("GoravelFramework").ChopStart("Goravel").String())
	assertEqual(t, "goravel.dev", Of("https://goravel.dev").ChopStart("https://").String())
	assertEqual(t, "goravel.dev", Of("https://goravel.dev").ChopStart("https://", "http://").String())
	assertEqual(t, "goravel.dev", Of("http://goravel.dev").ChopStart("https://", "http://").String())
	assertEqual(t, "goravel", "go"+Of("laravel").ChopStart("la").String())
}

// TestContains tests the Contains function.
func TestContains(t *testing.T) {
	assertTrue(t, Of("kkumar").Contains("uma"))
	assertTrue(t, Of("kkumar").Contains("kumar"))
	assertTrue(t, Of("kkumar").Contains("uma", "xyz"))
	assertFalse(t, Of("kkumar").Contains("xyz"))
	assertFalse(t, Of("kkumar").Contains(""))
}

// TestContainsAll tests the ContainsAll function.
func TestContainsAll(t *testing.T) {
	assertTrue(t, Of("krishan kumar").ContainsAll("krishan", "kumar"))
	assertTrue(t, Of("krishan kumar").ContainsAll("kumar"))
	assertFalse(t, Of("krishan kumar").ContainsAll("kumar", "xyz"))
}

// TestDirname tests the Dirname function.
func TestDirname(t *testing.T) {
	str := Of("/framework/support/str").Dirname().String()
	if env.IsWindows() {
		assertEqual(t, "\\framework\\support", str)
	} else {
		assertEqual(t, "/framework/support", str)
	}

	str = Of("/framework/support/str").Dirname(2).String()
	if env.IsWindows() {
		assertEqual(t, "\\framework", str)
	} else {
		assertEqual(t, "/framework", str)
	}

	assertEqual(t, ".", Of("framework").Dirname().String())
	assertEqual(t, ".", Of(".").Dirname().String())

	str = Of("/").Dirname().String()
	if env.IsWindows() {
		assertEqual(t, "\\", str)
	} else {
		assertEqual(t, "/", str)
	}

	str = Of("/framework/").Dirname(2).String()
	if env.IsWindows() {
		assertEqual(t, "\\", str)
	} else {
		assertEqual(t, "/", str)
	}
}

// TestEndsWith tests the EndsWith function.
func TestEndsWith(t *testing.T) {
	assertTrue(t, Of("bowen").EndsWith("wen"))
	assertTrue(t, Of("bowen").EndsWith("bowen"))
	assertTrue(t, Of("bowen").EndsWith("wen", "xyz"))
	assertFalse(t, Of("bowen").EndsWith("xyz"))
	assertFalse(t, Of("bowen").EndsWith(""))
	assertFalse(t, Of("bowen").EndsWith())
	assertFalse(t, Of("bowen").EndsWith("N"))
	assertTrue(t, Of("a7.12").EndsWith("7.12"))
	// Test for muti-byte string
	assertTrue(t, Of("你好").EndsWith("好"))
	assertTrue(t, Of("你好").EndsWith("你好"))
	assertTrue(t, Of("你好").EndsWith("好", "xyz"))
	assertFalse(t, Of("你好").EndsWith("xyz"))
	assertFalse(t, Of("你好").EndsWith(""))
}

// TestExactly tests the Exactly function.
func TestExactly(t *testing.T) {
	assertTrue(t, Of("foo").Exactly("foo"))
	assertFalse(t, Of("foo").Exactly("Foo"))
}

// TestExcerpt tests the Excerpt function.
func TestExcerpt(t *testing.T) {
	assertEqual(t, "...is a beautiful morn...", Of("This is a beautiful morning").Excerpt("beautiful", ExcerptOption{
		Radius: 5,
	}).String())
	assertEqual(t, "This is a beautiful morning", Of("This is a beautiful morning").Excerpt("foo", ExcerptOption{
		Radius: 5,
	}).String())
	assertEqual(t, "(...)is a beautiful morn(...)", Of("This is a beautiful morning").Excerpt("beautiful", ExcerptOption{
		Omission: "(...)",
		Radius:   5,
	}).String())
}

// TestExplode tests the Explode function.
func TestExplode(t *testing.T) {
	assertEqualStringSlice(t, []string{"Foo", "Bar", "Baz"}, Of("Foo Bar Baz").Explode(" "))
	// with limit
	assertEqualStringSlice(t, []string{"Foo", "Bar Baz"}, Of("Foo Bar Baz").Explode(" ", 2))
	assertEqualStringSlice(t, []string{"Foo", "Bar"}, Of("Foo Bar Baz").Explode(" ", -1))
	assertEqualStringSlice(t, []string{}, Of("Foo Bar Baz").Explode(" ", -10))
}

// TestFinish tests the Finish function.
func TestFinish(t *testing.T) {
	assertEqual(t, "abbc", Of("ab").Finish("bc").String())
	assertEqual(t, "abbc", Of("abbcbc").Finish("bc").String())
	assertEqual(t, "abcbbc", Of("abcbbcbc").Finish("bc").String())
}

// TestHeadline tests the Headline function.
func TestHeadline(t *testing.T) {
	assertEqual(t, "Hello", Of("hello").Headline().String())
	assertEqual(t, "This Is A Headline", Of("this is a headline").Headline().String())
	assertEqual(t, "Camelcase Is A Headline", Of("CamelCase is a headline").Headline().String())
	assertEqual(t, "Kebab-Case Is A Headline", Of("kebab-case is a headline").Headline().String())
}

// TestIs tests the Is function.
func TestIs(t *testing.T) {
	assertTrue(t, Of("foo").Is("foo", "bar", "baz"))
	assertTrue(t, Of("foo123").Is("bar*", "baz*", "foo*"))
	assertFalse(t, Of("foo").Is("bar", "baz"))
	assertTrue(t, Of("a.b").Is("a.b", "c.*"))
	assertFalse(t, Of("abc*").Is("abc\\*", "xyz*"))
	assertFalse(t, Of("").Is("foo"))
	assertTrue(t, Of("foo/bar/baz").Is("foo/*", "bar/*", "baz*"))
	// Is case-sensitive
	assertFalse(t, Of("foo/bar/baz").Is("*BAZ*"))
}

// TestIsEmpty tests the IsEmpty function.
func TestIsEmpty(t *testing.T) {
	assertTrue(t, Of("").IsEmpty())
	assertFalse(t, Of("F").IsEmpty())
}

// TestIsNotEmpty tests the IsNotEmpty function.
func TestIsNotEmpty(t *testing.T) {
	assertFalse(t, Of("").IsNotEmpty())
	assertTrue(t, Of("F").IsNotEmpty())
}

// TestIsAscii tests the IsAscii function.
func TestIsAscii(t *testing.T) {
	assertTrue(t, Of("abc").IsAscii())
	assertFalse(t, Of("你好").IsAscii())
}

// TestIsSlice tests the IsSlice function.
func TestIsSlice(t *testing.T) {
	// Test when the string represents a valid JSON array
	assertTrue(t, Of(`["apple", "banana", "cherry"]`).IsSlice())

	// Test when the string represents a valid JSON array with objects
	assertTrue(t, Of(`[{"name": "John"}, {"name": "Alice"}]`).IsSlice())

	// Test when the string represents an empty JSON array
	assertTrue(t, Of(`[]`).IsSlice())

	// Test when the string represents an invalid JSON object
	assertFalse(t, Of(`{"name": "John"}`).IsSlice())

	// Test when the string is not valid JSON
	assertFalse(t, Of(`Not a JSON array`).IsSlice())

	// Test when the string is empty
	assertFalse(t, Of("").IsSlice())
}

// TestIsMap tests the IsMap function.
func TestIsMap(t *testing.T) {
	// Test when the string represents a valid JSON object
	assertTrue(t, Of(`{"name": "John", "age": 30}`).IsMap())

	// Test when the string represents a valid JSON object with nested objects
	assertTrue(t, Of(`{"person": {"name": "Alice", "age": 25}}`).IsMap())

	// Test when the string represents an empty JSON object
	assertTrue(t, Of(`{}`).IsMap())

	// Test when the string represents an invalid JSON array
	assertFalse(t, Of(`["apple", "banana", "cherry"]`).IsMap())

	// Test when the string is not valid JSON
	assertFalse(t, Of(`Not a JSON object`).IsMap())

	// Test when the string is empty
	assertFalse(t, Of("").IsMap())
}

// TestIsUlid tests the IsUlid function.
func TestIsUlid(t *testing.T) {
	assertTrue(t, Of("01E65Z7XCHCR7X1P2MKF78ENRP").IsUlid())
	// lowercase characters are not allowed
	assertFalse(t, Of("01e65z7xchcr7x1p2mkf78enrp").IsUlid())
	// too short (ULIDS must be 26 characters long)
	assertFalse(t, Of("01E65Z7XCHCR7X1P2MKF78E").IsUlid())
	// contains invalid characters
	assertFalse(t, Of("01E65Z7XCHCR7X1P2MKF78ENR!").IsUlid())
}

// TestIsUuid tests the IsUuid function.
func TestIsUuid(t *testing.T) {
	assertTrue(t, Of("3f2504e0-4f89-41d3-9a0c-0305e82c3301").IsUuid())
	assertFalse(t, Of("3f2504e0-4f89-41d3-9a0c-0305e82c3301-extra").IsUuid())
}

// TestKebab tests the Kebab function.
func TestKebab(t *testing.T) {
	assertEqual(t, "goravel-framework", Of("GoravelFramework").Kebab().String())
}

// TestLcFirst tests the LcFirst function.
func TestLcFirst(t *testing.T) {
	assertEqual(t, "framework", Of("Framework").LcFirst().String())
	assertEqual(t, "framework", Of("framework").LcFirst().String())
}

// TestLength tests the Length function.
func TestLength(t *testing.T) {
	assertEqual(t, 11, Of("foo bar baz").Length())
	assertEqual(t, 0, Of("").Length())
}

// TestLimit tests the Limit function.
func TestLimit(t *testing.T) {
	assertEqual(t, "This is...", Of("This is a beautiful morning").Limit(7).String())
	assertEqual(t, "This is****", Of("This is a beautiful morning").Limit(7, "****").String())
	assertEqual(t, "这是一...", Of("这是一段中文").Limit(3).String())
	assertEqual(t, "这是一段中文", Of("这是一段中文").Limit(9).String())
}

// TestLower tests the Lower function.
func TestLower(t *testing.T) {
	assertEqual(t, "foo bar baz", Of("FOO BAR BAZ").Lower().String())
	assertEqual(t, "foo bar baz", Of("fOo Bar bAz").Lower().String())
}

// TestLTrim tests the LTrim function.
func TestLTrim(t *testing.T) {
	assertEqual(t, "foo ", Of(" foo ").LTrim().String())
}

// TestMask tests the Mask function.
func TestMask(t *testing.T) {
	assertEqual(t, "kri**************", Of("krishan@email.com").Mask("*", 3).String())
	assertEqual(t, "*******@email.com", Of("krishan@email.com").Mask("*", 0, 7).String())
	assertEqual(t, "kris*************", Of("krishan@email.com").Mask("*", -13).String())
	assertEqual(t, "kris***@email.com", Of("krishan@email.com").Mask("*", -13, 3).String())

	assertEqual(t, "*****************", Of("krishan@email.com").Mask("*", -17).String())
	assertEqual(t, "*****an@email.com", Of("krishan@email.com").Mask("*", -99, 5).String())

	assertEqual(t, "krishan@email.com", Of("krishan@email.com").Mask("*", 17).String())
	assertEqual(t, "krishan@email.com", Of("krishan@email.com").Mask("*", 17, 99).String())

	assertEqual(t, "krishan@email.com", Of("krishan@email.com").Mask("", 3).String())

	assertEqual(t, "krissssssssssssss", Of("krishan@email.com").Mask("something", 3).String())

	assertEqual(t, "这是****", Of("这是一段中文").Mask("*", 2).String())
	assertEqual(t, "**一段中文", Of("这是一段中文").Mask("*", 0, 2).String())
}

// TestMatch tests the Match function.
func TestMatch(t *testing.T) {
	assertEqual(t, "World", Of("Hello, World!").Match("World").String())
	assertEqual(t, "(test)", Of("This is a (test) string").Match(`\([^)]+\)`).String())
	assertEqual(t, "123", Of("abc123def456def").Match(`\d+`).String())
	assertEqual(t, "", Of("No match here").Match(`\d+`).String())
	assertEqual(t, "Hello, World!", Of("Hello, World!").Match("").String())
	assertEqual(t, "[456]", Of("123 [456]").Match(`\[456\]`).String())
}

// TestMatchAll tests the MatchAll function.
func TestMatchAll(t *testing.T) {
	assertEqualStringSlice(t, []string{"World"}, Of("Hello, World!").MatchAll("World"))
	assertEqualStringSlice(t, []string{"(test)"}, Of("This is a (test) string").MatchAll(`\([^)]+\)`))
	assertEqualStringSlice(t, []string{"123", "456"}, Of("abc123def456def").MatchAll(`\d+`))
	assertEqualStringSlice(t, []string(nil), Of("No match here").MatchAll(`\d+`))
	assertEqualStringSlice(t, []string{"Hello, World!"}, Of("Hello, World!").MatchAll(""))
	assertEqualStringSlice(t, []string{"[456]"}, Of("123 [456]").MatchAll(`\[456\]`))
}

// TestIsMatch tests the IsMatch function.
func TestIsMatch(t *testing.T) {
	// Test matching with a single pattern
	assertTrue(t, Of("Hello, Goravel!").IsMatch(`.*,.*!`))
	assertTrue(t, Of("Hello, Goravel!").IsMatch(`^.*$(.*)`))
	assertTrue(t, Of("Hello, Goravel!").IsMatch(`(?i)goravel`))
	assertTrue(t, Of("Hello, GOravel!").IsMatch(`^(.*(.*(.*)))`))

	// Test non-matching with a single pattern
	assertFalse(t, Of("Hello, Goravel!").IsMatch(`H.o`))
	assertFalse(t, Of("Hello, Goravel!").IsMatch(`^goravel!`))
	assertFalse(t, Of("Hello, Goravel!").IsMatch(`goravel!(.*)`))
	assertFalse(t, Of("Hello, Goravel!").IsMatch(`^[a-zA-Z,!]+$`))

	// Test with multiple patterns
	assertTrue(t, Of("Hello, Goravel!").IsMatch(`.*,.*!`, `H.o`))
	assertTrue(t, Of("Hello, Goravel!").IsMatch(`(?i)goravel`, `^.*$(.*)`))
	assertTrue(t, Of("Hello, Goravel!").IsMatch(`(?i)goravel`, `goravel!(.*)`))
	assertTrue(t, Of("Hello, Goravel!").IsMatch(`^[a-zA-Z,!]+$`, `^(.*(.*(.*)))`))
}

// TestNewLine tests the NewLine function.
func TestNewLine(t *testing.T) {
	assertEqual(t, "Goravel\n", Of("Goravel").NewLine().String())
	assertEqual(t, "Goravel\n\nbar", Of("Goravel").NewLine(2).Append("bar").String())
}

// TestPadBoth tests the PadBoth function.
func TestPadBoth(t *testing.T) {
	// Test padding with spaces
	assertEqual(t, "   Hello   ", Of("Hello").PadBoth(11, " ").String())
	assertEqual(t, "  World!  ", Of("World!").PadBoth(10, " ").String())
	assertEqual(t, "==Hello===", Of("Hello").PadBoth(10, "=").String())
	assertEqual(t, "Hello", Of("Hello").PadBoth(3, " ").String())
	assertEqual(t, "      ", Of("").PadBoth(6, " ").String())
}

// TestPadLeft tests the PadLeft function.
func TestPadLeft(t *testing.T) {
	assertEqual(t, "   Goravel", Of("Goravel").PadLeft(10, " ").String())
	assertEqual(t, "==Goravel", Of("Goravel").PadLeft(9, "=").String())
	assertEqual(t, "Goravel", Of("Goravel").PadLeft(3, " ").String())
}

// TestPadRight tests the PadRight function.
func TestPadRight(t *testing.T) {
	assertEqual(t, "Goravel   ", Of("Goravel").PadRight(10, " ").String())
	assertEqual(t, "Goravel==", Of("Goravel").PadRight(9, "=").String())
	assertEqual(t, "Goravel", Of("Goravel").PadRight(3, " ").String())
}

// TestPipe tests the Pipe function.
func TestPipe(t *testing.T) {
	callback := func(str string) string {
		return Of(str).Append("bar").String()
	}
	assertEqual(t, "foobar", Of("foo").Pipe(callback).String())
}

// TestPrepend tests the Prepend function.
func TestPrepend(t *testing.T) {
	assertEqual(t, "foobar", Of("bar").Prepend("foo").String())
	assertEqual(t, "foobar", Of("bar").Prepend("foo").Prepend("").String())
	assertEqual(t, "foobar", Of("bar").Prepend("foo").Prepend().String())
}

// TestRemove tests the Remove function.
func TestRemove(t *testing.T) {
	assertEqual(t, "Fbar", Of("Foobar").Remove("o").String())
	assertEqual(t, "Foo", Of("Foobar").Remove("bar").String())
	assertEqual(t, "oobar", Of("Foobar").Remove("F").String())
	assertEqual(t, "Foobar", Of("Foobar").Remove("f").String())

	assertEqual(t, "Fbr", Of("Foobar").Remove("o", "a").String())
	assertEqual(t, "Fooar", Of("Foobar").Remove("f", "b").String())
	assertEqual(t, "Foobar", Of("Foo|bar").Remove("f", "|").String())
}

// TestRepeat tests the Repeat function.
func TestRepeat(t *testing.T) {
	assertEqual(t, "aaaaa", Of("a").Repeat(5).String())
	assertEqual(t, "", Of("").Repeat(5).String())
}

// TestReplace tests the Replace function.
func TestReplace(t *testing.T) {
	assertEqual(t, "foo/foo/foo", Of("?/?/?").Replace("?", "foo").String())
	assertEqual(t, "foo/foo/foo", Of("x/x/x").Replace("X", "foo", false).String())
	assertEqual(t, "bar/bar", Of("?/?").Replace("?", "bar").String())
	assertEqual(t, "?/?/?", Of("? ? ?").Replace(" ", "/").String())
}

// TestReplaceEnd tests the ReplaceEnd function.
func TestReplaceEnd(t *testing.T) {
	assertEqual(t, "Golang is great!", Of("Golang is good!").ReplaceEnd("good!", "great!").String())
	assertEqual(t, "Hello, World!", Of("Hello, Earth!").ReplaceEnd("Earth!", "World!").String())
	assertEqual(t, "München Berlin", Of("München Frankfurt").ReplaceEnd("Frankfurt", "Berlin").String())
	assertEqual(t, "Café Latte", Of("Café Americano").ReplaceEnd("Americano", "Latte").String())
	assertEqual(t, "Golang is good!", Of("Golang is good!").ReplaceEnd("", "great!").String())
	assertEqual(t, "Golang is good!", Of("Golang is good!").ReplaceEnd("excellent!", "great!").String())
}

// TestReplaceFirst tests the ReplaceFirst function.
func TestReplaceFirst(t *testing.T) {
	assertEqual(t, "fooqux foobar", Of("foobar foobar").ReplaceFirst("bar", "qux").String())
	assertEqual(t, "foo/qux? foo/bar?", Of("foo/bar? foo/bar?").ReplaceFirst("bar?", "qux?").String())
	assertEqual(t, "foo foobar", Of("foobar foobar").ReplaceFirst("bar", "").String())
	assertEqual(t, "foobar foobar", Of("foobar foobar").ReplaceFirst("xxx", "yyy").String())
	assertEqual(t, "foobar foobar", Of("foobar foobar").ReplaceFirst("", "yyy").String())
	// Test for multibyte string support
	assertEqual(t, "Jxxxnköping Malmö", Of("Jönköping Malmö").ReplaceFirst("ö", "xxx").String())
	assertEqual(t, "Jönköping Malmö", Of("Jönköping Malmö").ReplaceFirst("", "yyy").String())
}

// TestReplaceLast tests the ReplaceLast function.
func TestReplaceLast(t *testing.T) {
	assertEqual(t, "foobar fooqux", Of("foobar foobar").ReplaceLast("bar", "qux").String())
	assertEqual(t, "foo/bar? foo/qux?", Of("foo/bar? foo/bar?").ReplaceLast("bar?", "qux?").String())
	assertEqual(t, "foobar foo", Of("foobar foobar").ReplaceLast("bar", "").String())
	assertEqual(t, "foobar foobar", Of("foobar foobar").ReplaceLast("xxx", "yyy").String())
	assertEqual(t, "foobar foobar", Of("foobar foobar").ReplaceLast("", "yyy").String())
	// Test for multibyte string support
	assertEqual(t, "Malmö Jönkxxxping", Of("Malmö Jönköping").ReplaceLast("ö", "xxx").String())
	assertEqual(t, "Malmö Jönköping", Of("Malmö Jönköping").ReplaceLast("", "yyy").String())
}

// TestReplaceMatches tests the ReplaceMatches function.
func TestReplaceMatches(t *testing.T) {
	assertEqual(t, "Golang is great!", Of("Golang is good!").ReplaceMatches("good", "great").String())
	assertEqual(t, "Hello, World!", Of("Hello, Earth!").ReplaceMatches("Earth", "World").String())
	assertEqual(t, "Apples, Apples, Apples", Of("Oranges, Oranges, Oranges").ReplaceMatches("Oranges", "Apples").String())
	assertEqual(t, "1, 2, 3, 4, 5", Of("10, 20, 30, 40, 50").ReplaceMatches("0", "").String())
	assertEqual(t, "München Berlin", Of("München Frankfurt").ReplaceMatches("Frankfurt", "Berlin").String())
	assertEqual(t, "Café Latte", Of("Café Americano").ReplaceMatches("Americano", "Latte").String())
	assertEqual(t, "The quick brown fox", Of("The quick brown fox").ReplaceMatches(`\b([a-z])`, `$1`).String())
	assertEqual(t, "One, One, One", Of("1, 2, 3").ReplaceMatches(`\d`, "One").String())
	assertEqual(t, "Hello, World!", Of("Hello, World!").ReplaceMatches("Earth", "").String())
	assertEqual(t, "Hello, World!", Of("Hello, World!").ReplaceMatches("Golang", "Great").String())
}

// TestReplaceStart tests the ReplaceStart function.
func TestReplaceStart(t *testing.T) {
	assertEqual(t, "foobar foobar", Of("foobar foobar").ReplaceStart("bar", "qux").String())
	assertEqual(t, "foo/bar? foo/bar?", Of("foo/bar? foo/bar?").ReplaceStart("bar?", "qux?").String())
	assertEqual(t, "quxbar foobar", Of("foobar foobar").ReplaceStart("foo", "qux").String())
	assertEqual(t, "qux? foo/bar?", Of("foo/bar? foo/bar?").ReplaceStart("foo/bar?", "qux?").String())
	assertEqual(t, "bar foobar", Of("foobar foobar").ReplaceStart("foo", "").String())
	assertEqual(t, "1", Of("0").ReplaceStart("0", "1").String())
	// Test for multibyte string support
	assertEqual(t, "xxxnköping Malmö", Of("Jönköping Malmö").ReplaceStart("Jö", "xxx").String())
	assertEqual(t, "Jönköping Malmö", Of("Jönköping Malmö").ReplaceStart("", "yyy").String())
}

// TestRTrim tests the RTrim function.
func TestRTrim(t *testing.T) {
	assertEqual(t, " foo", Of(" foo ").RTrim().String())
	assertEqual(t, " foo", Of(" foo__").RTrim("_").String())
}

// TestSnake tests the Snake function.
func TestSnake(t *testing.T) {
	assertEqual(t, "goravel_g_o_framework", Of("GoravelGOFramework").Snake().String())
	assertEqual(t, "goravel_go_framework", Of("GoravelGoFramework").Snake().String())
	assertEqual(t, "goravel go framework", Of("GoravelGoFramework").Snake(" ").String())
	assertEqual(t, "goravel_go_framework", Of("Goravel Go Framework").Snake().String())
	assertEqual(t, "goravel_go_framework", Of("Goravel    Go      Framework   ").Snake().String())
	assertEqual(t, "goravel__go__framework", Of("GoravelGoFramework").Snake("__").String())
	assertEqual(t, "żółta_łódka", Of("ŻółtaŁódka").Snake().String())
}

// TestSplit tests the Split function.
func TestSplit(t *testing.T) {
	assertEqualStringSlice(t, []string{"one", "two", "three", "four"}, Of("one-two-three-four").Split("-"))
	assertEqualStringSlice(t, []string{"", "", "D", "E", "", ""}, Of(",,D,E,,").Split(","))
	assertEqualStringSlice(t, []string{"one", "two", "three,four"}, Of("one,two,three,four").Split(",", 3))
}

// TestSquish tests the Squish function.
func TestSquish(t *testing.T) {
	assertEqual(t, "Hello World", Of("  Hello   World  ").Squish().String())
	assertEqual(t, "A B C", Of("A  B  C").Squish().String())
	assertEqual(t, "Lorem ipsum dolor sit amet", Of(" Lorem   ipsum \n  dolor  sit \t amet ").Squish().String())
	assertEqual(t, "Leading and trailing spaces", Of("  Leading  "+
		"and trailing "+
		" spaces  ").Squish().String())
	assertEqual(t, "", Of("").Squish().String())
}

// TestStart tests the Start function.
func TestStart(t *testing.T) {
	assertEqual(t, "/test/string", Of("test/string").Start("/").String())
	assertEqual(t, "/test/string", Of("/test/string").Start("/").String())
	assertEqual(t, "/test/string", Of("//test/string").Start("/").String())
}

// TestStartsWith tests the StartsWith function.
func TestStartsWith(t *testing.T) {
	assertTrue(t, Of("Wenbo Han").StartsWith("Wen"))
	assertTrue(t, Of("Wenbo Han").StartsWith("Wenbo"))
	assertTrue(t, Of("Wenbo Han").StartsWith("Han", "Wen"))
	assertFalse(t, Of("Wenbo Han").StartsWith())
	assertFalse(t, Of("Wenbo Han").StartsWith("we"))
	assertTrue(t, Of("Jönköping").StartsWith("Jö"))
	assertFalse(t, Of("Jönköping").StartsWith("Jonko"))
}

// TestStudly tests the Studly function.
func TestStudly(t *testing.T) {
	assertEqual(t, "GoravelGOFramework", Of("Goravel_g_o_framework").Studly().String())
	assertEqual(t, "GoravelGOFramework", Of("Goravel_gO_framework").Studly().String())
	assertEqual(t, "GoravelGoFramework", Of("Goravel -_- go -_-  framework  ").Studly().String())

	assertEqual(t, "FooBar", Of("FooBar").Studly().String())
	assertEqual(t, "FooBar", Of("foo_bar").Studly().String())
	assertEqual(t, "FooBar", Of("foo-Bar").Studly().String())
	assertEqual(t, "FooBar", Of("foo bar").Studly().String())
	assertEqual(t, "FooBar", Of("foo.bar").Studly().String())
}

// TestSubstrString tests the Substr function.
func TestSubstrString(t *testing.T) {
	assertEqual(t, "Ё", Of("БГДЖИЛЁ").Substr(-1).String())
	assertEqual(t, "ЛЁ", Of("БГДЖИЛЁ").Substr(-2).String())
	assertEqual(t, "И", Of("БГДЖИЛЁ").Substr(-3, 1).String())
	assertEqual(t, "ДЖИЛ", Of("БГДЖИЛЁ").Substr(2, -1).String())
	assertEqual(t, "", Of("БГДЖИЛЁ").Substr(4, -4).String())
	assertEqual(t, "ИЛ", Of("БГДЖИЛЁ").Substr(-3, -1).String())
	assertEqual(t, "ГДЖИЛЁ", Of("БГДЖИЛЁ").Substr(1).String())
	assertEqual(t, "ГДЖ", Of("БГДЖИЛЁ").Substr(1, 3).String())
	assertEqual(t, "БГДЖ", Of("БГДЖИЛЁ").Substr(0, 4).String())
	assertEqual(t, "Ё", Of("БГДЖИЛЁ").Substr(-1, 1).String())
	assertEqual(t, "", Of("Б").Substr(2).String())
}

// TestSwap tests the Swap function.
func TestSwap(t *testing.T) {
	assertEqual(t, "Go is excellent", Of("Golang is awesome").Swap(map[string]string{
		"Golang":  "Go",
		"awesome": "excellent",
	}).String())
	assertEqual(t, "Golang is awesome", Of("Golang is awesome").Swap(map[string]string{}).String())
	assertEqual(t, "Golang is awesome", Of("Golang is awesome").Swap(map[string]string{
		"":        "Go",
		"awesome": "excellent",
	}).String())
}

// TestTap tests the Tap function.
func TestTap(t *testing.T) {
	tap := Of("foobarbaz")
	fromTehTap := ""
	tap = tap.Tap(func(s String) {
		fromTehTap = s.Substr(0, 3).String()
	})
	assertEqual(t, "foo", fromTehTap)
	assertEqual(t, "foobarbaz", tap.String())
}

// TestTitle tests the Title function.
func TestTitle(t *testing.T) {
	assertEqual(t, "Krishan Kumar", Of("krishan kumar").Title().String())
	assertEqual(t, "Krishan Kumar", Of("kriSHan kuMAr").Title().String())
}

// TestTrim tests the Trim function.
func TestTrim(t *testing.T) {
	assertEqual(t, "foo", Of(" foo ").Trim().String())
	assertEqual(t, "foo", Of("_foo_").Trim("_").String())
}

// TestUcFirst tests the UcFirst function.
func TestUcFirst(t *testing.T) {
	assertEqual(t, "", Of("").UcFirst().String())
	assertEqual(t, "Framework", Of("framework").UcFirst().String())
	assertEqual(t, "Framework", Of("Framework").UcFirst().String())
	assertEqual(t, " framework", Of(" framework").UcFirst().String())
	assertEqual(t, "Goravel framework", Of("goravel framework").UcFirst().String())
}

// TestUcSplit tests the UcSplit function.
func TestUcSplit(t *testing.T) {
	assertEqualStringSlice(t, []string{"Krishan", "Kumar"}, Of("KrishanKumar").UcSplit())
	assertEqualStringSlice(t, []string{"Hello", "From", "Goravel"}, Of("HelloFromGoravel").UcSplit())
	assertEqualStringSlice(t, []string{"He_llo_", "World"}, Of("He_llo_World").UcSplit())
}

// TestUnless tests the Unless function.
func TestUnless(t *testing.T) {
	str := Of("Hello, World!")

	// Test case 1: The callback returns true, so the fallback should not be applied
	assertEqual(t, "Hello, World!", str.Unless(func(s *String) bool {
		return true
	}, func(s *String) *String {
		return Of("This should not be applied")
	}).String())

	// Test case 2: The callback returns false, so the fallback should be applied
	assertEqual(t, "Fallback Applied", str.Unless(func(s *String) bool {
		return false
	}, func(s *String) *String {
		return Of("Fallback Applied")
	}).String())

	// Test case 3: Testing with an empty string
	assertEqual(t, "Fallback Applied", Of("").Unless(func(s *String) bool {
		return false
	}, func(s *String) *String {
		return Of("Fallback Applied")
	}).String())
}

// TestUpper tests the Upper function.
func TestUpper(t *testing.T) {
	assertEqual(t, "FOO BAR BAZ", Of("foo bar baz").Upper().String())
	assertEqual(t, "FOO BAR BAZ", Of("foO bAr BaZ").Upper().String())
}

// TestWhen tests the When function.
func TestWhen(t *testing.T) {
	// true
	assertEqual(t, "when true", Of("when ").When(true, func(s *String) *String {
		return s.Append("true")
	}).String())
	assertEqual(t, "gets a value from if", Of("gets a value ").When(true, func(s *String) *String {
		return s.Append("from if")
	}).String())

	// false
	assertEqual(t, "when", Of("when").When(false, func(s *String) *String {
		return s.Append("true")
	}).String())

	assertEqual(t, "when false fallbacks to default", Of("when false ").When(false, func(s *String) *String {
		return s.Append("true")
	}, func(s *String) *String {
		return s.Append("fallbacks to default")
	}).String())
}

// TestWhenContains tests the WhenContains function.
func TestWhenContains(t *testing.T) {
	assertEqual(t, "Tony Stark", Of("stark").WhenContains("tar", func(s *String) *String {
		return s.Prepend("Tony ").Title()
	}, func(s *String) *String {
		return s.Prepend("Arno ").Title()
	}).String())

	assertEqual(t, "stark", Of("stark").WhenContains("xxx", func(s *String) *String {
		return s.Prepend("Tony ").Title()
	}).String())

	assertEqual(t, "Arno Stark", Of("stark").WhenContains("xxx", func(s *String) *String {
		return s.Prepend("Tony ").Title()
	}, func(s *String) *String {
		return s.Prepend("Arno ").Title()
	}).String())
}

// TestWhenContainsAll tests the WhenContainsAll function.
func TestWhenContainsAll(t *testing.T) {
	// Test when all values are present
	assertEqual(t, "Tony Stark", Of("tony stark").WhenContainsAll([]string{"tony", "stark"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())

	// Test when not all values are present
	assertEqual(t, "tony stark", Of("tony stark").WhenContainsAll([]string{"xxx"},
		func(s *String) *String {
			return s.Title()
		},
	).String())

	// Test when some values are present and some are not
	assertEqual(t, "TonyStark", Of("tony stark").WhenContainsAll([]string{"tony", "xxx"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())
}

// TestWhenEmpty tests the WhenEmpty function.
func TestWhenEmpty(t *testing.T) {
	// Test when the string is empty
	assertEqual(t, "DEFAULT", Of("").WhenEmpty(
		func(s *String) *String {
			return s.Append("default").Upper()
		}).String())

	// Test when the string is not empty
	assertEqual(t, "non-empty", Of("non-empty").WhenEmpty(
		func(s *String) *String {
			return s.Append("default")
		},
	).String())
}

// TestWhenIsAscii tests the WhenIsAscii function.
func TestWhenIsAscii(t *testing.T) {
	assertEqual(t, "Ascii: A", Of("A").WhenIsAscii(
		func(s *String) *String {
			return s.Prepend("Ascii: ")
		}).String())
	assertEqual(t, "ù", Of("ù").WhenIsAscii(
		func(s *String) *String {
			return s.Prepend("Ascii: ")
		}).String())
	assertEqual(t, "Not Ascii: ù", Of("ù").WhenIsAscii(
		func(s *String) *String {
			return s.Prepend("Ascii: ")
		},
		func(s *String) *String {
			return s.Prepend("Not Ascii: ")
		},
	).String())
}

// TestWhenNotEmpty tests the WhenNotEmpty function.
func TestWhenNotEmpty(t *testing.T) {
	// Test when the string is not empty
	assertEqual(t, "UPPERCASE", Of("uppercase").WhenNotEmpty(
		func(s *String) *String {
			return s.Upper()
		},
	).String())

	// Test when the string is empty
	assertEqual(t, "", Of("").WhenNotEmpty(
		func(s *String) *String {
			return s.Append("not empty")
		},
		func(s *String) *String {
			return s.Upper()
		},
	).String())
}

// TestWhenStartsWith tests the WhenStartsWith function.
func TestWhenStartsWith(t *testing.T) {
	// Test when the string starts with a specific prefix
	assertEqual(t, "Tony Stark", Of("tony stark").WhenStartsWith([]string{"ton"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())

	// Test when the string starts with any of the specified prefixes
	assertEqual(t, "Tony Stark", Of("tony stark").WhenStartsWith([]string{"ton", "not"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())

	// Test when the string does not start with the specified prefix
	assertEqual(t, "tony stark", Of("tony stark").WhenStartsWith([]string{"xxx"},
		func(s *String) *String {
			return s.Title()
		},
	).String())

	// Test when the string starts with one of the specified prefixes and not the other
	assertEqual(t, "Tony Stark", Of("tony stark").WhenStartsWith([]string{"tony", "xxx"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())
}

// TestWhenEndsWith tests the WhenEndsWith function.
func TestWhenEndsWith(t *testing.T) {
	// Test when the string ends with a specific suffix
	assertEqual(t, "Tony Stark", Of("tony stark").WhenEndsWith([]string{"ark"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())

	// Test when the string ends with any of the specified suffixes
	assertEqual(t, "Tony Stark", Of("tony stark").WhenEndsWith([]string{"kra", "ark"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())

	// Test when the string does not end with the specified suffix
	assertEqual(t, "tony stark", Of("tony stark").WhenEndsWith([]string{"xxx"},
		func(s *String) *String {
			return s.Title()
		},
	).String())

	// Test when the string ends with one of the specified suffixes and not the other
	assertEqual(t, "TonyStark", Of("tony stark").WhenEndsWith([]string{"tony", "xxx"},
		func(s *String) *String {
			return s.Title()
		},
		func(s *String) *String {
			return s.Studly()
		},
	).String())
}

// TestWhenExactly tests the WhenExactly function.
func TestWhenExactly(t *testing.T) {
	// Test when the string exactly matches the expected value
	assertEqual(t, "Nailed it...!", Of("Tony Stark").WhenExactly("Tony Stark",
		func(s *String) *String {
			return Of("Nailed it...!")
		},
		func(s *String) *String {
			return Of("Swing and a miss...!")
		},
	).String())

	// Test when the string does not exactly match the expected value
	assertEqual(t, "Swing and a miss...!", Of("Tony Stark").WhenExactly("Iron Man",
		func(s *String) *String {
			return Of("Nailed it...!")
		},
		func(s *String) *String {
			return Of("Swing and a miss...!")
		},
	).String())

	// Test when the string exactly matches the expected value with no "else" callback
	assertEqual(t, "Tony Stark", Of("Tony Stark").WhenExactly("Iron Man",
		func(s *String) *String {
			return Of("Nailed it...!")
		},
	).String())
}

// TestWhenNotExactly tests the WhenNotExactly function.
func TestWhenNotExactly(t *testing.T) {
	// Test when the string does not exactly match the expected value with an "else" callback
	assertEqual(t, "Iron Man", Of("Tony").WhenNotExactly("Tony Stark",
		func(s *String) *String {
			return Of("Iron Man")
		},
	).String())

	// Test when the string does not exactly match the expected value with both "if" and "else" callbacks
	assertEqual(t, "Swing and a miss...!", Of("Tony Stark").WhenNotExactly("Tony Stark",
		func(s *String) *String {
			return Of("Iron Man")
		},
		func(s *String) *String {
			return Of("Swing and a miss...!")
		},
	).String())
}

// TestWhenIs tests the WhenIs function.
func TestWhenIs(t *testing.T) {
	// Test when the string exactly matches the expected value with an "if" callback
	assertEqual(t, "Winner: /", Of("/").WhenIs("/",
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
		func(s *String) *String {
			return Of("Try again")
		},
	).String())

	// Test when the string does not exactly match the expected value with an "if" callback
	assertEqual(t, "/", Of("/").WhenIs(" /",
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
	).String())

	// Test when the string does not exactly match the expected value with both "if" and "else" callbacks
	assertEqual(t, "Try again", Of("/").WhenIs(" /",
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
		func(s *String) *String {
			return Of("Try again")
		},
	).String())

	// Test when the string matches a pattern using wildcard and "if" callback
	assertEqual(t, "Winner: foo/bar/baz", Of("foo/bar/baz").WhenIs("foo/*",
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
	).String())
}

// TestWhenIsUlid tests the WhenIsUlid function.
func TestWhenIsUlid(t *testing.T) {
	// Test when the string is a valid ULID with an "if" callback
	assertEqual(t, "Ulid: 01GJSNW9MAF792C0XYY8RX6QFT", Of("01GJSNW9MAF792C0XYY8RX6QFT").WhenIsUlid(
		func(s *String) *String {
			return s.Prepend("Ulid: ")
		},
		func(s *String) *String {
			return s.Prepend("Not Ulid: ")
		},
	).String())

	// Test when the string is not a valid ULID with an "if" callback
	assertEqual(t, "2cdc7039-65a6-4ac7-8e5d-d554a98", Of("2cdc7039-65a6-4ac7-8e5d-d554a98").WhenIsUlid(
		func(s *String) *String {
			return s.Prepend("Ulid: ")
		},
	).String())

	// Test when the string is not a valid ULID with both "if" and "else" callbacks
	assertEqual(t, "Not Ulid: ss-01GJSNW9MAF792C0XYY8RX6QFT", Of("ss-01GJSNW9MAF792C0XYY8RX6QFT").WhenIsUlid(
		func(s *String) *String {
			return s.Prepend("Ulid: ")
		},
		func(s *String) *String {
			return s.Prepend("Not Ulid: ")
		},
	).String())
}

// TestWhenIsUuid tests the WhenIsUuid function.
func TestWhenIsUuid(t *testing.T) {
	// Test when the string is a valid UUID with an "if" callback
	assertEqual(t, "Uuid: 2cdc7039-65a6-4ac7-8e5d-d554a98e7b15", Of("2cdc7039-65a6-4ac7-8e5d-d554a98e7b15").WhenIsUuid(
		func(s *String) *String {
			return s.Prepend("Uuid: ")
		},
		func(s *String) *String {
			return s.Prepend("Not Uuid: ")
		},
	).String())

	assertEqual(t, "2cdc7039-65a6-4ac7-8e5d-d554a98", Of("2cdc7039-65a6-4ac7-8e5d-d554a98").WhenIsUuid(
		func(s *String) *String {
			return s.Prepend("Uuid: ")
		},
	).String())

	assertEqual(t, "Not Uuid: 2cdc7039-65a6-4ac7-8e5d-d554a98", Of("2cdc7039-65a6-4ac7-8e5d-d554a98").WhenIsUuid(
		func(s *String) *String {
			return s.Prepend("Uuid: ")
		},
		func(s *String) *String {
			return s.Prepend("Not Uuid: ")
		},
	).String())
}

// TestWhenTest tests the WhenTest function.
func TestWhenTest(t *testing.T) {
	// Test when the regular expression matches with an "if" callback
	assertEqual(t, "Winner: foo bar", Of("foo bar").WhenTest(`bar*`,
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
		func(s *String) *String {
			return Of("Try again")
		},
	).String())

	// Test when the regular expression does not match with an "if" callback
	assertEqual(t, "Try again", Of("foo bar").WhenTest(`/link/`,
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
		func(s *String) *String {
			return Of("Try again")
		},
	).String())

	// Test when the regular expression does not match with both "if" and "else" callbacks
	assertEqual(t, "foo bar", Of("foo bar").WhenTest(`/link/`,
		func(s *String) *String {
			return s.Prepend("Winner: ")
		},
	).String())
}

// TestWordCount tests the WordCount function.
func TestWordCount(t *testing.T) {
	assertEqual(t, 2, Of("Hello, world!").WordCount())
	assertEqual(t, 10, Of("Hi, this is my first contribution to the Goravel framework.").WordCount())
}

// TestWords tests the Words function.
func TestWords(t *testing.T) {
	assertEqual(t, "Perfectly balanced, as >>>", Of("Perfectly balanced, as all things should be.").Words(3, " >>>").String())
	assertEqual(t, "Perfectly balanced, as all things should be.", Of("Perfectly balanced, as all things should be.").Words(100).String())
}

// TestFieldsFunc tests the fieldsFunc helper function.
func TestFieldsFunc(t *testing.T) {
	tests := []struct {
		input          string
		shouldPreserve []func(rune) bool
		expected       []string
	}{
		// Test case 1: Basic word splitting with space separator.
		{
			input:    "Hello World",
			expected: []string{"Hello", "World"},
		},
		// Test case 2: Splitting with space and preserving hyphen.
		{
			input:          "Hello-World",
			shouldPreserve: []func(rune) bool{func(r rune) bool { return r == '-' }},
			expected:       []string{"Hello", "-World"},
		},
		// Test case 3: Splitting with space and preserving multiple characters.
		{
			input: "Hello-World,This,Is,a,Test",
			shouldPreserve: []func(rune) bool{
				func(r rune) bool { return r == '-' },
				func(r rune) bool { return r == ',' },
			},
			expected: []string{"Hello", "-World", ",This", ",Is", ",a", ",Test"},
		},
		// Test case 4: No splitting when no separator is found.
		{
			input:    "HelloWorld",
			expected: []string{"HelloWorld"},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := fieldsFunc(test.input, func(r rune) bool { return r == ' ' }, test.shouldPreserve...)
			assertEqualStringSlice(t, test.expected, result)
		})
	}
}

// TestSubstr tests the Substr helper function.
func TestSubstr(t *testing.T) {
	assertEqual(t, "world", Substr("Hello, world!", 7, 5))
	assertEqual(t, "", Substr("Golang", 10))
	assertEqual(t, "tine", Substr("Goroutines", -5, 4))
	assertEqual(t, "ic", Substr("Unicode", 2, -3))
	assertEqual(t, "esting", Substr("Testing", 1, 10))
	assertEqual(t, "", Substr("", 0, 5))
	assertEqual(t, "世界！", Substr("你好，世界！", 3, 3))
}

// TestMaximum tests the maximum helper function.
func TestMaximum(t *testing.T) {
	assertEqual(t, 10, maximum(5, 10))
	assertEqual(t, 3.14, maximum(3.14, 2.71))
	assertEqual(t, "banana", maximum("apple", "banana"))
	assertEqual(t, -5, maximum(-5, -10))
	assertEqual(t, 42, maximum(42, 42))
}

// TestRandom tests the Random helper function.
func TestRandom(t *testing.T) {
	assertLen(t, Random(10), 10)
	assertEmpty(t, Random(0))
	assertPanics(t, func() {
		Random(-1)
	})
}

// TestCase2Camel tests the Case2Camel helper function.
func TestCase2Camel(t *testing.T) {
	assertEqual(t, "GoravelFramework", Case2Camel("goravel_framework"))
	assertEqual(t, "GoravelFramework1", Case2Camel("goravel_framework1"))
	assertEqual(t, "GoravelFramework", Case2Camel("GoravelFramework"))
}

// TestCamel2Case tests the Camel2Case helper function.
func TestCamel2Case(t *testing.T) {
	assertEqual(t, "goravel_framework", Camel2Case("GoravelFramework"))
	assertEqual(t, "goravel_framework1", Camel2Case("GoravelFramework1"))
	assertEqual(t, "goravel_framework", Camel2Case("goravel_framework"))
}
